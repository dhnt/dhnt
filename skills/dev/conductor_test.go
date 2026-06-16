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
		"vi": cmd("true"),  // reviewed
		"ni": cmd("false"), // not parallel-safe → route to SEQUENTIAL (the safe default)
		"lo": cmd("true"),  // SEQUENTIAL arm
		"tu": cmd("false"), // not stuck → ESCALATE skipped
		"ke": cmd("true"),  // ESCALATE arm
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

// fakeJudge returns a Completer that always replies with reply, so the
// judge path is exercised hermetically (no agent CLI, no tokens).
func fakeJudge(reply string) skills.Completer {
	return func(string) (string, error) { return reply, nil }
}

func runJudge(t *testing.T, spec dev.Spec) (skills.Attestation, error) {
	t.Helper()
	env, _, err := dev.NewEnv(spec)
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}
	return skills.Run(dev.ConductorJudgeSkill(), env, "test")
}

func TestConductor_JudgeMet(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false")
	}
	spec := conductorSpec(map[string]dev.Command{"go": cmd("echo", "feature shipped, tests green")})
	spec.Goal = "ship the feature"
	spec.Judge = fakeJudge("<verdict>YES</verdict>")
	att, err := runJudge(t, spec)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !att.Valid {
		t.Fatalf("judge said YES but skill reported invalid; att=%+v", att)
	}
	if !att.Consistent(dev.ConductorJudgeSkill()) {
		t.Errorf("attestation not consistent with the judge skill: %+v", att)
	}
}

func TestConductor_JudgeUnmet(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false")
	}
	spec := conductorSpec(map[string]dev.Command{"go": cmd("echo", "nothing done yet")})
	spec.Goal = "ship the feature"
	spec.Judge = fakeJudge("Reasoning: incomplete.\n<verdict>NO</verdict>")
	att, err := runJudge(t, spec)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if att.Valid {
		t.Errorf("judge said NO but skill reported valid")
	}
	if !slices.Contains(att.Failed, "meto(value=go)") {
		t.Errorf("expected judged goal-met check in Failed; got %v", att.Failed)
	}
}

// TestConductor_JudgeUnboundWithoutCompleter proves the judge skill needs a
// Spec.Judge: running it against a non-judge env is a hard error, not a
// silent pass (an executor cannot skip the goal-met check).
func TestConductor_JudgeUnboundWithoutCompleter(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false")
	}
	if _, err := runJudge(t, conductorSpec(nil)); err == nil {
		t.Errorf("expected error running judge skill without a Spec.Judge completer")
	}
}

func TestConductor_ReviewFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX true/false")
	}
	att, err := runConductor(t, conductorSpec(map[string]dev.Command{"vi": cmd("false")}))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if att.Valid {
		t.Errorf("review gate failed but skill reported valid")
	}
	if !slices.Contains(att.Failed, "exito(value=vi)") {
		t.Errorf("expected review check in Failed; got %v", att.Failed)
	}
}

// TestConductor_FanoutRouting proves the 2-way fan-out router: one open
// issue routes to SOLO, multiple route to FLEET (the else arm). Each arm's
// stub touches a distinct sentinel.
func TestConductor_FanoutRouting(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX sh/touch")
	}
	for _, tc := range []struct {
		name      string
		parallel  string // ni command (parallel-safe?)
		wantSolo  bool
		wantFleet bool
	}{
		// parallel-safe (many disjoint) → FLEET; else (single or shared
		// implementation) → SEQUENTIAL. Encodes the dogfood lesson that
		// parallel agents on shared source collide at merge.
		{"parallel-safe-routes-fleet", "true", false, true},
		{"shared-impl-routes-sequential", "false", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			soloMark := filepath.Join(dir, "solo.marker")
			fleetMark := filepath.Join(dir, "fleet.marker")
			spec := conductorSpec(map[string]dev.Command{
				"ni": cmd(tc.parallel),
				"lo": cmd("sh", "-c", "touch "+soloMark),
				"fa": cmd("sh", "-c", "touch "+fleetMark),
			})
			spec.Dir = dir
			att, err := runConductor(t, spec)
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			if !att.Valid {
				t.Fatalf("expected valid; att=%+v", att)
			}
			if _, e := os.Stat(soloMark); (e == nil) != tc.wantSolo {
				t.Errorf("solo-ran=%v, want %v", e == nil, tc.wantSolo)
			}
			if _, e := os.Stat(fleetMark); (e == nil) != tc.wantFleet {
				t.Errorf("fleet-ran=%v, want %v", e == nil, tc.wantFleet)
			}
		})
	}
}

// TestConductor_Composition proves the per-task P6 composition: the
// composed conductor's closure includes the task skill, and effect
// containment holds (the task's cap is within the conductor's).
func TestConductor_Composition(t *testing.T) {
	lib, composed := dev.ConductorLibrary()
	clo, err := lib.Closure(composed)
	if err != nil {
		t.Fatalf("closure: %v", err)
	}
	if !slices.Contains(clo, dev.ConductorTaskSkill().Name) {
		t.Errorf("closure %v should include the task skill %q", clo, dev.ConductorTaskSkill().Name)
	}
	viol, err := lib.EffectViolations(composed)
	if err != nil {
		t.Fatalf("effect-violations: %v", err)
	}
	if len(viol) != 0 {
		t.Errorf("expected no effect violations (task cap ⊆ conductor cap), got %v", viol)
	}
}

// TestConductor_CompositionCatchesOverCapTask proves the audit is real: a
// task whose effect cap exceeds the conductor's (here: destroy) is reported
// as a violation.
func TestConductor_CompositionCatchesOverCapTask(t *testing.T) {
	lib := skills.NewLibrary()
	bad := dev.ConductorTaskSkill()
	bad.EffectCap = append(bad.EffectCap, skills.EffDestroy) // wider than conductor's cap
	if err := lib.Add(bad); err != nil {
		t.Fatalf("add: %v", err)
	}
	composed := dev.ConductorComposedSkill()
	if err := lib.Add(composed); err != nil {
		t.Fatalf("add: %v", err)
	}
	viol, err := lib.EffectViolations(composed)
	if err != nil {
		t.Fatalf("effect-violations: %v", err)
	}
	if len(viol) == 0 {
		t.Errorf("an over-cap task should be reported as a violation")
	}
}

func TestConductor_DeclaresBlockersPolicy(t *testing.T) {
	if got := dev.ConductorSkill().OnFail; got != skills.PolicyBlockers {
		t.Errorf("ConductorSkill OnFail = %v, want PolicyBlockers", got)
	}
	if got := dev.ConductorJudgeSkill().OnFail; got != skills.PolicyBlockers {
		t.Errorf("ConductorJudgeSkill OnFail = %v, want PolicyBlockers", got)
	}
}

func TestConductor_PhaseCommandMissing(t *testing.T) {
	att, err := runConductor(t, conductorSpec(map[string]dev.Command{
		"pa": cmd("definitely-not-a-real-binary-xyz"), // PLAN always runs (unlike routed arms)
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
