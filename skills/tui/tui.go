// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package tui is a Layer 3 executor leaf for driving PTY/terminal
// applications (TUIs) from a dhnt skill — the expect(1) idiom, made
// machine-checkable. It binds the abstract dhnt primitives
// (spawn/send/expect/quit) and the `clean-exit` predicate to a real
// pseudo-terminal via github.com/creack/pty (MIT).
//
// The point of the split between the dhnt skill and this Spec: the skill
// encodes the *protocol* (spawn → expect READY → send CMD → expect
// RESULT → quit), language-neutral and tool-agnostic; the Spec supplies
// the concrete argv, regex patterns, inputs, and timeout for one tool.
// The same skill drives cat, bash, or an agentic CLI (claude / codex /
// gemini / aider / opencode) by swapping the Spec.
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/dhnt/dhnt/skills"
)

// dhnt ids this leaf binds (see EncodeWord of the English words).
const (
	PrimSpawn  = "sopawuni"  // spawn
	PrimSend   = "senida"    // send
	PrimExpect = "exupecato" // expect
	PrimQuit   = "quito"     // quit
	PredClean  = "caleani"   // clean (exited with status 0)
)

// Spec configures the driver for one terminal tool.
type Spec struct {
	Argv     []string          // program + args to spawn (required)
	Patterns map[string]string // expect ref -> regexp ("" matches immediately)
	Inputs   map[string]string // send ref -> literal text to write
	Quit     string            // text the quit primitive writes ("" => send EOF)
	Timeout  time.Duration     // per-expect timeout (default 5s)
}

// Session is the live PTY-backed process. Construct it via NewEnv and
// always Close it (the caller should defer Close to avoid leaking the
// child on a failed expect).
type Session struct {
	spec   Spec
	res    map[string]*regexp.Regexp
	cmd    *exec.Cmd
	ptmx   *os.File
	mu     sync.Mutex
	buf    []byte
	exited bool
	exitOK bool
}

// NewEnv compiles the spec and returns a skills.Env wiring the dhnt
// primitives/predicate to this session.
func NewEnv(spec Spec) (skills.Env, *Session, error) {
	if len(spec.Argv) == 0 {
		return skills.Env{}, nil, fmt.Errorf("tui: empty argv")
	}
	if spec.Timeout == 0 {
		spec.Timeout = 5 * time.Second
	}
	s := &Session{spec: spec, res: make(map[string]*regexp.Regexp)}
	for ref, pat := range spec.Patterns {
		re, err := regexp.Compile(pat)
		if err != nil {
			return skills.Env{}, nil, fmt.Errorf("tui: pattern %q: %w", ref, err)
		}
		s.res[ref] = re
	}
	env := skills.Env{
		Primitives: map[string]skills.PrimitiveFn{
			PrimSpawn: s.spawn,
			PrimSend:  s.send,
			PrimQuit:  s.quit,
		},
		Predicates: map[string]skills.PredicateFn{
			PredClean: s.clean,
		},
	}
	// expect causes read+time and may fail; it is modelled as a primitive
	// (a checkpoint) — a timeout aborts the run.
	env.Primitives[PrimExpect] = s.expect
	return env, s, nil
}

func argRef(args []skills.Arg) (string, error) {
	for _, a := range args {
		if a.Name == "value" && a.Value.Kind == skills.ExprRef {
			return a.Value.Ref, nil
		}
	}
	return "", fmt.Errorf("tui: step missing a value ref")
}

func (s *Session) spawn(args []skills.Arg) ([]skills.Effect, error) {
	cmd := exec.Command(s.spec.Argv[0], s.spec.Argv[1:]...)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("tui: spawn %v: %w", s.spec.Argv, err)
	}
	s.cmd = cmd
	s.ptmx = ptmx
	go s.drain()
	return []skills.Effect{skills.EffWrite, skills.EffTime}, nil
}

// drain copies PTY output into the buffer until the child closes it.
func (s *Session) drain() {
	b := make([]byte, 4096)
	for {
		n, err := s.ptmx.Read(b)
		if n > 0 {
			s.mu.Lock()
			s.buf = append(s.buf, b[:n]...)
			s.mu.Unlock()
		}
		if err != nil {
			return
		}
	}
}

func (s *Session) snapshot() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, len(s.buf))
	copy(out, s.buf)
	return out
}

func (s *Session) expect(args []skills.Arg) ([]skills.Effect, error) {
	ref, err := argRef(args)
	if err != nil {
		return nil, err
	}
	re, ok := s.res[ref]
	if !ok {
		return nil, fmt.Errorf("tui: no pattern bound for %q", ref)
	}
	deadline := time.Now().Add(s.spec.Timeout)
	for {
		if re.Match(s.snapshot()) {
			return []skills.Effect{skills.EffRead, skills.EffTime}, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("tui: expect %q (/%s/) timed out after %s; last output:\n%s",
				ref, re.String(), s.spec.Timeout, tail(s.snapshot(), 240))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Session) send(args []skills.Arg) ([]skills.Effect, error) {
	ref, err := argRef(args)
	if err != nil {
		return nil, err
	}
	text, ok := s.spec.Inputs[ref]
	if !ok {
		return nil, fmt.Errorf("tui: no input bound for %q", ref)
	}
	if _, err := s.ptmx.WriteString(text); err != nil {
		return nil, fmt.Errorf("tui: send %q: %w", ref, err)
	}
	return []skills.Effect{skills.EffWrite}, nil
}

func (s *Session) quit(args []skills.Arg) ([]skills.Effect, error) {
	if s.ptmx != nil {
		if s.spec.Quit != "" {
			_, _ = s.ptmx.WriteString(s.spec.Quit)
		} else {
			_, _ = s.ptmx.Write([]byte{0x04}) // Ctrl-D (EOF)
		}
	}
	s.wait()
	return []skills.Effect{skills.EffWrite, skills.EffTime}, nil
}

// wait reaps the child with a timeout and records the outcome.
func (s *Session) wait() {
	if s.cmd == nil || s.exited {
		return
	}
	done := make(chan error, 1)
	go func() { done <- s.cmd.Wait() }()
	select {
	case err := <-done:
		s.exited = true
		s.exitOK = err == nil
	case <-time.After(s.spec.Timeout):
		_ = s.cmd.Process.Kill()
		<-done
		s.exited = true
		s.exitOK = false
	}
}

func (s *Session) clean(args []skills.Arg) (bool, []skills.Effect, error) {
	if !s.exited {
		s.wait()
	}
	return s.exited && s.exitOK, []skills.Effect{skills.EffRead}, nil
}

// Close kills the child if it is still running and releases the PTY.
func (s *Session) Close() {
	if s.cmd != nil && s.cmd.Process != nil && !s.exited {
		_ = s.cmd.Process.Kill()
	}
	if s.ptmx != nil {
		_ = s.ptmx.Close()
	}
}

// Output returns everything read from the PTY so far (for status
// reporting / interrogation).
func (s *Session) Output() string { return string(s.snapshot()) }

func tail(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return "..." + string(b[len(b)-n:])
}
