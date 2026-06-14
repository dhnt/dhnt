// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command aider_drive drives the real `aider` agentic coding CLI
// end-to-end through the dhnt branching skill, over a pseudo-terminal.
//
// It exercises aider's interactive TUI without spending model tokens:
// repo-map is disabled and only slash commands (/help, /exit) are sent,
// neither of which calls the LLM. The branch handles aider's optional
// ".gitignore" confirmation prompt — react if shown, skip if not.
//
// Requires `aider` and `git` on PATH (skips cleanly otherwise).
//
//	go run .
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

// driveAider: spawn → expect the first prompt → if a gitignore prompt is
// shown, decline it → ask for /help → confirm the help listing → /exit.
func driveAider() skills.Skill {
	ref := func(k string) []skills.Arg {
		return []skills.Arg{{Name: "value", Value: skills.NewRef(k)}}
	}
	return skills.Skill{
		Name:      "darive",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffNet, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: tui.PrimSpawn},
			{Name: "se", Primitive: tui.PrimExpect, Args: ref("readayu")}, // first interaction point
			{Branch: &skills.Branch{ // handle the optional gitignore prompt
				Cond: skills.Check{Predicate: tui.PredSeen, Args: ref("marokero")},
				Then: []skills.Step{{Name: "ta", Primitive: tui.PrimSend, Args: ref("nope")}},
			}},
			{Name: "si", Primitive: tui.PrimSend, Args: ref("helepe")},     // send /help
			{Name: "so", Primitive: tui.PrimExpect, Args: ref("resulito")}, // help listing shown
			{Name: "su", Primitive: tui.PrimQuit},                          // /exit
		},
		Contract: []skills.Check{{Predicate: tui.PredClean}},
	}
}

func main() {
	if _, err := exec.LookPath("aider"); err != nil {
		fmt.Println("aider not found on PATH; skipping")
		return
	}
	dir, err := os.MkdirTemp("", "dhnt-aider-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	gitInit(dir)

	spec := tui.Spec{
		Argv: []string{
			"aider",
			"--no-pretty", "--no-stream", "--no-check-update",
			"--no-auto-commits", "--map-tokens", "0",
		},
		Dir: dir,
		Patterns: map[string]string{
			"readayu":  `(?i)(gitignore|Aider v|>)`,      // first prompt or banner
			"marokero": `(?i)gitignore`,                  // the confirmation we branch on
			"resulito": `(?i)(/help|/tokens|/exit|/add)`, // /help command listing
		},
		Inputs: map[string]string{
			"nope":   "n\n",     // decline adding to .gitignore
			"helepe": "/help\n", // ask for the command list
		},
		Quit:    "/exit\n",
		Timeout: 45 * time.Second,
	}

	// aider must run inside the repo (spec.Dir set above).
	env, sess, err := tui.NewEnv(spec)
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	skill := driveAider()
	id, _ := skills.Identity(skill)
	fmt.Printf("dhnt skill %q identity=%s…\n", skill.Name, id[:18])
	fmt.Printf("driving aider 0.86 over a real PTY in %s\n\n", dir)

	att, runErr := skills.Run(skill, env, "aider")
	fmt.Printf("--- aider PTY output (tail) ---\n%s\n--- end ---\n\n", tail(sess.Output(), 1200))
	if runErr != nil {
		fmt.Printf("run error: %v\n", runErr)
		os.Exit(1)
	}
	fmt.Printf("attestation: valid=%v passed=%v effects=%v consistent=%v\n",
		att.Valid, att.Passed, att.Effects, att.Consistent(skill))
}

func gitInit(dir string) {
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "dhnt@example.com"},
		{"config", "user.name", "dhnt"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		_ = cmd.Run()
	}
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}
