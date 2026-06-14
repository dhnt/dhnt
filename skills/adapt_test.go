// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"fmt"
	"testing"
)

func TestContextKey_StableAndDistinct(t *testing.T) {
	a := ContextKey([]EnvProbe{{"os", "darwin"}, {"bash", "5.2"}})
	b := ContextKey([]EnvProbe{{"bash", "5.2"}, {"os", "darwin"}}) // order-independent
	if a != b {
		t.Errorf("ContextKey not order-independent: %q vs %q", a, b)
	}
	c := ContextKey([]EnvProbe{{"os", "linux"}, {"bash", "5.2"}})
	if a == c {
		t.Errorf("different env produced same key")
	}
}

// adaptEnv builds an Env whose `borokeni` primitive fails (simulating a
// platform/version incompatibility) and whose `fikiso` primitive works
// (the fix). The `doniu` predicate reports whether the work was done.
// flag is shared so the predicate observes the primitives' effect.
func adaptEnv(flag *bool) Env {
	return Env{
		Primitives: map[string]PrimitiveFn{
			"borokeni": func(args []Arg) ([]Effect, error) {
				return nil, fmt.Errorf("not supported in this environment")
			},
			"fikiso": func(args []Arg) ([]Effect, error) {
				*flag = true
				return []Effect{EffWrite}, nil
			},
			"rogue": func(args []Arg) ([]Effect, error) {
				*flag = true
				return []Effect{EffWrite, EffDestroy}, nil // exceeds the cap
			},
		},
		Predicates: map[string]PredicateFn{
			"doniu": func(args []Arg) (bool, []Effect, error) {
				return *flag, []Effect{EffRead}, nil
			},
		},
	}
}

func brokenSkill() Skill {
	return Skill{
		Name:      "taso",
		EffectCap: []Effect{EffRead, EffWrite},
		Contract:  []Check{{Predicate: "doniu"}},
		Steps:     []Step{{Name: "alile", Primitive: "borokeni"}},
	}
}

func TestRunAdaptive_LearnsRepairsAndReuses(t *testing.T) {
	g := loadSeedGlossary(t)
	store := NewMemStore()
	probes := []EnvProbe{{"os", "darwin"}, {"bash", "5.2"}}

	// The model's "fix": use the working primitive instead.
	repairCalls := 0
	repair := &Repairer{
		Glossary: g, Lang: "en", MaxAttempts: 2,
		Complete: func(prompt string) (string, error) {
			repairCalls++
			return "here:\n<dhnt>skill taso step alile fikiso</dhnt>", nil
		},
	}

	// Run 1: baseline fails → repair → verify → cache.
	flag1 := false
	att, outcome, err := RunAdaptive(brokenSkill(), adaptEnv(&flag1), "agent-1", probes, store, repair)
	if err != nil {
		t.Fatalf("run 1: %v", err)
	}
	if outcome != OutcomeRepaired || !att.Valid {
		t.Fatalf("run 1: outcome=%v valid=%v, want repaired+valid", outcome, att.Valid)
	}
	if repairCalls != 1 {
		t.Fatalf("run 1: repairCalls=%d, want 1", repairCalls)
	}

	// Run 2: cache hit → reuse the learned variant, no repair.
	flag2 := false
	att, outcome, err = RunAdaptive(brokenSkill(), adaptEnv(&flag2), "agent-1", probes, store, repair)
	if err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if outcome != OutcomeCached || !att.Valid {
		t.Errorf("run 2: outcome=%v valid=%v, want cached+valid", outcome, att.Valid)
	}
	if repairCalls != 1 {
		t.Errorf("run 2: repairCalls=%d, want still 1 (cache should serve it)", repairCalls)
	}

	// A DIFFERENT agent, same store + context: must reuse without repairing.
	noRepair := &Repairer{
		Glossary: g, Lang: "en",
		Complete: func(prompt string) (string, error) {
			t.Fatalf("second agent should not need to repair")
			return "", nil
		},
	}
	flag3 := false
	_, outcome, err = RunAdaptive(brokenSkill(), adaptEnv(&flag3), "agent-2", probes, store, noRepair)
	if err != nil {
		t.Fatalf("agent-2: %v", err)
	}
	if outcome != OutcomeCached {
		t.Errorf("agent-2: outcome=%v, want cached", outcome)
	}

	// A DIFFERENT context: cache miss → repairs again (separate key).
	repairCalls = 0
	flag4 := false
	_, outcome, err = RunAdaptive(brokenSkill(), adaptEnv(&flag4), "agent-1",
		[]EnvProbe{{"os", "linux"}, {"bash", "5.1"}}, store, repair)
	if err != nil {
		t.Fatalf("other context: %v", err)
	}
	if outcome != OutcomeRepaired || repairCalls != 1 {
		t.Errorf("other context: outcome=%v repairCalls=%d, want repaired+1", outcome, repairCalls)
	}
}

// TestRunAdaptive_RejectsOverCapRepair: a "fix" that exceeds the original
// effect cap must NOT be accepted or cached — the spec is the law.
func TestRunAdaptive_RejectsOverCapRepair(t *testing.T) {
	g := loadSeedGlossary(t)
	store := NewMemStore()
	probes := []EnvProbe{{"os", "darwin"}}

	repair := &Repairer{
		Glossary: g, Lang: "en", MaxAttempts: 1,
		Complete: func(prompt string) (string, error) {
			// satisfies the contract but causes a forbidden `destroy` effect
			return "<dhnt>skill taso step alile rogue</dhnt>", nil
		},
	}
	flag := false
	_, outcome, err := RunAdaptive(brokenSkill(), adaptEnv(&flag), "agent-1", probes, store, repair)
	if outcome != OutcomeFailed || err == nil {
		t.Fatalf("over-cap repair: outcome=%v err=%v, want failed", outcome, err)
	}
	if _, ok, _ := store.Get(mustID(t), ContextKey(probes)); ok {
		t.Errorf("an over-cap variant was cached")
	}
}

func TestFileStore_RoundTrip(t *testing.T) {
	store := &FileStore{Dir: t.TempDir()}
	a := Adaptation{SkillID: "h123", ContextKey: "cabc", Variant: "sokilili taso fini", Attest: Attestation{Valid: true, Tier: "x"}}
	if err := store.Put(a); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, ok, err := store.Get("h123", "cabc")
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if got.Variant != a.Variant || !got.Attest.Valid {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if _, ok, _ := store.Get("h123", "missing"); ok {
		t.Errorf("expected miss")
	}
}

func mustID(t *testing.T) string {
	t.Helper()
	id, err := Identity(brokenSkill())
	if err != nil {
		t.Fatal(err)
	}
	return id
}
