// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command eval_all runs the hermetic dhnt evaluation suite (E1–E5) and
// prints a consolidated report. These are illustrative numbers from
// synthetic tiers; live numbers come from eval_convergence --real and a
// real task suite (see eval/README.md).
//
//	go run .
package main

import (
	"fmt"

	"github.com/dhnt/dhnt/eval"
	"github.com/dhnt/dhnt/skills"
)

func fixed(s string) func(string) (string, error) {
	return func(string) (string, error) { return s, nil }
}

func main() {
	g, _ := skills.SeedGlossary()

	// E1 — cross-model convergence.
	e1, _ := eval.RunConvergence("6*7?", `\b42\b`, []eval.Tier{
		{Name: "strong", Model: fixed("the answer is 42")},
		{Name: "weak", Model: fixed("maybe 7")},
		{Name: "other", Model: fixed("42 it is")},
	})
	s1 := eval.Summarize(e1)

	// E2 — repeatability (a deterministic tier).
	e2, _ := eval.RunRepeatability(eval.Tier{Name: "code", Model: fixed("42")}, "q", `\b42\b`, 5)

	// E3 — self-healing amortization across 2 contexts.
	ctxA := []skills.EnvProbe{{Name: "os", Value: "darwin"}}
	ctxB := []skills.EnvProbe{{Name: "os", Value: "linux"}}
	e3, _ := eval.RunSelfHealing([][]skills.EnvProbe{ctxA, ctxA, ctxB, ctxA, ctxB},
		fixed("<dhnt>skill taso step alile fikiso</dhnt>"), g)

	// E4 — effect containment.
	e4 := eval.RunContainment(
		[]skills.Effect{skills.EffRead, skills.EffWrite},
		[][]skills.Effect{
			{skills.EffRead}, {skills.EffRead, skills.EffWrite},
			{skills.EffRead, skills.EffWrite, skills.EffDestroy}, {skills.EffNet},
		})

	// E5 — interop conformance.
	e5, _ := eval.RunInterop(g,
		[]skills.Skill{
			{Name: "taso", EffectCap: []skills.Effect{skills.EffRead}, Contract: []skills.Check{{Predicate: "gereeni"}}, Steps: []skills.Step{{Name: "alile", Primitive: "loge"}}},
			{Name: "balile", Contract: []skills.Check{{Predicate: "sigeneda"}}},
		},
		[]skills.SkillMeta{{Name: "a", Description: "a"}, {Name: "b", Description: "b"}})

	fmt.Println("dhnt evaluation suite (hermetic, illustrative)")
	fmt.Println("=============================================")
	fmt.Printf("E1 convergence : %d tiers, %d distinct outputs; accepted without-contract=%d with-contract=%d\n",
		s1.Tiers, s1.DistinctOutputs, s1.WithoutContract, s1.WithContract)
	fmt.Printf("E2 determinism : %d trials, %d distinct outputs, %d distinct verdicts (stable=%v)\n",
		e2.Trials, e2.DistinctOutputs, e2.DistinctVerdicts, e2.DistinctVerdicts == 1)
	fmt.Printf("E3 self-heal   : %d invocations, %d repairs, %d cached (amortized model calls = repairs)\n",
		e3.Invocations, e3.Repairs, e3.Cached)
	fmt.Printf("E4 containment : %d candidates, %d accepted, %d rejected, %d false-accepts\n",
		e4.Candidates, e4.Accepted, e4.Rejected, e4.FalseAccepts)
	fmt.Printf("E5 interop     : %d skills, %d well-formed SKILL.md, %d round-trip\n",
		e5.Skills, e5.WellFormed, e5.Roundtrips)
}
