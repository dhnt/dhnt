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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Folding turns a learned, context-scoped fix into a branch woven back
// into the skill (Phase 2 of adaptation). The result is a NEW skill
// version — identity is content-addressed, so any change is a new
// version, never an in-place mutation — that handles more environments
// while keeping the original contract and effect cap. Repeated folds
// accrete arms, so the skill self-heals into a more general playbook.
//
// "Fold back into the skill" means: derive this new version and write it
// to a host-local overlay (VersionStore). It does NOT mean mutating the
// upstream catalog — that is a separate, human-reviewed promotion.

// ContextFingerprint is a numeric form of ContextKey, usable as a branch
// argument (which must be a number or a canonical ref, not arbitrary text).
func ContextFingerprint(probes []EnvProbe) uint64 {
	lines := make([]string, len(probes))
	for i, p := range probes {
		lines[i] = p.Name + "=" + p.Value
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return binary.BigEndian.Uint64(sum[:8])
}

// EnvPredicateID is the glossary id of the built-in environment-match
// predicate (dhnt of "context"). A folded env-branch tests it; the
// runtime binds it via EnvPredicate.
const EnvPredicateID = "conitexuto"

// EnvMatchCheck builds the branch condition "current environment matches
// these probes", for an automatic fold keyed on context.
func EnvMatchCheck(probes []EnvProbe) Check {
	return Check{Predicate: EnvPredicateID, Args: []Arg{{Name: "value", Value: NewNumber(ContextFingerprint(probes))}}}
}

// EnvPredicate returns the PredicateFn that evaluates EnvMatchCheck in
// the current environment: true iff the run's fingerprint equals the
// branch's. Bind it into the Env that runs a folded skill.
func EnvPredicate(currentProbes []EnvProbe) PredicateFn {
	cur := ContextFingerprint(currentProbes)
	return func(args []Arg) (bool, []Effect, error) {
		for _, a := range args {
			if a.Name == "value" && a.Value.Kind == ExprNumber {
				return a.Value.Number == cur, []Effect{EffRead}, nil
			}
		}
		return false, []Effect{EffRead}, fmt.Errorf("skills: env predicate missing numeric value")
	}
}

// WithEnvPredicate adds the environment-match predicate (bound to the
// current probes) to env, so a folded env-branch can be executed.
func WithEnvPredicate(env Env, probes []EnvProbe) Env {
	if env.Predicates == nil {
		env.Predicates = map[string]PredicateFn{}
	}
	env.Predicates[EnvPredicateID] = EnvPredicate(probes)
	return env
}

// FoldBranch derives a new skill version in which the original steps are
// guarded by cond: when cond holds, fixedSteps run; otherwise the
// original steps. The Contract and EffectCap are preserved unchanged —
// folding adapts the how, never the what. Folding an already-folded
// skill nests, accreting one arm per environment.
func FoldBranch(original Skill, cond Check, fixedSteps []Step) (Skill, error) {
	folded := Skill{
		Name:      original.Name,
		Caps:      original.Caps,
		EffectCap: original.EffectCap,
		Contract:  original.Contract,
		Steps: []Step{{Branch: &Branch{
			Cond: cond,
			Then: fixedSteps,
			Else: original.Steps,
		}}},
	}
	if _, err := LineariseDhnt(folded); err != nil {
		return Skill{}, err
	}
	return folded, nil
}

// FoldForContext folds fixedSteps in under an automatic environment-match
// condition for probes (the no-named-predicate fallback).
func FoldForContext(original Skill, probes []EnvProbe, fixedSteps []Step) (Skill, error) {
	return FoldBranch(original, EnvMatchCheck(probes), fixedSteps)
}

// DerivedVersion is a folded skill plus its lineage.
type DerivedVersion struct {
	ParentID   string      `json:"parent_id"`
	ID         string      `json:"id"`
	Canonical  string      `json:"canonical"`
	ContextKey string      `json:"context_key"`
	Attest     Attestation `json:"attest"`
}

// VersionStore is a host-local overlay of derived skill versions, keyed
// by the parent (original) skill identity. This is where "fold back into
// the skill" actually writes — the host self-heals; the catalog is
// untouched.
type VersionStore interface {
	Latest(parentID string) (DerivedVersion, bool, error)
	Save(v DerivedVersion) error
}

// MemVersionStore is an in-memory VersionStore.
type MemVersionStore struct {
	mu sync.Mutex
	m  map[string]DerivedVersion
}

func NewMemVersionStore() *MemVersionStore { return &MemVersionStore{m: map[string]DerivedVersion{}} }

func (s *MemVersionStore) Latest(parentID string) (DerivedVersion, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.m[parentID]
	return v, ok, nil
}

func (s *MemVersionStore) Save(v DerivedVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[v.ParentID] = v
	return nil
}

// FileVersionStore persists overlay versions under <Dir>/<parentID>/latest.json.
type FileVersionStore struct{ Dir string }

func (s *FileVersionStore) path(parentID string) string {
	return filepath.Join(s.Dir, parentID, "latest.json")
}

func (s *FileVersionStore) Latest(parentID string) (DerivedVersion, bool, error) {
	data, err := os.ReadFile(s.path(parentID))
	if os.IsNotExist(err) {
		return DerivedVersion{}, false, nil
	}
	if err != nil {
		return DerivedVersion{}, false, err
	}
	var v DerivedVersion
	if err := json.Unmarshal(data, &v); err != nil {
		return DerivedVersion{}, false, err
	}
	return v, true, nil
}

func (s *FileVersionStore) Save(v DerivedVersion) error {
	p := s.path(v.ParentID)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// ResolveLatest returns the host's latest derived version of original if
// one exists (and parses), else original. This is how a host prefers its
// self-healed version transparently.
func ResolveLatest(vs VersionStore, original Skill) (Skill, bool) {
	id, err := Identity(original)
	if err != nil {
		return original, false
	}
	v, ok, err := vs.Latest(id)
	if err != nil || !ok {
		return original, false
	}
	s, err := ParseDhnt(v.Canonical)
	if err != nil {
		return original, false
	}
	return s, true
}
