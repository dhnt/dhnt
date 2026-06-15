// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package eval provides reproducible harnesses for the dhnt paper.
//
// RunConvergence is the E1 experiment: run ONE skill under several model
// tiers and report, per tier, the raw model output and the contract
// verdict. The skill, contract, and effect cap are identical across
// tiers; only the model differs. This isolates the paper's central claim:
// the *verdict* (did it meet spec) is a deterministic, provider-
// independent function of correctness, whereas the *output* varies — and a
// prose skill, having no verdict, would silently accept every tier.
package eval

import (
	"regexp"

	"github.com/dhnt/dhnt/skills"
)

// Tier is one executor under test: a labeled text model.
type Tier struct {
	Name  string
	Model skills.Completer
}

// Result is one tier's trial.
type Result struct {
	Tier   string
	Output string
	Valid  bool // the contract verdict (machine-decided, provider-independent)
	Err    string
}

// answerSkill: a model-leaf step (LatJudge) produces an answer that the
// contract predicate must accept; the effect cap allows read/net/time.
func answerSkill() skills.Skill {
	return skills.Skill{
		Name:      "asiko",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffNet, skills.EffTime},
		Contract:  []skills.Check{{Predicate: "veri"}},
		Steps:     []skills.Step{{Name: "sa", Primitive: "asoki", Latitude: skills.LatJudge}},
	}
}

// RunConvergence runs the same answer-producing skill under each tier.
// wantsRegex defines correctness (the contract predicate matches it
// against the model's output).
func RunConvergence(prompt, wantsRegex string, tiers []Tier) ([]Result, error) {
	re, err := regexp.Compile(wantsRegex)
	if err != nil {
		return nil, err
	}
	skill := answerSkill()
	results := make([]Result, 0, len(tiers))
	for _, t := range tiers {
		t := t
		var answer string
		env := skills.Env{
			Primitives: map[string]skills.PrimitiveFn{
				"asoki": func(args []skills.Arg) ([]skills.Effect, error) {
					o, err := t.Model(prompt)
					answer = o
					if err != nil {
						return nil, err
					}
					return []skills.Effect{skills.EffNet, skills.EffTime}, nil
				},
			},
			Predicates: map[string]skills.PredicateFn{
				"veri": func(args []skills.Arg) (bool, []skills.Effect, error) {
					return re.MatchString(answer), []skills.Effect{skills.EffRead}, nil
				},
			},
		}
		att, err := skills.Run(skill, env, t.Name)
		r := Result{Tier: t.Name, Output: answer}
		if err != nil {
			r.Err = err.Error()
		} else {
			r.Valid = att.Valid
		}
		results = append(results, r)
	}
	return results, nil
}

// Summary aggregates the convergence story for reporting.
type Summary struct {
	Tiers           int
	DistinctOutputs int // raw-output variance across tiers
	WithContract    int // tiers the contract accepts (tracks correctness)
	// WithoutContract is the prose-skill baseline: every tier "succeeds"
	// because there is no machine verdict to catch a wrong answer.
	WithoutContract int
}

// Summarize computes the contrast between contract-gated and prose
// (no-verdict) acceptance.
func Summarize(rs []Result) Summary {
	set := map[string]struct{}{}
	s := Summary{Tiers: len(rs), WithoutContract: len(rs)}
	for _, r := range rs {
		set[r.Output] = struct{}{}
		if r.Valid {
			s.WithContract++
		}
	}
	s.DistinctOutputs = len(set)
	return s
}
