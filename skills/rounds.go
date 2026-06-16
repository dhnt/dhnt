// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

// RoundReason explains why a bounded round loop stopped.
type RoundReason string

const (
	ReasonValid      RoundReason = "valid"       // the contract held
	ReasonMaxRounds  RoundReason = "max-rounds"  // exhausted the round budget
	ReasonOverBudget RoundReason = "over-budget" // a budget check refused another round
)

// RoundResult is the outcome of RunRounds: the final attestation, how many
// rounds actually executed, and why the loop stopped.
type RoundResult struct {
	Attestation Attestation
	Rounds      int
	Reason      RoundReason
}

// RunRounds drives a skill in bounded rounds until its contract holds — the
// explicit, terminating form of "goal-oriented until done". It bounds the
// loop two ways: a hard round ceiling (maxRounds; <=0 means 1) and an
// optional spend budget (budgetOK, nil = unbounded). Before each round it
// consults budgetOK; a false reading stops the loop *without* running
// another round (ReasonOverBudget) — the budget is a stop condition, not a
// contract clause, so being over budget is not a validity failure. The loop
// returns as soon as a run is valid (ReasonValid). A hard runtime error
// from Run propagates immediately (it is not a contract failure to retry).
//
// budgetOK keeps the skill honest about spend without faking token
// metering: it delegates measurement to a caller-supplied probe wired to a
// real cost source (e.g. a command comparing accumulated pool spend to a
// ceiling), the same command-gate shape as the goal verifier.
func RunRounds(s Skill, env Env, tier string, maxRounds int, budgetOK func() (bool, error)) (RoundResult, error) {
	if maxRounds <= 0 {
		maxRounds = 1
	}
	var att Attestation
	rounds := 0
	for rounds < maxRounds {
		if budgetOK != nil {
			ok, err := budgetOK()
			if err != nil {
				return RoundResult{Attestation: att, Rounds: rounds, Reason: ReasonOverBudget}, err
			}
			if !ok {
				return RoundResult{Attestation: att, Rounds: rounds, Reason: ReasonOverBudget}, nil
			}
		}
		a, err := Run(s, env, tier)
		rounds++
		if err != nil {
			return RoundResult{Attestation: a, Rounds: rounds, Reason: ReasonMaxRounds}, err
		}
		att = a
		if att.Valid {
			return RoundResult{Attestation: att, Rounds: rounds, Reason: ReasonValid}, nil
		}
	}
	return RoundResult{Attestation: att, Rounds: rounds, Reason: ReasonMaxRounds}, nil
}
