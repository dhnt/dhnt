// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"strings"
	"testing"
)

// TestRun_RepeatedPredicateDistinctArgs guards the fix for predicate-id
// collision: a contract that uses one predicate twice with DIFFERENT args
// must evaluate each independently. Here `okp(value=good)` passes and
// `okp(value=bad)` fails, so the run must be invalid — and would wrongly
// pass if results were keyed by predicate id alone.
func TestRun_RepeatedPredicateDistinctArgs(t *testing.T) {
	skill := Skill{
		Name: "taso",
		Contract: []Check{
			{Predicate: "veri", Args: []Arg{{Name: "value", Value: NewRef("gudo")}}},
			{Predicate: "veri", Args: []Arg{{Name: "value", Value: NewRef("bado")}}},
		},
	}
	env := Env{Predicates: map[string]PredicateFn{
		"veri": func(args []Arg) (bool, []Effect, error) {
			ref := ""
			for _, a := range args {
				if a.Name == "value" {
					ref = a.Value.Ref
				}
			}
			return ref == "gudo", nil, nil
		},
	}}
	att, err := Run(skill, env, "t")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if att.Valid {
		t.Errorf("one check failed; run must be invalid, got valid (passed=%v failed=%v)", att.Passed, att.Failed)
	}
	if len(att.Failed) != 1 {
		t.Errorf("expected exactly the bad check to fail, got failed=%v", att.Failed)
	}
	if !att.Consistent(skill) {
		t.Errorf("attestation should be self-consistent")
	}
}

func contractedSkill() Skill {
	return Skill{
		Name:      "rilease",
		EffectCap: []Effect{EffRead, EffWrite},
		Contract:  []Check{{Predicate: "gereeni"}, {Predicate: "sigeneda"}},
	}
}

// TestIdentity_StableAndDistinct: equal canonical form ⇒ equal identity;
// any change to the canonical form changes the identity (pillar P0).
func TestIdentity_StableAndDistinct(t *testing.T) {
	a := contractedSkill()
	id1, err := Identity(a)
	if err != nil {
		t.Fatalf("Identity: %v", err)
	}
	id2, _ := Identity(contractedSkill())
	if id1 != id2 {
		t.Errorf("identity not stable: %q vs %q", id1, id2)
	}
	if !strings.HasPrefix(id1, "h") {
		t.Errorf("identity missing 'h' prefix: %q", id1)
	}
	b := contractedSkill()
	b.Contract = append(b.Contract, Check{Predicate: "builida"})
	id3, _ := Identity(b)
	if id3 == id1 {
		t.Errorf("distinct skills share identity %q", id1)
	}
}

func TestAttest_ValidRun(t *testing.T) {
	s := contractedSkill()
	a, err := Attest(s, "claude", map[string]bool{"gereeni": true, "sigeneda": true},
		[]Effect{EffRead, EffWrite}, "tagged v1.2.3")
	if err != nil {
		t.Fatalf("Attest: %v", err)
	}
	if !a.Valid {
		t.Errorf("expected Valid run, got Failed=%v", a.Failed)
	}
	if len(a.Passed) != 2 || len(a.Failed) != 0 {
		t.Errorf("unexpected passed/failed: %v / %v", a.Passed, a.Failed)
	}
	if !a.Consistent(s) {
		t.Errorf("valid attestation reported inconsistent")
	}
}

func TestAttest_FailedCheck(t *testing.T) {
	s := contractedSkill()
	a, _ := Attest(s, "weak-model", map[string]bool{"gereeni": true, "sigeneda": false},
		[]Effect{EffRead}, "")
	if a.Valid {
		t.Errorf("expected invalid run when a contract check fails")
	}
	if len(a.Failed) != 1 || a.Failed[0] != "sigeneda" {
		t.Errorf("expected sigeneda in Failed, got %v", a.Failed)
	}
	if !a.Consistent(s) {
		t.Errorf("honest failing attestation reported inconsistent")
	}
}

func TestAttest_EffectOverflowInvalidatesRun(t *testing.T) {
	s := contractedSkill()
	// destroy is outside the {read,write} cap
	a, _ := Attest(s, "rogue", map[string]bool{"gereeni": true, "sigeneda": true},
		[]Effect{EffRead, EffWrite, EffDestroy}, "")
	if a.Valid {
		t.Errorf("expected invalid run when effects exceed the cap")
	}
	if !a.Consistent(s) {
		t.Errorf("honest over-cap attestation reported inconsistent")
	}
}

// TestAttest_DetectsTamperedValid: flipping Valid to true on a receipt
// whose evidence says otherwise must be caught by Consistent.
func TestAttest_DetectsTamperedValid(t *testing.T) {
	s := contractedSkill()
	a, _ := Attest(s, "liar", map[string]bool{"gereeni": true, "sigeneda": false},
		[]Effect{EffRead}, "")
	a.Valid = true // forge the verdict
	if a.Consistent(s) {
		t.Errorf("Consistent failed to catch tampered Valid flag")
	}
}
