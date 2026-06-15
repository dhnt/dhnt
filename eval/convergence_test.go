// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package eval

import "testing"

func model(out string) func(string) (string, error) {
	return func(string) (string, error) { return out, nil }
}

// TestConvergence_VerdictTracksCorrectness is the hermetic E1 demonstration:
// several tiers give DIFFERENT outputs; the prose baseline would accept all;
// the contract accepts only the correct ones; and two correct-but-differently-
// phrased tiers get the SAME verdict — i.e. the verdict converges on
// correctness, independent of phrasing/provider.
func TestConvergence_VerdictTracksCorrectness(t *testing.T) {
	tiers := []Tier{
		{Name: "strong", Model: model("The answer is 42.")},
		{Name: "weak", Model: model("I think it's 7?")},
		{Name: "other-strong", Model: model("42, definitely.")},
	}
	res, err := RunConvergence("what is 6 times 7?", `\b42\b`, tiers)
	if err != nil {
		t.Fatalf("RunConvergence: %v", err)
	}
	byTier := map[string]Result{}
	for _, r := range res {
		byTier[r.Tier] = r
	}
	if !byTier["strong"].Valid {
		t.Errorf("strong should be valid: %+v", byTier["strong"])
	}
	if byTier["weak"].Valid {
		t.Errorf("weak (wrong answer) must be caught by the contract")
	}
	if !byTier["other-strong"].Valid {
		t.Errorf("other-strong should be valid: %+v", byTier["other-strong"])
	}

	s := Summarize(res)
	if s.DistinctOutputs != 3 {
		t.Errorf("expected 3 distinct outputs (raw variance), got %d", s.DistinctOutputs)
	}
	if s.WithoutContract != 3 {
		t.Errorf("prose baseline should accept all 3, got %d", s.WithoutContract)
	}
	if s.WithContract != 2 {
		t.Errorf("contract should accept exactly the 2 correct tiers, got %d", s.WithContract)
	}
}

// TestConvergence_ModelErrorIsCaught: a tier whose model errors does not
// silently pass.
func TestConvergence_ModelErrorIsCaught(t *testing.T) {
	boom := Tier{Name: "down", Model: func(string) (string, error) { return "", errBoom }}
	res, err := RunConvergence("q", "x", []Tier{boom})
	if err != nil {
		t.Fatalf("RunConvergence: %v", err)
	}
	if res[0].Valid || res[0].Err == "" {
		t.Errorf("a failing model tier must be recorded as not-valid with an error: %+v", res[0])
	}
}

var errBoom = &boomErr{}

type boomErr struct{}

func (*boomErr) Error() string { return "model unavailable" }
