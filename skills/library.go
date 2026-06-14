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

	"github.com/dhnt/dhnt"
)

// Library is a collection of skills that may call one another (pillar
// P6). A step whose Primitive names a skill in the library is a
// sub-call; any other primitive is a leaf (a glossary primitive bound
// at Layer 3). Composition is content-addressed and recursive: skills
// form a directed graph, and the library resolves and audits it
// without needing a runtime.
type Library struct {
	byName map[string]Skill
}

// NewLibrary returns an empty library.
func NewLibrary() *Library { return &Library{byName: make(map[string]Skill)} }

// Add registers a skill under its canonical name. It rejects a
// non-canonical name or a duplicate.
func (l *Library) Add(s Skill) error {
	if !dhnt.IsCanonical(s.Name) {
		return fmt.Errorf("skills: library skill name %q is not canonical dhnt", s.Name)
	}
	if _, dup := l.byName[s.Name]; dup {
		return fmt.Errorf("skills: duplicate skill %q in library", s.Name)
	}
	l.byName[s.Name] = s
	return nil
}

// Lookup returns the skill registered under name, if any.
func (l *Library) Lookup(name string) (Skill, bool) { s, ok := l.byName[name]; return s, ok }

// Len reports how many skills the library holds.
func (l *Library) Len() int { return len(l.byName) }

// isCall reports whether a step primitive resolves to a sub-skill.
func (l *Library) isCall(primitive string) bool { _, ok := l.byName[primitive]; return ok }

// Dependencies returns the direct sub-skill names s calls, in
// first-seen order, deduplicated. Leaf primitives are not included.
func (l *Library) Dependencies(s Skill) []string {
	var deps []string
	seen := make(map[string]struct{})
	l.collectDeps(s.Steps, seen, &deps)
	return deps
}

// collectDeps walks a step list (recursing into branch arms) gathering
// sub-skill calls in first-seen order.
func (l *Library) collectDeps(steps []Step, seen map[string]struct{}, deps *[]string) {
	for i := range steps {
		st := &steps[i]
		if st.Branch != nil {
			l.collectDeps(st.Branch.Then, seen, deps)
			l.collectDeps(st.Branch.Else, seen, deps)
			continue
		}
		p := st.Primitive
		if !l.isCall(p) {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		*deps = append(*deps, p)
	}
}

// Closure returns the transitive set of sub-skill names reachable from
// s (excluding s itself), in deterministic sorted order. It errors on a
// cycle or a dangling call.
func (l *Library) Closure(s Skill) ([]string, error) {
	result := make(map[string]struct{})
	onPath := make(map[string]struct{})
	var visit func(cur Skill) error
	visit = func(cur Skill) error {
		onPath[cur.Name] = struct{}{}
		for _, dep := range l.Dependencies(cur) {
			if _, looping := onPath[dep]; looping {
				return fmt.Errorf("skills: cycle detected at %q", dep)
			}
			child, ok := l.Lookup(dep)
			if !ok {
				return fmt.Errorf("skills: dangling call to %q", dep)
			}
			if _, done := result[dep]; done {
				continue
			}
			result[dep] = struct{}{}
			if err := visit(child); err != nil {
				return err
			}
		}
		delete(onPath, cur.Name)
		return nil
	}
	if err := visit(s); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(result))
	for n := range result {
		out = append(out, n)
	}
	sortStrings(out)
	return out, nil
}

// EffectViolations audits effect containment across composition (pillar
// P3 over P6): for every caller→callee edge reachable from s, the
// callee's EffectCap must be within the caller's — a skill must not call
// something whose declared blast radius exceeds its own. Each violation
// is reported as "caller -> callee". A cycle is returned as an error.
func (l *Library) EffectViolations(s Skill) ([]string, error) {
	if _, err := l.Closure(s); err != nil {
		return nil, err
	}
	var viol []string
	seen := make(map[string]struct{})
	var walk func(cur Skill)
	walk = func(cur Skill) {
		for _, dep := range l.Dependencies(cur) {
			child, _ := l.Lookup(dep)
			if !EffectsWithin(child.EffectCap, cur.EffectCap) {
				viol = append(viol, fmt.Sprintf("%s -> %s", cur.Name, dep))
			}
			if _, ok := seen[dep]; ok {
				continue
			}
			seen[dep] = struct{}{}
			walk(child)
		}
	}
	walk(s)
	sortStrings(viol)
	return viol, nil
}
