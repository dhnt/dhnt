// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// Identity returns the content address of a skill (pillar P0): the
// hash of its canonical Layer 1.5 form. Two skills with the same
// canonical form share one identity — the basis for dedup, caching,
// diffing, and a global namespace of procedures. The "h" prefix marks
// the value as a dhnt skill hash.
func Identity(s Skill) (string, error) {
	canon, err := LineariseDhnt(s)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(canon))
	return "h" + hex.EncodeToString(sum[:]), nil
}

// Attestation is the portable receipt of one skill run (pillar P5).
// Anyone holding the Skill and the Attestation can re-check it with
// Consistent — "consistently close" thereby becomes "provably within
// contract." It is pure data: no behaviour, no dependencies beyond the
// effect lattice.
type Attestation struct {
	Skill     string   // Identity(skill) the run targeted
	Tier      string   // executor tier: a model id, "wasm", "builtin", …
	Passed    []string // contract predicate ids that evaluated to bua
	Failed    []string // contract predicate ids that evaluated to bub
	Effects   []Effect // effects actually observed during the run
	WorldDiff string   // opaque, executor-supplied summary of changes
	Valid     bool     // (every contract check bua) && (Effects ⊆ EffectCap)
}

// Attest builds the receipt for a run of s by an executor `tier`.
// results maps each contract predicate id to the boolean it evaluated
// to; observed is the effect set the run actually exercised. Valid is
// computed, not supplied: a run is valid iff every contract predicate
// is true and the observed effects stay within the skill's EffectCap.
func Attest(s Skill, tier string, results map[string]bool, observed []Effect, worldDiff string) (Attestation, error) {
	id, err := Identity(s)
	if err != nil {
		return Attestation{}, err
	}
	a := Attestation{
		Skill:     id,
		Tier:      tier,
		Effects:   dedupEffects(observed),
		WorldDiff: worldDiff,
	}

	valid := EffectsWithin(observed, s.EffectCap)
	seen := make(map[string]struct{})
	for i := range s.Contract {
		pred := s.Contract[i].Predicate
		if _, dup := seen[pred]; dup {
			continue
		}
		seen[pred] = struct{}{}
		if results[pred] {
			a.Passed = append(a.Passed, pred)
		} else {
			a.Failed = append(a.Failed, pred)
			valid = false
		}
	}
	sort.Strings(a.Passed)
	sort.Strings(a.Failed)
	a.Valid = valid
	return a, nil
}

// Consistent re-checks an attestation against the skill it claims to
// attest, recomputing validity from the receipt's own evidence: the
// recorded Failed set must be empty, every contract predicate must be
// accounted for, the observed effects must stay within the cap, and the
// claimed Skill identity must match. It detects tampering with Valid as
// well as receipts that don't cover the contract.
func (a Attestation) Consistent(s Skill) bool {
	if id, err := Identity(s); err != nil || id != a.Skill {
		return false
	}
	passed := make(map[string]struct{}, len(a.Passed))
	for _, p := range a.Passed {
		passed[p] = struct{}{}
	}
	allPass := len(a.Failed) == 0
	for i := range s.Contract {
		if _, ok := passed[s.Contract[i].Predicate]; !ok {
			allPass = false
		}
	}
	want := allPass && EffectsWithin(a.Effects, s.EffectCap)
	return a.Valid == want
}

func dedupEffects(in []Effect) []Effect {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[Effect]struct{}, len(in))
	var out []Effect
	for _, e := range in {
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
