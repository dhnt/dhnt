// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"reflect"
	"testing"
)

// TestFoldBranch_PreservesSpecAndRunsRightArm: folding wraps the original
// steps in a branch under a named condition; the derived version keeps
// the original contract + cap, round-trips, and runs the then-arm when
// the condition holds, the else-arm otherwise.
func TestFoldBranch_PreservesSpecAndRunsRightArm(t *testing.T) {
	orig := Skill{
		Name:      "taso",
		EffectCap: []Effect{EffRead, EffWrite},
		Contract:  []Check{{Predicate: "doniu"}},
		Steps:     []Step{{Name: "alile", Primitive: "borokeni"}}, // original (fails on the new platform)
	}
	fixed := []Step{{Name: "alile", Primitive: "fikiso"}}
	folded, err := FoldBranch(orig, Check{Predicate: "bisidi"}, fixed)
	if err != nil {
		t.Fatalf("FoldBranch: %v", err)
	}
	// spec preserved
	if !reflect.DeepEqual(folded.Contract, orig.Contract) || !reflect.DeepEqual(folded.EffectCap, orig.EffectCap) {
		t.Errorf("fold changed the spec")
	}
	if len(folded.Steps) != 1 || folded.Steps[0].Branch == nil {
		t.Fatalf("fold did not produce a single branch")
	}
	// round-trips
	enc, err := LineariseDhnt(folded)
	if err != nil {
		t.Fatalf("linearise: %v", err)
	}
	if parsed, err := ParseDhnt(enc); err != nil || !reflect.DeepEqual(normaliseSkill(parsed), normaliseSkill(folded)) {
		t.Fatalf("folded skill does not round-trip: %v", err)
	}

	// run with bisidi=true → then-arm (fikiso) → success
	flagThen := false
	envThen := foldEnv(&flagThen, true)
	att, err := Run(folded, envThen, "bsd")
	if err != nil || !att.Valid {
		t.Errorf("then-arm: err=%v valid=%v", err, att.Valid)
	}
	if !flagThen {
		t.Errorf("then-arm did not run fikiso")
	}

	// run with bisidi=false → else-arm (borokeni) → fails (as the original did)
	flagElse := false
	envElse := foldEnv(&flagElse, false)
	if _, err := Run(folded, envElse, "gnu"); err == nil {
		t.Errorf("else-arm should have run the failing original primitive")
	}
}

// foldEnv binds the broken/fixed primitives, the doniu predicate, and a
// `bisidi` (is-bsd) condition predicate fixed to isBSD.
func foldEnv(flag *bool, isBSD bool) Env {
	return Env{
		Primitives: map[string]PrimitiveFn{
			"borokeni": func(a []Arg) ([]Effect, error) { return nil, errFailEnv },
			"fikiso":   func(a []Arg) ([]Effect, error) { *flag = true; return []Effect{EffWrite}, nil },
		},
		Predicates: map[string]PredicateFn{
			"doniu":  func(a []Arg) (bool, []Effect, error) { return *flag, []Effect{EffRead}, nil },
			"bisidi": func(a []Arg) (bool, []Effect, error) { return isBSD, []Effect{EffRead}, nil },
		},
	}
}

var errFailEnv = &simpleErr{"not supported in this environment"}

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }

// TestEnvMatchFold runs an env-fingerprint-keyed fold: the built-in
// context predicate selects the then-arm only in the matching environment.
func TestEnvMatchFold(t *testing.T) {
	probes := []EnvProbe{{"os", "darwin"}, {"bash", "5.2"}}
	orig := Skill{
		Name:      "taso",
		EffectCap: []Effect{EffRead, EffWrite},
		Contract:  []Check{{Predicate: "doniu"}},
		Steps:     []Step{{Name: "alile", Primitive: "borokeni"}},
	}
	folded, err := FoldForContext(orig, probes, []Step{{Name: "alile", Primitive: "fikiso"}})
	if err != nil {
		t.Fatalf("FoldForContext: %v", err)
	}
	// matching env → then-arm
	flag := false
	env := WithEnvPredicate(foldEnv(&flag, false), probes) // bisidi unused here
	att, err := Run(folded, env, "darwin")
	if err != nil || !att.Valid || !flag {
		t.Errorf("matching env should run the fix: err=%v valid=%v flag=%v", err, att.Valid, flag)
	}
	// different env → else-arm (original, fails)
	flag2 := false
	env2 := WithEnvPredicate(foldEnv(&flag2, false), []EnvProbe{{"os", "linux"}})
	if _, err := Run(folded, env2, "linux"); err == nil {
		t.Errorf("non-matching env should run the failing original")
	}
}

func TestVersionStore_ResolveLatest(t *testing.T) {
	orig := Skill{Name: "taso", Contract: []Check{{Predicate: "doniu"}}, Steps: []Step{{Name: "alile", Primitive: "borokeni"}}}
	folded, err := FoldBranch(orig, Check{Predicate: "bisidi"}, []Step{{Name: "alile", Primitive: "fikiso"}})
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := Identity(orig)
	canon, _ := LineariseDhnt(folded)

	for _, vs := range []VersionStore{NewMemVersionStore(), &FileVersionStore{Dir: t.TempDir()}} {
		// before save → resolves to original
		if s, ok := ResolveLatest(vs, orig); ok {
			t.Errorf("resolved a version before any save: %v", s)
		}
		if err := vs.Save(DerivedVersion{ParentID: pid, ID: "h2", Canonical: canon, ContextKey: "cX"}); err != nil {
			t.Fatalf("Save: %v", err)
		}
		got, ok := ResolveLatest(vs, orig)
		if !ok {
			t.Fatalf("ResolveLatest miss after save")
		}
		if g, _ := LineariseDhnt(got); g != canon {
			t.Errorf("resolved wrong version")
		}
	}
}
