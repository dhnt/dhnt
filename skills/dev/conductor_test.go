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
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
)

// These tests stub the `ycode weave` orchestration surface with POSIX
// true/false/echo so they are hermetic (no fleet, no agent CLIs, no
// tokens) yet exercise the real executor: the goal-met ∧ converged
// contract, the complexity-gated RESEARCH branch, and the effect cap.
// The command-id literals mirror conductor.go's unexported consts.

func conductorSpec(over map[string]dev.Command) dev.Spec {
	base := map[string]dev.Command{
		"pa": cmd("true"),  // PLAN
		"bo": cmd("false"), // not complex → RESEARCH skipped (common case)
		"re": cmd("true"),  // RESEARCH
		"fa": cmd("true"),  // FAN-OUT
		"wo": cmd("true"),  // STEER
		"vo": cmd("true"),  // CONVERGE
		"ru": cmd("true"),  // RETRO
		"go": cmd("true"),  // goal-met
		"cu": cmd("true"),  // converged
	}
	for k, v := range over {
		base[k] = v
	}
	return dev.Spec{Commands: base}
}

func runConductor(t *testing.T, spec dev.Spec) (skills.Attestation, error) {
	t.Helper()
	env, _, err := dev.NewEnv(spec)
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}
	return skills.Run(dev.ConductorSkill(), env, "test")
}

func TestConductor_GoalMet(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false")
	}
	att, err := runConductor(t, conductorSpec(nil))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !att.Valid {
		t.Fatalf("expected valid; att=%+v", att)
	}
	if !att.Consistent(dev.ConductorSkill()) {
		t.Errorf("attestation not consistent with the skill: %+v", att)
	}
}

func TestConductor_GoalUnmet(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false")
	}
	att, err := runConductor(t, conductorSpec(map[string]dev.Command{"go": cmd("false")}))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if att.Valid {
		t.Errorf("goal verifier failed but skill reported valid")
	}
	if !slices.Contains(att.Failed, "exito(value=go)") {
		t.Errorf("expected goal-met check in Failed; got %v", att.Failed)
	}
}

func TestConductor_NotConverged(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false")
	}
	att, err := runConductor(t, conductorSpec(map[string]dev.Command{"cu": cmd("false")}))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if att.Valid {
		t.Errorf("work unconverged but skill reported valid")
	}
	if !slices.Contains(att.Failed, "exito(value=cu)") {
		t.Errorf("expected converged check in Failed; got %v", att.Failed)
	}
}

// TestConductor_ResearchBranch proves the complexity-gated RESEARCH step
// runs iff the goal is complex: the research stub touches a sentinel file,
// which must exist when complex (bo=true) and not when simple (bo=false).
func TestConductor_ResearchBranch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX sh/touch")
	}
	for _, tc := range []struct {
		name    string
		complex string // bo command
		want    bool
	}{
		{"complex-runs-research", "true", true},
		{"simple-skips-research", "false", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			marker := filepath.Join(dir, "researched.marker")
			spec := conductorSpec(map[string]dev.Command{
				"bo": cmd(tc.complex),
				"re": cmd("sh", "-c", "touch "+marker),
			})
			spec.Dir = dir
			att, err := runConductor(t, spec)
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			if !att.Valid {
				t.Fatalf("expected valid; att=%+v", att)
			}
			_, statErr := os.Stat(marker)
			got := statErr == nil
			if got != tc.want {
				t.Errorf("research-ran=%v, want %v (complex=%s)", got, tc.want, tc.complex)
			}
		})
	}
}

func TestConductor_PhaseCommandMissing(t *testing.T) {
	att, err := runConductor(t, conductorSpec(map[string]dev.Command{
		"fa": cmd("definitely-not-a-real-binary-xyz"),
	}))
	if err == nil {
		t.Errorf("expected error when a phase command is missing; att=%+v", att)
	}
}

// TestConductor_Linearises proves the skill linearises to canonical L1.5,
// has a stable identity, and round-trips through ParseDhnt unchanged (the
// transpilability rule).
func TestConductor_Linearises(t *testing.T) {
	s := dev.ConductorSkill()
	canon, err := skills.LineariseDhnt(s)
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	id, err := skills.Identity(s)
	if err != nil {
		t.Fatalf("Identity: %v", err)
	}
	if id == "" || canon == "" {
		t.Fatalf("empty canon/identity")
	}
	parsed, err := skills.ParseDhnt(canon)
	if err != nil {
		t.Fatalf("ParseDhnt(%q): %v", canon, err)
	}
	got, err := skills.LineariseDhnt(parsed)
	if err != nil {
		t.Fatalf("re-LineariseDhnt: %v", err)
	}
	if got != canon {
		t.Errorf("round-trip not byte-stable:\n  in:  %q\n  out: %q", canon, got)
	}
}
