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
	"strings"
	"testing"
)

// onFailSkill is a tiny skill whose single contract check is bound to a
// caller-controlled boolean, so RunPolicy's per-policy behaviour is
// exercised without any real world-state.
func onFailSkill(pol Policy) Skill {
	return Skill{
		Name:     "tasoki",
		Contract: []Check{{Predicate: "gereeni"}},
		OnFail:   pol,
	}
}

// boolEnv binds the `gereeni` predicate to a fixed sequence of verdicts,
// returning the last value once exhausted, and counts calls.
func boolEnv(verdicts ...bool) (Env, *int) {
	calls := 0
	fn := func([]Arg) (bool, []Effect, error) {
		v := verdicts[len(verdicts)-1]
		if calls < len(verdicts) {
			v = verdicts[calls]
		}
		calls++
		return v, nil, nil // pure check, no effects (skill has no cap)
	}
	return Env{Predicates: map[string]PredicateFn{"gereeni": fn}}, &calls
}

func TestOnFail_RoundtripDhnt(t *testing.T) {
	for _, pol := range []Policy{PolicyAbort, PolicyRetry, PolicyBlockers} {
		s := onFailSkill(pol)
		canon, err := LineariseDhnt(s)
		if err != nil {
			t.Fatalf("LineariseDhnt(%v): %v", pol, err)
		}
		got, err := ParseDhnt(canon)
		if err != nil {
			t.Fatalf("ParseDhnt(%q): %v", canon, err)
		}
		if got.OnFail != pol {
			t.Errorf("policy %v: round-trip got OnFail %v (canon=%q)", pol, got.OnFail, canon)
		}
		// PolicyAbort (default) must be omitted from the canonical form so
		// pre-policy skills stay byte-identical.
		if pol == PolicyAbort && strings.Contains(canon, keywordOnFail) {
			t.Errorf("default policy should be omitted, got %q", canon)
		}
		if pol != PolicyAbort && !strings.Contains(canon, keywordOnFail) {
			t.Errorf("policy %v should be emitted, got %q", pol, canon)
		}
	}
}

func TestOnFail_RoundtripLang(t *testing.T) {
	g, err := SeedGlossary()
	if err != nil {
		t.Fatalf("SeedGlossary: %v", err)
	}
	for _, pol := range []Policy{PolicyAbort, PolicyRetry, PolicyBlockers} {
		s := onFailSkill(pol)
		lang, err := LineariseLang(s, g, "en")
		if err != nil {
			t.Fatalf("LineariseLang(%v): %v", pol, err)
		}
		got, err := ParseLang(lang, g, "en")
		if err != nil {
			t.Fatalf("ParseLang(%q): %v", lang, err)
		}
		if got.OnFail != pol {
			t.Errorf("policy %v: lang round-trip got OnFail %v (lang=%q)", pol, got.OnFail, lang)
		}
	}
}

func TestRunPolicy_Abort(t *testing.T) {
	env, calls := boolEnv(false)
	att, outcome, err := RunPolicy(onFailSkill(PolicyAbort), env, "t", 3)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if att.Valid || outcome != OutcomeFailed {
		t.Errorf("abort: want invalid+failed, got valid=%v outcome=%v", att.Valid, outcome)
	}
	if *calls != 1 {
		t.Errorf("abort must run exactly once, ran %d", *calls)
	}
}

func TestRunPolicy_RetryEventuallyPasses(t *testing.T) {
	// fails, fails, then passes — the eventual-consistency case.
	env, calls := boolEnv(false, false, true)
	att, outcome, err := RunPolicy(onFailSkill(PolicyRetry), env, "t", 3)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !att.Valid || outcome != OutcomeBaseline {
		t.Errorf("retry: want valid, got valid=%v outcome=%v", att.Valid, outcome)
	}
	if *calls != 3 {
		t.Errorf("retry should have run 3 times, ran %d", *calls)
	}
}

func TestRunPolicy_RetryBounded(t *testing.T) {
	env, calls := boolEnv(false) // never passes
	_, outcome, err := RunPolicy(onFailSkill(PolicyRetry), env, "t", 2)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if outcome != OutcomeFailed {
		t.Errorf("retry-bounded: want failed, got %v", outcome)
	}
	if *calls != 3 { // 1 initial + 2 retries
		t.Errorf("retry should cap at 1+2 runs, ran %d", *calls)
	}
}

func TestRunPolicy_BlockersGraceful(t *testing.T) {
	env, calls := boolEnv(false)
	att, outcome, err := RunPolicy(onFailSkill(PolicyBlockers), env, "t", 3)
	if err != nil {
		t.Fatalf("blockers must not error on an invalid contract: %v", err)
	}
	if att.Valid || outcome != OutcomeFailed {
		t.Errorf("blockers: want invalid+failed, got valid=%v outcome=%v", att.Valid, outcome)
	}
	if *calls != 1 {
		t.Errorf("blockers must run once (no retry), ran %d", *calls)
	}
}

func TestRunPolicy_BlockersPropagatesHardError(t *testing.T) {
	// A missing predicate binding is a hard error even under blockers.
	env := Env{Predicates: map[string]PredicateFn{}}
	if _, _, err := RunPolicy(onFailSkill(PolicyBlockers), env, "t", 3); err == nil {
		t.Errorf("blockers must still propagate a hard runtime error")
	}
}

// ensure the example-style usage compiles and the atom is stable.
func TestPolicyAtom_Stable(t *testing.T) {
	cases := map[Policy]string{PolicyAbort: policyAbort, PolicyRetry: policyRetry, PolicyBlockers: policyBlockers}
	for pol, want := range cases {
		if got := policyAtom(pol); got != want {
			t.Errorf("policyAtom(%v)=%q want %q", pol, got, want)
		}
	}
	_ = fmt.Sprint(PolicyAbort) // Policy has no String(); just touch it
}
