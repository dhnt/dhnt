// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package dev is a Layer 3 executor leaf for software-development skills:
// it binds the dhnt primitive `run` and the predicates `exit-zero` and
// `empty-output` to real process execution and filesystem checks. It is
// the process/filesystem analogue of skills/tui.
//
// As with tui, the dhnt skill names abstract command/predicate ids; the
// concrete argv lives in a Spec's command table, so per-repo specifics
// (e.g. `go test` vs `make test`) are runtime config and the skill stays
// portable. This is what lets one go-verify skill learn each repo's test
// invocation as an environment-guarded branch.
package dev

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/dhnt/dhnt/skills"
)

// dhnt ids this leaf binds.
const (
	PrimRun      = "rune"    // run a command (a step that must exit 0)
	PredExit     = "exito"   // the named command exits 0
	PredEmptyOut = "enupoto" // the named command produces empty stdout
)

// Command is one bound command: its argv and the effects it declares.
type Command struct {
	Argv    []string
	Effects []skills.Effect
}

// Spec maps abstract command ids (referenced by skill steps/predicates)
// to concrete commands, run in Dir.
type Spec struct {
	Dir      string
	Commands map[string]Command
	Timeout  time.Duration
}

// Session runs a Spec and memoizes each command's result within a run so
// a command referenced by both a step and a predicate runs once.
type Session struct {
	spec Spec
	mu   sync.Mutex
	ran  map[string]result
}

type result struct {
	exit int
	out  string // stdout
	err  error  // exec error (e.g. missing binary), not a non-zero exit
}

// NewEnv binds the dev primitives/predicates for spec.
func NewEnv(spec Spec) (skills.Env, *Session, error) {
	if len(spec.Commands) == 0 {
		return skills.Env{}, nil, fmt.Errorf("dev: empty command table")
	}
	if spec.Timeout == 0 {
		spec.Timeout = 10 * time.Minute
	}
	s := &Session{spec: spec, ran: map[string]result{}}
	env := skills.Env{
		Primitives: map[string]skills.PrimitiveFn{PrimRun: s.run},
		Predicates: map[string]skills.PredicateFn{
			PredExit:     s.exitZero,
			PredEmptyOut: s.emptyOutput,
		},
	}
	return env, s, nil
}

func argRef(args []skills.Arg) (string, error) {
	for _, a := range args {
		if a.Name == "value" && a.Value.Kind == skills.ExprRef {
			return a.Value.Ref, nil
		}
	}
	return "", fmt.Errorf("dev: step missing a value ref")
}

// exec runs the command for ref (memoized) and returns its result.
func (s *Session) exec(ref string) result {
	s.mu.Lock()
	if r, ok := s.ran[ref]; ok {
		s.mu.Unlock()
		return r
	}
	s.mu.Unlock()

	c, ok := s.spec.Commands[ref]
	if !ok || len(c.Argv) == 0 {
		return result{err: fmt.Errorf("dev: no command bound for %q", ref)}
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.spec.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.Argv[0], c.Argv[1:]...)
	cmd.Dir = s.spec.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	r := result{out: stdout.String()}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			r.exit = ee.ExitCode()
		} else {
			r.err = fmt.Errorf("%v: %s", err, tail(stderr.String(), 200))
		}
	}
	s.mu.Lock()
	s.ran[ref] = r
	s.mu.Unlock()
	return r
}

// run executes a command step; a non-zero exit (or exec error) fails it.
func (s *Session) run(args []skills.Arg) ([]skills.Effect, error) {
	ref, err := argRef(args)
	if err != nil {
		return nil, err
	}
	c := s.spec.Commands[ref]
	r := s.exec(ref)
	if r.err != nil {
		return nil, fmt.Errorf("dev: run %q: %w", ref, r.err)
	}
	if r.exit != 0 {
		return nil, fmt.Errorf("dev: run %q exited %d: %s", ref, r.exit, tail(r.out, 200))
	}
	return c.Effects, nil
}

// exitZero reports whether the named command exits 0.
func (s *Session) exitZero(args []skills.Arg) (bool, []skills.Effect, error) {
	ref, err := argRef(args)
	if err != nil {
		return false, nil, err
	}
	r := s.exec(ref)
	return r.err == nil && r.exit == 0, []skills.Effect{skills.EffRead, skills.EffTime}, nil
}

// emptyOutput reports whether the named command produced no stdout (the
// "nothing to fix" idiom, e.g. `gofmt -l .` or `git status --porcelain`).
func (s *Session) emptyOutput(args []skills.Arg) (bool, []skills.Effect, error) {
	ref, err := argRef(args)
	if err != nil {
		return false, nil, err
	}
	r := s.exec(ref)
	if r.err != nil {
		return false, []skills.Effect{skills.EffRead}, nil
	}
	return strings.TrimSpace(r.out) == "", []skills.Effect{skills.EffRead}, nil
}

// Output returns the captured stdout of a command ref (for reporting).
func (s *Session) Output(ref string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ran[ref].out
}

func tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}
