// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev_test

import (
	"runtime"
	"testing"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
)

// These tests bind go-verify's command ids to POSIX true/false/echo so they
// are hermetic (no Go toolchain dependency) yet exercise the real executor:
// a step that must exit 0, an exit-zero predicate, and an empty-output
// predicate.
func cmd(argv ...string) dev.Command {
	return dev.Command{Argv: argv, Effects: []skills.Effect{skills.EffRead}}
}

func run(t *testing.T, spec dev.Spec) (skills.Attestation, error) {
	t.Helper()
	env, _, err := dev.NewEnv(spec)
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}
	return skills.Run(dev.GoVerifySkill(), env, "test")
}

func TestGoVerify_AllGreen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false/echo")
	}
	spec := dev.Spec{Commands: map[string]dev.Command{
		"fa": cmd("true"), // gofmt -w
		"fi": cmd("true"), // gofmt -l → empty output (clean)
		"ve": cmd("true"), // go vet ok
		"te": cmd("true"), // tests ok
	}}
	att, err := run(t, spec)
	if err != nil || !att.Valid {
		t.Fatalf("expected valid; err=%v att=%+v", err, att)
	}
}

func TestGoVerify_TestsFail(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false/echo")
	}
	spec := dev.Spec{Commands: map[string]dev.Command{
		"fa": cmd("true"), "fi": cmd("true"), "ve": cmd("true"),
		"te": cmd("false"), // tests fail
	}}
	att, err := run(t, spec)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if att.Valid {
		t.Errorf("tests failed but skill reported valid")
	}
}

func TestGoVerify_FmtDirty(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false/echo")
	}
	spec := dev.Spec{Commands: map[string]dev.Command{
		"fa": cmd("true"),
		"fi": cmd("echo", "needs_format.go"), // gofmt -l lists a file → not clean
		"ve": cmd("true"), "te": cmd("true"),
	}}
	att, err := run(t, spec)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if att.Valid {
		t.Errorf("unformatted file present but skill reported valid")
	}
}

func TestGoVerify_StepCommandMissing(t *testing.T) {
	spec := dev.Spec{Commands: map[string]dev.Command{
		"fa": cmd("definitely-not-a-real-binary-xyz"),
		"fi": cmd("true"), "ve": cmd("true"), "te": cmd("true"),
	}}
	if _, err := run(t, spec); err == nil {
		t.Errorf("expected error when a step command is missing")
	}
}
