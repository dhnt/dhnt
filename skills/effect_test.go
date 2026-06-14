// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"testing"

	"github.com/dhnt/dhnt"
)

// TestEffect_DhntAtomsMatchEncoder mirrors the glossary sync invariant
// for the effect lattice: each effect's Layer 1.5 atom must equal
// dhnt.EncodeWord of its stable lowercase name, so the two never drift.
func TestEffect_DhntAtomsMatchEncoder(t *testing.T) {
	for e, name := range effectName {
		want, err := dhnt.EncodeWord(name)
		if err != nil {
			t.Errorf("effect %q: EncodeWord: %v", name, err)
			continue
		}
		if got := e.Dhnt(); got != want {
			t.Errorf("effect %q: atom %q != EncodeWord(%q)=%q", name, got, name, want)
		}
	}
}

func TestEffectsWithin(t *testing.T) {
	cap := []Effect{EffRead, EffWrite}
	cases := []struct {
		used []Effect
		want bool
	}{
		{nil, true},
		{[]Effect{EffRead}, true},
		{[]Effect{EffRead, EffWrite}, true},
		{[]Effect{EffSpend}, false},
		{[]Effect{EffRead, EffDestroy}, false},
	}
	for i, c := range cases {
		if got := EffectsWithin(c.used, cap); got != c.want {
			t.Errorf("case %d: EffectsWithin(%v, %v) = %v, want %v", i, c.used, cap, got, c.want)
		}
	}
}
