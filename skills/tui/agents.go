// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package tui

import (
	"time"

	"github.com/dhnt/dhnt/skills"
)

// This file is a catalog of Specs for driving real agentic coding CLIs
// through a dhnt skill.
//
// Two surfaces per tool:
//
//   - Headless: a one-shot, non-interactive invocation (claude -p,
//     codex exec, gemini -p, opencode run, aider --message). It prints
//     the model's answer and exits — line-oriented and reliably
//     drivable by expect. This is the surface you automate. It spends
//     model tokens.
//   - Version: `<tool> --version`, which prints a version and exits.
//     Token-free; used to smoke-test that the driver wires up against
//     the real binary.
//
// The full-screen *interactive* TUIs (codex/gemini/opencode/claude)
// query the terminal for capabilities (DA1, kitty-keyboard, OSC colors)
// and block until answered, so a naive PTY cannot drive them; their
// headless mode is the supported automation path. (aider's interactive
// prompt is line-oriented, so it is also drivable interactively — see
// examples/aider_drive.)

// Per-tool runtime prerequisites (the dhnt skill drives the CLI; the
// CLI's own auth/trust is the operator's responsibility):
//
//   - claude    — authenticated `claude` (subscription or API key).
//   - codex     — `codex login` (or OPENAI_API_KEY) beforehand.
//   - gemini    — auth + a *trusted* workspace: set
//     GEMINI_CLI_TRUST_WORKSPACE=true (or run in a trusted dir), else
//     headless mode refuses with a "not a trusted directory" message.
//   - opencode  — a configured provider (`opencode auth`).
//   - aider     — a model + key (e.g. DEEPSEEK_API_KEY) and ~/.aider.conf.yml.
//
// When a prerequisite is missing the tool prints a refusal and exits or
// stalls — and the dhnt contract correctly reports the run invalid,
// which is the system working, not a driver bug.

// AgentCLI describes how to invoke one coding CLI.
type AgentCLI struct {
	Name string
	// Headless builds argv for a one-shot prompt that prints and exits.
	Headless func(prompt string) []string
	// Version builds argv that prints a version and exits (token-free).
	Version []string
}

// Agents is the catalog, keyed by tool name.
var Agents = map[string]AgentCLI{
	"claude": {
		Name:     "claude",
		Headless: func(p string) []string { return []string{"claude", "-p", p} },
		Version:  []string{"claude", "--version"},
	},
	"codex": {
		Name:     "codex",
		Headless: func(p string) []string { return []string{"codex", "exec", p} },
		Version:  []string{"codex", "--version"},
	},
	"gemini": {
		Name:     "gemini",
		Headless: func(p string) []string { return []string{"gemini", "-p", p} },
		Version:  []string{"gemini", "--version"},
	},
	"opencode": {
		Name:     "opencode",
		Headless: func(p string) []string { return []string{"opencode", "run", p} },
		Version:  []string{"opencode", "--version"},
	},
	"aider": {
		Name: "aider",
		Headless: func(p string) []string {
			return []string{"aider", "--message", p, "--no-pretty", "--no-auto-commits", "--map-tokens", "0", "--yes-always"}
		},
		Version: []string{"aider", "--version"},
	},
}

// VersionSpec returns a token-free Spec that runs `<tool> --version` and
// confirms a semver in the output. The process exits on its own.
func VersionSpec(name, dir string) (Spec, bool) {
	a, ok := Agents[name]
	if !ok {
		return Spec{}, false
	}
	return Spec{
		Argv:     a.Version,
		Dir:      dir,
		Patterns: map[string]string{"readayu": `\d+\.\d+\.\d+`},
		Quit:     "", // the process self-exits; quit just reaps it
		Timeout:  20 * time.Second,
	}, true
}

// HeadlessSpec returns a Spec that runs a one-shot prompt and waits for
// answerPat in the output. It spends model tokens.
func HeadlessSpec(name, prompt, answerPat, dir string) (Spec, bool) {
	a, ok := Agents[name]
	if !ok {
		return Spec{}, false
	}
	return Spec{
		Argv:     a.Headless(prompt),
		Dir:      dir,
		Patterns: map[string]string{"readayu": answerPat},
		Quit:     "",
		Timeout:  120 * time.Second,
	}, true
}

// Completer returns a skills.Completer backed by an agent CLI in headless
// mode: it runs `<agent> <prompt>` to completion over a PTY and returns
// everything the tool printed. This lets the L0→L2 normaliser use a real
// coding CLI as its model — the language bootstrapping itself.
//
// dir is the working directory for the tool ("" inherits). Auth/trust
// prerequisites for the chosen agent still apply (see the catalog notes).
func Completer(agent, dir string, timeout time.Duration) (skills.Completer, bool) {
	a, ok := Agents[agent]
	if !ok {
		return nil, false
	}
	if timeout == 0 {
		timeout = 180 * time.Second
	}
	return func(prompt string) (string, error) {
		spec := Spec{Argv: a.Headless(prompt), Dir: dir, Quit: "", Timeout: timeout}
		env, sess, err := NewEnv(spec)
		if err != nil {
			return "", err
		}
		defer sess.Close()
		// spawn → reap; capture everything the tool emitted.
		skill := skills.Skill{
			Name:      "capituru",
			EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffNet, skills.EffTime},
			Steps: []skills.Step{
				{Name: "sa", Primitive: PrimSpawn},
				{Name: "su", Primitive: PrimQuit},
			},
		}
		if _, err := skills.Run(skill, env, agent); err != nil {
			return sess.Output(), err
		}
		return sess.Output(), nil
	}, true
}

// DriveOnceSkill is the dhnt skill behind both Specs above: spawn the
// tool, expect the looked-for output (a version string or the model's
// answer), then reap it — with the contract that it exited cleanly. One
// skill drives every tool; the Spec selects which and what to look for.
func DriveOnceSkill() skills.Skill {
	ref := func(k string) []skills.Arg {
		return []skills.Arg{{Name: "value", Value: skills.NewRef(k)}}
	}
	return skills.Skill{
		Name:      "darive",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffNet, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: PrimSpawn},
			{Name: "se", Primitive: PrimExpect, Args: ref("readayu")},
			{Name: "su", Primitive: PrimQuit},
		},
		Contract: []skills.Check{{Predicate: PredClean}},
	}
}
