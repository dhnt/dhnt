// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import "testing"

// TestRuntime_SelfHealsAndAccretes drives the full Phase-2 loop: a skill
// that fails in each new environment is repaired, the fix is folded into a
// host-overlay version, and subsequent runs in that environment reuse it
// without repair. Across two environments the version accretes both arms.
func TestRuntime_SelfHealsAndAccretes(t *testing.T) {
	g := loadSeedGlossary(t)
	vs := NewMemVersionStore()

	calls := 0
	rt := &Runtime{
		Glossary: g, Lang: "en", Versions: vs, Tier: "host",
		Repair: &Repairer{
			Glossary: g, Lang: "en", MaxAttempts: 2,
			Complete: func(prompt string) (string, error) {
				calls++
				return "<dhnt>skill taso step alile fikiso</dhnt>", nil
			},
		},
	}

	mkEnv := func(flag *bool) Env {
		return Env{
			Primitives: map[string]PrimitiveFn{
				"borokeni": func(a []Arg) ([]Effect, error) { return nil, errFailEnv },
				"fikiso":   func(a []Arg) ([]Effect, error) { *flag = true; return []Effect{EffWrite}, nil },
			},
			Predicates: map[string]PredicateFn{
				"doniu": func(a []Arg) (bool, []Effect, error) { return *flag, []Effect{EffRead}, nil },
			},
		}
	}
	run := func(probes []EnvProbe) (Attestation, RunOutcome) {
		flag := false
		rt.Probes = probes
		att, outcome, err := rt.Run(brokenSkill(), mkEnv(&flag))
		if outcome != OutcomeFailed && err != nil {
			t.Fatalf("run %v: %v", probes, err)
		}
		return att, outcome
	}

	ctxA := []EnvProbe{{"os", "darwin"}, {"bash", "5.2"}}
	ctxB := []EnvProbe{{"os", "linux"}, {"bash", "5.1"}}

	// 1. ctxA: baseline fails → repair → fold → save.
	if att, o := run(ctxA); o != OutcomeRepaired || !att.Valid {
		t.Fatalf("ctxA run1: outcome=%v valid=%v", o, att.Valid)
	}
	if calls != 1 {
		t.Fatalf("ctxA run1: calls=%d want 1", calls)
	}

	// 2. ctxA again: the folded version is reused — no repair.
	if att, o := run(ctxA); o != OutcomeCached || !att.Valid {
		t.Errorf("ctxA run2: outcome=%v valid=%v want cached", o, att.Valid)
	}
	if calls != 1 {
		t.Errorf("ctxA run2: calls=%d want still 1", calls)
	}

	// 3. ctxB: the folded version's else-arm (original) fails here → repair
	//    again, folding a SECOND arm onto the version.
	if att, o := run(ctxB); o != OutcomeRepaired || !att.Valid {
		t.Errorf("ctxB run3: outcome=%v valid=%v want repaired", o, att.Valid)
	}
	if calls != 2 {
		t.Errorf("ctxB run3: calls=%d want 2", calls)
	}

	// 4. ctxA again: the now-accreted version still handles ctxA — no repair.
	if att, o := run(ctxA); o != OutcomeCached || !att.Valid {
		t.Errorf("ctxA run4: outcome=%v valid=%v want cached", o, att.Valid)
	}
	if calls != 2 {
		t.Errorf("ctxA run4: calls=%d want still 2 (both arms learned)", calls)
	}

	// 5. ctxB again: also reused now.
	if att, o := run(ctxB); o != OutcomeCached || !att.Valid {
		t.Errorf("ctxB run5: outcome=%v valid=%v want cached", o, att.Valid)
	}
	if calls != 2 {
		t.Errorf("ctxB run5: calls=%d want still 2", calls)
	}
}
