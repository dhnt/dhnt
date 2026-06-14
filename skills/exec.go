// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import "fmt"

// PrimitiveFn implements a leaf step against the real world. It returns
// the effects it actually caused, so the executor can hold the run to
// the skill's EffectCap.
type PrimitiveFn func(args []Arg) ([]Effect, error)

// PredicateFn evaluates a contract check against the real world. It
// returns the truth value (the dhnt boolean) plus the effects the
// check itself caused (e.g. a read).
type PredicateFn func(args []Arg) (bool, []Effect, error)

// Env is the Layer 3 leaf: it binds dhnt primitive/predicate ids to real
// implementations. This is where the spec meets the world — the same AST
// run against two Envs (two executor tiers) is judged by the same
// contract, which is the whole convergence claim.
type Env struct {
	Primitives map[string]PrimitiveFn
	Predicates map[string]PredicateFn
}

// Run executes a skill against an Env on behalf of executor `tier`:
// every step's primitive runs, then every contract predicate is
// evaluated against the resulting world-state, and the run is sealed
// into an Attestation. Valid is computed from the contract (P1) and the
// effect cap (P3) — never asserted by the executor. A missing binding
// is a hard error: an executor cannot silently skip a step or a check.
func Run(s Skill, env Env, tier string) (Attestation, error) {
	var observed []Effect

	if err := runSteps(s.Steps, env, &observed); err != nil {
		return Attestation{}, err
	}

	results := make(map[string]bool, len(s.Contract))
	for i := range s.Contract {
		c := &s.Contract[i]
		fn, ok := env.Predicates[c.Predicate]
		if !ok {
			return Attestation{}, fmt.Errorf("skills: no binding for predicate %q", c.Predicate)
		}
		ok2, effs, err := fn(c.Args)
		if err != nil {
			return Attestation{}, fmt.Errorf("skills: predicate %q: %w", c.Predicate, err)
		}
		observed = append(observed, effs...)
		results[c.Predicate] = ok2
	}

	return Attest(s, tier, results, observed, "")
}

// runSteps executes a step list, recursing into branches. A branch
// evaluates its condition predicate and runs the Then or Else arm.
func runSteps(steps []Step, env Env, observed *[]Effect) error {
	for i := range steps {
		st := &steps[i]
		if st.Branch != nil {
			pred := st.Branch.Cond.Predicate
			fn, ok := env.Predicates[pred]
			if !ok {
				return fmt.Errorf("skills: no binding for branch condition %q", pred)
			}
			took, effs, err := fn(st.Branch.Cond.Args)
			if err != nil {
				return fmt.Errorf("skills: branch %q: %w", pred, err)
			}
			*observed = append(*observed, effs...)
			arm := st.Branch.Else
			if took {
				arm = st.Branch.Then
			}
			if err := runSteps(arm, env, observed); err != nil {
				return err
			}
			continue
		}
		fn, ok := env.Primitives[st.Primitive]
		if !ok {
			return fmt.Errorf("skills: no binding for primitive %q (step %q)", st.Primitive, st.Name)
		}
		effs, err := fn(st.Args)
		if err != nil {
			return fmt.Errorf("skills: step %q (%s): %w", st.Name, st.Primitive, err)
		}
		*observed = append(*observed, effs...)
	}
	return nil
}
