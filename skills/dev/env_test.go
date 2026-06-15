// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
)

func TestSanitizedPATH_DropsWrappers(t *testing.T) {
	old := os.Getenv("PATH")
	defer os.Setenv("PATH", old)
	os.Setenv("PATH", "/tmp/ycode-wrap/abc/bin:/opt/homebrew/bin:/usr/bin")
	got := dev.SanitizedPATH()
	if strings.Contains(got, "ycode-wrap") {
		t.Errorf("wrapper dir not removed: %q", got)
	}
	if !strings.Contains(got, "/opt/homebrew/bin") || !strings.Contains(got, "/usr/bin") || !strings.Contains(got, "/bin") {
		t.Errorf("expected real dirs retained/ensured: %q", got)
	}
}

// TestCommandEnv: a Command.Env override reaches the executed process.
func TestCommandEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	spec := dev.Spec{Commands: map[string]dev.Command{
		"te": {Argv: []string{"sh", "-c", `[ "$DHNT_T" = ok ]`}, Env: []string{"DHNT_T=ok"}},
		"fi": {Argv: []string{"true"}},
		"ve": {Argv: []string{"true"}},
		"fa": {Argv: []string{"true"}},
	}}
	env, _, err := dev.NewEnv(spec)
	if err != nil {
		t.Fatal(err)
	}
	att, err := skills.Run(dev.GoCheckSkill(), env, "env")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !att.Valid {
		t.Errorf("Command.Env not applied: the test command should see DHNT_T=ok (failed=%v)", att.Failed)
	}
}
