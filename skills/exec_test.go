// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// This is an applied end-to-end test, not a structural one: it runs a
// real skill against the real filesystem through the Layer 3 executor,
// and shows that ONE contract levels TWO executor tiers — a diligent
// one converges (Valid), a lazy one is caught, and a rogue one that
// exceeds the effect cap is caught too.
func TestRun_ContractLevelsExecutors(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "release.txt")

	// The skill: leave release.txt present, using only read+write.
	skill := Skill{
		Name:      "rilease",
		EffectCap: []Effect{EffRead, EffWrite},
		Contract:  []Check{{Predicate: "poresenito"}},
		Steps:     []Step{{Name: "wurite", Primitive: "wurite"}},
	}

	// Real predicate binding: does the file exist? (a read effect)
	filePresent := func(args []Arg) (bool, []Effect, error) {
		_, err := os.Stat(target)
		return err == nil, []Effect{EffRead}, nil
	}

	// Tier A — diligent: actually writes the file (a write effect).
	diligent := Env{
		Primitives: map[string]PrimitiveFn{
			"wurite": func(args []Arg) ([]Effect, error) {
				return []Effect{EffWrite}, os.WriteFile(target, []byte("v1\n"), 0o644)
			},
		},
		Predicates: map[string]PredicateFn{"poresenito": filePresent},
	}
	a, err := Run(skill, diligent, "diligent")
	if err != nil {
		t.Fatalf("diligent Run: %v", err)
	}
	if !a.Valid {
		t.Errorf("diligent run should be Valid; failed=%v effects=%v", a.Failed, a.Effects)
	}
	if !a.Consistent(skill) {
		t.Errorf("diligent attestation inconsistent")
	}

	// Tier B — lazy: claims to work but writes nothing. Same contract.
	_ = os.Remove(target)
	lazy := Env{
		Primitives: map[string]PrimitiveFn{
			"wurite": func(args []Arg) ([]Effect, error) { return nil, nil }, // does nothing
		},
		Predicates: map[string]PredicateFn{"poresenito": filePresent},
	}
	b, err := Run(skill, lazy, "lazy")
	if err != nil {
		t.Fatalf("lazy Run: %v", err)
	}
	if b.Valid {
		t.Errorf("lazy run must be caught by the contract, got Valid")
	}

	// Both attestations are for the same skill identity — same spec, two
	// tiers, divergent reality, one verdict each.
	if a.Skill != b.Skill {
		t.Errorf("attestations target different identities: %q vs %q", a.Skill, b.Skill)
	}

	// Tier C — rogue: does the work but also deletes something (a destroy
	// effect outside the {read,write} cap). The cap catches it.
	_ = os.Remove(target)
	rogue := Env{
		Primitives: map[string]PrimitiveFn{
			"wurite": func(args []Arg) ([]Effect, error) {
				if err := os.WriteFile(target, []byte("v1\n"), 0o644); err != nil {
					return nil, err
				}
				return []Effect{EffWrite, EffDestroy}, nil // over-reach
			},
		},
		Predicates: map[string]PredicateFn{"poresenito": filePresent},
	}
	c, err := Run(skill, rogue, "rogue")
	if err != nil {
		t.Fatalf("rogue Run: %v", err)
	}
	if c.Valid {
		t.Errorf("rogue run exceeded the effect cap but was marked Valid")
	}
}

// TestRun_MissingBindingErrors: an executor cannot silently skip a step
// or a check.
func TestRun_MissingBindingErrors(t *testing.T) {
	skill := Skill{
		Name:     "rilease",
		Contract: []Check{{Predicate: "poresenito"}},
		Steps:    []Step{{Name: "x", Primitive: "wurite"}},
	}
	if _, err := Run(skill, Env{}, "broken"); err == nil {
		t.Errorf("expected error for missing primitive binding")
	}
}
