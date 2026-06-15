// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package eval

import (
	"fmt"
	"testing"

	"github.com/dhnt/dhnt/skills"
)

// E2: a deterministic tier is perfectly stable; a varying-but-correct tier
// has variable output yet a stable verdict.
func TestE2_Repeatability(t *testing.T) {
	det := Tier{Name: "code-leaf", Model: model("42")}
	r, err := RunRepeatability(det, "q", `\b42\b`, 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.DistinctOutputs != 1 || r.DistinctVerdicts != 1 || !r.AllValid {
		t.Errorf("deterministic tier not stable: %+v", r)
	}

	n := 0
	vary := Tier{Name: "noisy", Model: func(string) (string, error) {
		n++
		return fmt.Sprintf("attempt %d: the answer is 42", n), nil
	}}
	r2, err := RunRepeatability(vary, "q", `\b42\b`, 5)
	if err != nil {
		t.Fatal(err)
	}
	if r2.DistinctOutputs != 5 {
		t.Errorf("expected varying output, got %d distinct", r2.DistinctOutputs)
	}
	if r2.DistinctVerdicts != 1 || !r2.AllValid {
		t.Errorf("verdict should be stable despite output variance: %+v", r2)
	}
}

// E3: each distinct context costs one repair; repeats are cached.
func TestE3_SelfHealingAmortization(t *testing.T) {
	g, err := skills.SeedGlossary()
	if err != nil {
		t.Fatal(err)
	}
	repair := func(string) (string, error) { return "<dhnt>skill taso step alile fikiso</dhnt>", nil }
	ctxA := []skills.EnvProbe{{Name: "os", Value: "darwin"}}
	ctxB := []skills.EnvProbe{{Name: "os", Value: "linux"}}
	schedule := [][]skills.EnvProbe{ctxA, ctxA, ctxB, ctxA, ctxB}
	s, err := RunSelfHealing(schedule, repair, g)
	if err != nil {
		t.Fatal(err)
	}
	if s.Invocations != 5 || s.Repairs != 2 || s.Cached != 3 {
		t.Errorf("amortization wrong: %+v (want 5 inv, 2 repairs, 3 cached)", s)
	}
}

// E4: acceptance equals within-cap; no over-cap variant slips through.
func TestE4_EffectContainment(t *testing.T) {
	cap := []skills.Effect{skills.EffRead, skills.EffWrite}
	candidates := [][]skills.Effect{
		{skills.EffRead},
		{skills.EffRead, skills.EffWrite},
		{skills.EffRead, skills.EffWrite, skills.EffDestroy},
		{skills.EffNet},
	}
	s := RunContainment(cap, candidates)
	if s.Accepted != 2 || s.Rejected != 2 || s.FalseAccepts != 0 {
		t.Errorf("containment wrong: %+v (want 2 accepted, 2 rejected, 0 false-accepts)", s)
	}
}

// E5: skills export to well-formed, round-trippable SKILL.md.
func TestE5_Interop(t *testing.T) {
	g, err := skills.SeedGlossary()
	if err != nil {
		t.Fatal(err)
	}
	items := []skills.Skill{
		{Name: "taso", EffectCap: []skills.Effect{skills.EffRead}, Contract: []skills.Check{{Predicate: "gereeni"}}, Steps: []skills.Step{{Name: "alile", Primitive: "loge"}}},
		{Name: "balile", Contract: []skills.Check{{Predicate: "sigeneda"}}},
	}
	metas := []skills.SkillMeta{
		{Name: "a", Description: "skill a"},
		{Name: "b", Description: "skill b"},
	}
	s, err := RunInterop(g, items, metas)
	if err != nil {
		t.Fatal(err)
	}
	if s.Skills != 2 || s.WellFormed != 2 || s.Roundtrips != 2 {
		t.Errorf("interop wrong: %+v (want 2/2/2)", s)
	}
}
