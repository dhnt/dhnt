// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package eval

import (
	"strings"

	"github.com/dhnt/dhnt/skills"
)

// E2 — determinism / repeatability ------------------------------------------

// RepeatSummary reports the stability of a tier across repeated runs.
type RepeatSummary struct {
	Trials           int
	DistinctOutputs  int  // raw-output variance
	DistinctVerdicts int  // verdict variance (1 == perfectly stable)
	AllValid         bool // every trial passed the contract
}

// RunRepeatability runs the same skill under one tier `trials` times. A
// code-leaf (deterministic) tier yields DistinctOutputs=1; a varying-but-
// correct tier yields DistinctOutputs>1 with DistinctVerdicts=1 — showing
// that verification is stable even when the model is not.
func RunRepeatability(tier Tier, prompt, wantsRegex string, trials int) (RepeatSummary, error) {
	outs := map[string]struct{}{}
	verd := map[bool]struct{}{}
	allValid := true
	for i := 0; i < trials; i++ {
		res, err := RunConvergence(prompt, wantsRegex, []Tier{tier})
		if err != nil {
			return RepeatSummary{}, err
		}
		outs[res[0].Output] = struct{}{}
		verd[res[0].Valid] = struct{}{}
		if !res[0].Valid {
			allValid = false
		}
	}
	return RepeatSummary{Trials: trials, DistinctOutputs: len(outs), DistinctVerdicts: len(verd), AllValid: allValid}, nil
}

// E3 — self-healing amortization --------------------------------------------

// HealSummary reports how a skill self-heals across a schedule of contexts.
type HealSummary struct {
	Invocations int
	Repairs     int // model calls (the cost)
	Cached      int // served from a learned version
	Baseline    int
}

// RunSelfHealing replays a schedule of contexts against a skill that fails
// in each new environment, using `repair` as the model. It reports how
// repairs amortize: each distinct context costs one repair; repeats are
// served from the accreted version.
func RunSelfHealing(schedule [][]skills.EnvProbe, repair skills.Completer, g *skills.Glossary) (HealSummary, error) {
	vs := skills.NewMemVersionStore()
	rt := &skills.Runtime{
		Glossary: g, Lang: "en", Versions: vs, Tier: "host",
		Repair: &skills.Repairer{Complete: repair, Glossary: g, Lang: "en", MaxAttempts: 2},
	}
	var sum HealSummary
	for _, probes := range schedule {
		flag := false
		rt.Probes = probes
		_, outcome, err := rt.Run(brokenSkill(), healEnv(&flag))
		if err != nil && outcome != skills.OutcomeFailed {
			return HealSummary{}, err
		}
		sum.Invocations++
		switch outcome {
		case skills.OutcomeRepaired:
			sum.Repairs++
		case skills.OutcomeCached:
			sum.Cached++
		case skills.OutcomeBaseline:
			sum.Baseline++
		}
	}
	return sum, nil
}

// E4 — effect containment ----------------------------------------------------

// ContainSummary reports how the effect cap gates candidate variants.
type ContainSummary struct {
	Candidates   int
	Accepted     int // within cap and contract-valid
	Rejected     int // exceeded the cap (or failed the contract)
	FalseAccepts int // exceeded the cap yet accepted (MUST be 0)
}

// RunContainment runs, for each candidate effect set, a skill capped at
// `cap` whose single step causes those effects and whose contract is
// trivially true. Acceptance must equal "within cap": FalseAccepts==0.
func RunContainment(cap []skills.Effect, candidates [][]skills.Effect) ContainSummary {
	var s ContainSummary
	for _, c := range candidates {
		c := c
		env := skills.Env{
			Primitives: map[string]skills.PrimitiveFn{
				"alile": func(a []skills.Arg) ([]skills.Effect, error) { return c, nil },
			},
			Predicates: map[string]skills.PredicateFn{
				"doniu": func(a []skills.Arg) (bool, []skills.Effect, error) { return true, nil, nil },
			},
		}
		skill := skills.Skill{
			Name: "taso", EffectCap: cap,
			Contract: []skills.Check{{Predicate: "doniu"}},
			Steps:    []skills.Step{{Name: "alile", Primitive: "alile"}},
		}
		att, err := skills.Run(skill, env, "x")
		accepted := err == nil && att.Valid
		s.Candidates++
		if accepted {
			s.Accepted++
			if !skills.EffectsWithin(c, cap) {
				s.FalseAccepts++
			}
		} else {
			s.Rejected++
		}
	}
	return s
}

// E5 — interop conformance ---------------------------------------------------

// InteropSummary reports how many skills export to a well-formed,
// round-trippable SKILL.md (the conformance any Agent-Skills tool needs).
type InteropSummary struct {
	Skills     int
	WellFormed int // frontmatter + contract + canonical block present
	Roundtrips int // embedded canonical form re-parses
}

// RunInterop exports each skill to SKILL.md and checks it is conformant.
func RunInterop(g *skills.Glossary, items []skills.Skill, metas []skills.SkillMeta) (InteropSummary, error) {
	var s InteropSummary
	for i := range items {
		s.Skills++
		md, err := skills.ExportSkillMD(items[i], g, metas[i])
		if err != nil {
			continue
		}
		if strings.Contains(md, "name: ") && strings.Contains(md, "executor: cnl") &&
			strings.Contains(md, "## Success criteria") && strings.Contains(md, "## Constraints") {
			s.WellFormed++
		}
		if canon, ok := extractCanon(md); ok {
			if _, err := skills.ParseDhnt(canon); err == nil {
				s.Roundtrips++
			}
		}
	}
	return s, nil
}

// --- synthetic scenario helpers (shared by E3 and tests) -------------------

var errFail = &failErr{}

type failErr struct{}

func (*failErr) Error() string { return "not supported in this environment" }

func brokenSkill() skills.Skill {
	return skills.Skill{
		Name:      "taso",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite},
		Contract:  []skills.Check{{Predicate: "doniu"}},
		Steps:     []skills.Step{{Name: "alile", Primitive: "borokeni"}},
	}
}

func healEnv(flag *bool) skills.Env {
	return skills.Env{
		Primitives: map[string]skills.PrimitiveFn{
			"borokeni": func(a []skills.Arg) ([]skills.Effect, error) { return nil, errFail },
			"fikiso": func(a []skills.Arg) ([]skills.Effect, error) {
				*flag = true
				return []skills.Effect{skills.EffWrite}, nil
			},
		},
		Predicates: map[string]skills.PredicateFn{
			"doniu": func(a []skills.Arg) (bool, []skills.Effect, error) {
				return *flag, []skills.Effect{skills.EffRead}, nil
			},
		},
	}
}

// extractCanon returns the first fenced canonical block in a SKILL.md.
func extractCanon(md string) (string, bool) {
	i := strings.Index(md, "```\n")
	if i < 0 {
		return "", false
	}
	rest := md[i+4:]
	j := strings.Index(rest, "\n```")
	if j < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:j]), true
}
