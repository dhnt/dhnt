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

	"gopkg.in/yaml.v3"
)

// Effect is one element of the closed effect lattice (pillar P3). Every
// primitive and predicate declares the effects it may cause; a skill
// declares an upper bound (Skill.EffectCap). The runtime (Layer 3,
// out of scope here) enforces that the effects actually exercised by a
// run are a subset of the declared cap — so an executor cannot exceed
// the declared blast radius.
//
// Each effect has two surface forms: a stable lowercase name used in
// glossary YAML and Effect.String, and a canonical dhnt atom used in
// Layer 1.5 (the dhnt atom is EncodeWord(name), the same rule the
// glossary sync property test enforces).
type Effect int

const (
	EffRead    Effect = iota // read state
	EffWrite                 // mutate state
	EffNet                   // network access
	EffSpend                 // incur cost
	EffDestroy               // irreversible deletion
	EffTime                  // consume wall-clock / sleep
)

var effectName = map[Effect]string{
	EffRead:    "read",
	EffWrite:   "write",
	EffNet:     "net",
	EffSpend:   "spend",
	EffDestroy: "destroy",
	EffTime:    "time",
}

// effectDhnt is the Layer 1.5 atom for each effect. Each value equals
// dhnt.EncodeWord(effectName[e]); it is asserted in effect_test.go so
// the two never drift.
var effectDhnt = map[Effect]string{
	EffRead:    "reada",
	EffWrite:   "wurite",
	EffNet:     "neto",
	EffSpend:   "sopenida",
	EffDestroy: "desotoroyu",
	EffTime:    "time",
}

var (
	effectByName = invert(effectName)
	effectByDhnt = invert(effectDhnt)
)

func invert(m map[Effect]string) map[string]Effect {
	out := make(map[string]Effect, len(m))
	for e, s := range m {
		out[s] = e
	}
	return out
}

// String returns the stable lowercase name (e.g. "read").
func (e Effect) String() string {
	if s, ok := effectName[e]; ok {
		return s
	}
	return fmt.Sprintf("Effect(%d)", int(e))
}

// Dhnt returns the canonical Layer 1.5 atom (e.g. "reada"), or "" if
// the effect is out of range.
func (e Effect) Dhnt() string { return effectDhnt[e] }

// EffectByName resolves a stable lowercase name to an Effect.
func EffectByName(name string) (Effect, bool) { e, ok := effectByName[name]; return e, ok }

// effectByDhntAtom resolves a Layer 1.5 atom to an Effect.
func effectByDhntAtom(atom string) (Effect, bool) { e, ok := effectByDhnt[atom]; return e, ok }

// UnmarshalYAML lets glossary entries declare effects as lowercase
// names: `effects: [read, write]`.
func (e *Effect) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	got, ok := effectByName[s]
	if !ok {
		return fmt.Errorf("skills: unknown effect %q (want one of read/write/net/spend/destroy/time)", s)
	}
	*e = got
	return nil
}

// EffectsWithin reports whether every effect in used is also present in
// cap — i.e. used ⊆ cap. It is the containment check the runtime and
// the attestation (pillar P5) apply against a skill's EffectCap.
func EffectsWithin(used, cap []Effect) bool {
	allowed := make(map[Effect]struct{}, len(cap))
	for _, e := range cap {
		allowed[e] = struct{}{}
	}
	for _, e := range used {
		if _, ok := allowed[e]; !ok {
			return false
		}
	}
	return true
}
