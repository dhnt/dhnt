// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command agent_drive drives any catalogued agentic CLI (claude, codex,
// gemini, opencode, aider) through the same dhnt skill, selected by the
// Spec catalog in skills/tui.
//
//	go run . <agent>              # token-free: drive `<agent> --version`
//	go run . <agent> --headless   # spends tokens: one-shot "reply PONG"
//
// The point: ONE dhnt skill (tui.DriveOnceSkill) drives every tool; only
// the Spec changes per tool.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: agent_drive <claude|codex|gemini|opencode|aider> [--headless]")
		os.Exit(2)
	}
	name := os.Args[1]
	headless := len(os.Args) > 2 && os.Args[2] == "--headless"

	a, ok := tui.Agents[name]
	if !ok {
		fmt.Printf("unknown agent %q (have: claude codex gemini opencode aider)\n", name)
		os.Exit(2)
	}
	if _, err := exec.LookPath(a.Version[0]); err != nil {
		fmt.Printf("%s not on PATH; skipping\n", name)
		return
	}

	var spec tui.Spec
	if headless {
		// Tiny, ~free prompt; the answer must contain PONG.
		spec, _ = tui.HeadlessSpec(name, "Reply with exactly one word: PONG", `(?i)PONG`, "")
		fmt.Printf("driving %q headless (spends model tokens)\n", name)
	} else {
		spec, _ = tui.VersionSpec(name, "")
		fmt.Printf("driving %q --version (token-free)\n", name)
	}

	env, sess, err := tui.NewEnv(spec)
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	skill := tui.DriveOnceSkill()
	att, runErr := skills.Run(skill, env, name)
	fmt.Printf("--- %s output (tail) ---\n%s\n--- end ---\n", name, tail(sess.Output(), 600))
	if runErr != nil {
		fmt.Printf("run error: %v\n", runErr)
		os.Exit(1)
	}
	fmt.Printf("attestation: valid=%v passed=%v effects=%v consistent=%v\n",
		att.Valid, att.Passed, att.Effects, att.Consistent(skill))
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}
