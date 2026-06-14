// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package tui_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

// TestDrive_AiderEndToEnd drives the real `aider` agentic CLI through the
// dhnt branching skill over a PTY. It spends no model tokens (repo-map
// off; only /help and /exit, which do not call the LLM) and is gated
// behind DHNT_AIDER_E2E so normal `go test` stays fast and hermetic.
func TestDrive_AiderEndToEnd(t *testing.T) {
	if os.Getenv("DHNT_AIDER_E2E") == "" {
		t.Skip("set DHNT_AIDER_E2E=1 to run the live aider end-to-end test")
	}
	if _, err := exec.LookPath("aider"); err != nil {
		t.Skip("aider not on PATH")
	}
	dir := t.TempDir()
	for _, a := range [][]string{{"init", "-q"}, {"config", "user.email", "d@e.com"}, {"config", "user.name", "d"}} {
		c := exec.Command("git", a...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatalf("git %v: %v", a, err)
		}
	}

	// spawn → expect first prompt → branch: decline the gitignore prompt
	// if shown → /help → confirm the listing → /exit.
	skill := skills.Skill{
		Name:      "darive",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffNet, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: tui.PrimSpawn},
			{Name: "se", Primitive: tui.PrimExpect, Args: ref("readayu")},
			{Branch: &skills.Branch{
				Cond: skills.Check{Predicate: tui.PredSeen, Args: ref("marokero")},
				Then: []skills.Step{{Name: "ta", Primitive: tui.PrimSend, Args: ref("nope")}},
			}},
			{Name: "si", Primitive: tui.PrimSend, Args: ref("helepe")},
			{Name: "so", Primitive: tui.PrimExpect, Args: ref("resulito")},
			{Name: "su", Primitive: tui.PrimQuit},
		},
		Contract: []skills.Check{{Predicate: tui.PredClean}},
	}
	spec := tui.Spec{
		Argv: []string{"aider", "--no-pretty", "--no-stream", "--no-check-update", "--no-auto-commits", "--map-tokens", "0"},
		Dir:  dir,
		Patterns: map[string]string{
			"readayu":  `(?i)(gitignore|Aider v|>)`,
			"marokero": `(?i)gitignore`,
			"resulito": `(?i)(/help|/tokens|/exit|/add)`,
		},
		Inputs:  map[string]string{"nope": "n\n", "helepe": "/help\n"},
		Quit:    "/exit\n",
		Timeout: 60 * time.Second,
	}

	env, sess, err := tui.NewEnv(spec)
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}
	defer sess.Close()

	att, err := skills.Run(skill, env, "aider")
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, sess.Output())
	}
	if !att.Valid || !att.Consistent(skill) {
		t.Errorf("aider run not valid/consistent: %+v\noutput:\n%s", att, sess.Output())
	}
	if !strings.Contains(sess.Output(), "/tokens") {
		t.Errorf("expected aider /help listing in output")
	}
}
