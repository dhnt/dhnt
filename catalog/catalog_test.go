// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package catalog

import (
	"strings"
	"testing"
)

// TestLoadCatalog asserts the embedded catalog parses cleanly. This
// is the gate that prevents a malformed skill.md from shipping in a
// release.
func TestLoadCatalog(t *testing.T) {
	if err := LoadError(); err != nil {
		t.Fatalf("catalog failed to load: %v", err)
	}
	if Count() == 0 {
		t.Fatal("catalog is empty — at least the daily-use tier should be present")
	}
}

// TestStructuralInvariants enforces the catalog conventions every
// skill must satisfy. If a contributor adds a skill that violates
// any invariant, this test catches it before tag.
func TestStructuralInvariants(t *testing.T) {
	for _, s := range All() {
		if s.Name == "" {
			t.Errorf("skill %q: empty name", s.Path)
		}
		if !isLowerSlug(s.Name) {
			t.Errorf("skill %q: name %q is not a lowercase slug", s.Path, s.Name)
		}
		if s.Description == "" {
			t.Errorf("skill %q: empty description", s.Path)
		}
		if len(s.Description) > 200 {
			t.Errorf("skill %q: description >200 chars (%d)", s.Path, len(s.Description))
		}
		if !isValidPhase(s.Phase) {
			t.Errorf("skill %q: phase %q invalid", s.Path, s.Phase)
		}
		if !validExecutors[s.Executor] {
			t.Errorf("skill %q: executor %q invalid", s.Path, s.Executor)
		}
		if strings.TrimSpace(s.Body) == "" {
			t.Errorf("skill %q: empty body", s.Path)
		}
		for _, a := range s.Aliases {
			if !isLowerSlug(a) {
				t.Errorf("skill %q: alias %q is not a lowercase slug", s.Path, a)
			}
			if a == s.Name {
				t.Errorf("skill %q: alias duplicates name", s.Path)
			}
		}
	}
}

// TestLookup verifies primary-name and alias lookups both resolve to
// the same skill, and that case-insensitive lookup works.
func TestLookup(t *testing.T) {
	for _, s := range All() {
		got, ok := Lookup(s.Name)
		if !ok {
			t.Errorf("Lookup(%q) returned !ok", s.Name)
			continue
		}
		if got.Name != s.Name {
			t.Errorf("Lookup(%q) returned name %q", s.Name, got.Name)
		}
		got2, ok := Lookup(strings.ToUpper(s.Name))
		if !ok || got2.Name != s.Name {
			t.Errorf("Lookup(%q) case-insensitive failed", s.Name)
		}
		for _, a := range s.Aliases {
			got, ok := Lookup(a)
			if !ok || got.Name != s.Name {
				t.Errorf("Lookup(alias %q) for skill %q failed", a, s.Name)
			}
		}
	}
}

// TestByPhase verifies every skill is reachable via its phase, and
// that ByPhase returns a name-sorted slice.
func TestByPhase(t *testing.T) {
	for _, p := range Phases() {
		entries := ByPhase(p)
		if len(entries) == 0 {
			t.Errorf("phase %q listed in Phases() but ByPhase returned empty", p)
		}
		for i := 1; i < len(entries); i++ {
			if entries[i-1].Name >= entries[i].Name {
				t.Errorf("phase %q: ByPhase not sorted at index %d (%q >= %q)",
					p, i, entries[i-1].Name, entries[i].Name)
			}
		}
	}
}

// TestPhasesCanonicalOrder verifies the lifecycle ordering is preserved.
func TestPhasesCanonicalOrder(t *testing.T) {
	got := Phases()
	rank := make(map[string]int, len(validPhases))
	for i, p := range validPhases {
		rank[p] = i
	}
	for i := 1; i < len(got); i++ {
		if rank[got[i-1]] >= rank[got[i]] {
			t.Errorf("phases not in canonical order at index %d: %v", i, got)
		}
	}
}

func isLowerSlug(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	// can't start or end with hyphen
	if s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	return true
}
