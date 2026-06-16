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

// reuses onFailSkill + boolEnv from policy_test.go (same package).

func TestRunRounds_ValidFirstRound(t *testing.T) {
	env, calls := boolEnv(true)
	res, err := RunRounds(onFailSkill(PolicyAbort), env, "t", 5, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !res.Attestation.Valid || res.Reason != ReasonValid || res.Rounds != 1 {
		t.Errorf("want valid in 1 round, got %+v", res)
	}
	if *calls != 1 {
		t.Errorf("should stop after the first valid round, ran %d", *calls)
	}
}

func TestRunRounds_EventuallyValid(t *testing.T) {
	env, calls := boolEnv(false, false, true)
	res, err := RunRounds(onFailSkill(PolicyAbort), env, "t", 5, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !res.Attestation.Valid || res.Reason != ReasonValid || res.Rounds != 3 {
		t.Errorf("want valid in 3 rounds, got %+v", res)
	}
	if *calls != 3 {
		t.Errorf("ran %d rounds, want 3", *calls)
	}
}

func TestRunRounds_ExhaustsRounds(t *testing.T) {
	env, calls := boolEnv(false) // never valid
	res, err := RunRounds(onFailSkill(PolicyAbort), env, "t", 2, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Attestation.Valid || res.Reason != ReasonMaxRounds || res.Rounds != 2 {
		t.Errorf("want max-rounds stop at 2, got %+v", res)
	}
	if *calls != 2 {
		t.Errorf("ran %d rounds, want the 2-round ceiling", *calls)
	}
}

func TestRunRounds_DefaultsToOneRound(t *testing.T) {
	env, calls := boolEnv(false)
	res, _ := RunRounds(onFailSkill(PolicyAbort), env, "t", 0, nil)
	if res.Rounds != 1 || *calls != 1 {
		t.Errorf("maxRounds<=0 should run exactly once, got rounds=%d calls=%d", res.Rounds, *calls)
	}
}

func TestRunRounds_OverBudgetMidLoop(t *testing.T) {
	env, calls := boolEnv(false) // would never pass on its own
	budgetCalls := 0
	budgetOK := func() (bool, error) {
		budgetCalls++
		return budgetCalls <= 2, nil // allow rounds 1 and 2, refuse the 3rd
	}
	res, err := RunRounds(onFailSkill(PolicyAbort), env, "t", 10, budgetOK)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Reason != ReasonOverBudget || res.Rounds != 2 {
		t.Errorf("want over-budget after 2 rounds, got %+v", res)
	}
	if *calls != 2 {
		t.Errorf("budget should have capped runs at 2, ran %d", *calls)
	}
}

func TestRunRounds_OverBudgetImmediately(t *testing.T) {
	env, calls := boolEnv(true)
	res, err := RunRounds(onFailSkill(PolicyAbort), env, "t", 5, func() (bool, error) { return false, nil })
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Reason != ReasonOverBudget || res.Rounds != 0 || *calls != 0 {
		t.Errorf("want no rounds run when budget refuses up front, got %+v calls=%d", res, *calls)
	}
}

func TestRunRounds_BudgetProbeErrorPropagates(t *testing.T) {
	env, _ := boolEnv(true)
	want := fmt.Errorf("probe broke")
	if _, err := RunRounds(onFailSkill(PolicyAbort), env, "t", 5, func() (bool, error) { return false, want }); err == nil {
		t.Errorf("a budget-probe error must propagate")
	}
}
