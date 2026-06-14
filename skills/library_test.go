// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"reflect"
	"testing"
)

// call is a step that invokes another skill by name.
func call(stepName, skillName string) Step {
	return Step{Name: stepName, Primitive: skillName}
}

func TestLibrary_AddLookup(t *testing.T) {
	l := NewLibrary()
	if err := l.Add(Skill{Name: "alile"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := l.Add(Skill{Name: "alile"}); err == nil {
		t.Errorf("expected duplicate rejection")
	}
	if err := l.Add(Skill{Name: "Bad"}); err == nil {
		t.Errorf("expected non-canonical name rejection")
	}
	if _, ok := l.Lookup("alile"); !ok {
		t.Errorf("Lookup failed for added skill")
	}
	if l.Len() != 1 {
		t.Errorf("Len = %d, want 1", l.Len())
	}
}

func TestLibrary_DependenciesAndClosure(t *testing.T) {
	l := NewLibrary()
	// leaf "loge" is a glossary primitive (not in the library)
	must(t, l.Add(Skill{Name: "dalile", Steps: []Step{{Name: "x", Primitive: "loge"}}}))
	must(t, l.Add(Skill{Name: "kalile", Steps: []Step{call("y", "dalile")}}))
	// balile calls both kalile and dalile (diamond), plus a leaf and a dup call
	parent := Skill{Name: "balile", Steps: []Step{
		call("a", "kalile"),
		call("b", "dalile"),
		{Name: "c", Primitive: "loge"},
		call("d", "kalile"), // dup
	}}
	must(t, l.Add(parent))

	if got := l.Dependencies(parent); !reflect.DeepEqual(got, []string{"kalile", "dalile"}) {
		t.Errorf("Dependencies = %v, want [kalile dalile]", got)
	}
	clo, err := l.Closure(parent)
	if err != nil {
		t.Fatalf("Closure: %v", err)
	}
	if !reflect.DeepEqual(clo, []string{"dalile", "kalile"}) { // sorted
		t.Errorf("Closure = %v, want [dalile kalile]", clo)
	}
}

func TestLibrary_CycleDetected(t *testing.T) {
	l := NewLibrary()
	must(t, l.Add(Skill{Name: "alile", Steps: []Step{call("x", "balile")}}))
	must(t, l.Add(Skill{Name: "balile", Steps: []Step{call("y", "alile")}}))
	a, _ := l.Lookup("alile")
	if _, err := l.Closure(a); err == nil {
		t.Errorf("expected cycle error")
	}
}

func TestLibrary_EffectViolations(t *testing.T) {
	l := NewLibrary()
	// child may read+write; net is outside any tighter caller
	must(t, l.Add(Skill{Name: "dalile", EffectCap: []Effect{EffRead, EffWrite}}))

	// safe parent: cap covers the child
	safe := Skill{Name: "kalile", EffectCap: []Effect{EffRead, EffWrite}, Steps: []Step{call("x", "dalile")}}
	must(t, l.Add(safe))
	if v, err := l.EffectViolations(safe); err != nil || len(v) != 0 {
		t.Errorf("safe parent: violations=%v err=%v, want none", v, err)
	}

	// unsafe parent: cap narrower than the child it calls
	unsafe := Skill{Name: "balile", EffectCap: []Effect{EffRead}, Steps: []Step{call("y", "dalile")}}
	must(t, l.Add(unsafe))
	v, err := l.EffectViolations(unsafe)
	if err != nil {
		t.Fatalf("EffectViolations: %v", err)
	}
	if !reflect.DeepEqual(v, []string{"balile -> dalile"}) {
		t.Errorf("violations = %v, want [balile -> dalile]", v)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
}
