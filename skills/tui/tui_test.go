// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package tui_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

func ref(k string) []skills.Arg {
	return []skills.Arg{{Name: "value", Value: skills.NewRef(k)}}
}

// driveSkill is the abstract, tool-agnostic TUI protocol as a dhnt skill:
// spawn → expect READY → send INPUT → expect RESULT → quit, with the
// contract that the tool exited cleanly. The concrete patterns/inputs
// come from the Spec, so this one skill drives any terminal tool.
func driveSkill() skills.Skill {
	return skills.Skill{
		Name:      "darive",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: tui.PrimSpawn},
			{Name: "se", Primitive: tui.PrimExpect, Args: ref("readayu")},  // expect ready
			{Name: "si", Primitive: tui.PrimSend, Args: ref("iniputo")},    // send input
			{Name: "so", Primitive: tui.PrimExpect, Args: ref("resulito")}, // expect result
			{Name: "su", Primitive: tui.PrimQuit},                          // quit
		},
		Contract: []skills.Check{{Predicate: tui.PredClean}},
	}
}

// TestDrive_RealPTY drives `cat` through a real pseudo-terminal: send a
// line, expect the echo, EOF to exit. One dhnt skill, real TTY, verified.
func TestDrive_RealPTY(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY driver is unix-only")
	}
	spec := tui.Spec{
		Argv:     []string{"cat"},
		Patterns: map[string]string{"readayu": "", "resulito": "PONG"},
		Inputs:   map[string]string{"iniputo": "PONG\n"},
		Quit:     "", // EOF via Ctrl-D
		Timeout:  3 * time.Second,
	}
	env, sess, err := tui.NewEnv(spec)
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}
	defer sess.Close()

	att, err := skills.Run(driveSkill(), env, "cat-driver")
	if err != nil {
		t.Fatalf("Run: %v\noutput:\n%s", err, sess.Output())
	}
	if !att.Valid {
		t.Errorf("expected Valid run; failed=%v effects=%v output=%q", att.Failed, att.Effects, sess.Output())
	}
	if !att.Consistent(driveSkill()) {
		t.Errorf("attestation inconsistent")
	}
}

// TestDrive_ExpectTimeoutCaught: if the tool never produces the expected
// output, the expect checkpoint fails — the run is caught, not silently
// reported as success.
func TestDrive_ExpectTimeoutCaught(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY driver is unix-only")
	}
	spec := tui.Spec{
		Argv:     []string{"cat"},
		Patterns: map[string]string{"readayu": "", "resulito": "WILL_NOT_APPEAR"},
		Inputs:   map[string]string{"iniputo": "PONG\n"},
		Timeout:  400 * time.Millisecond,
	}
	env, sess, err := tui.NewEnv(spec)
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}
	defer sess.Close()

	if _, err := skills.Run(driveSkill(), env, "cat-driver"); err == nil {
		t.Errorf("expected expect-timeout error, got nil")
	}
}
