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

// TestParseLang_RoundtripsAllProjections is the core P7 property: for
// every registered language, parsing a skill's Layer 1 projection
// reconstructs the identical AST — so editing any human-language face
// re-normalises to the one identity.
func TestParseLang_RoundtripsAllProjections(t *testing.T) {
	g := loadSeedGlossary(t)
	skills := []Skill{
		sampleSkill(),
		{
			Name:      "rilease",
			Caps:      []string{"core"},
			EffectCap: []Effect{EffRead, EffWrite},
			Contract:  []Check{{Predicate: "gereeni"}, {Predicate: "sigeneda"}},
			Steps:     []Step{{Name: "alile", Primitive: "loge"}},
		},
		// contract-only (an SOP)
		{Name: "rilease", Contract: []Check{{Predicate: "gereeni"}}},
	}
	for si, s := range skills {
		for _, lang := range []string{"en", "zh"} {
			l1, err := LineariseLang(s, g, lang)
			if err != nil {
				t.Fatalf("skill %d %s: LineariseLang: %v", si, lang, err)
			}
			got, err := ParseLang(l1, g, lang)
			if err != nil {
				t.Fatalf("skill %d %s: ParseLang(%q): %v", si, lang, l1, err)
			}
			if !reflect.DeepEqual(normaliseSkill(got), normaliseSkill(s)) {
				t.Errorf("skill %d %s: roundtrip mismatch\n want %#v\n  got %#v\n  L1 %q",
					si, lang, normaliseSkill(s), normaliseSkill(got), l1)
			}
		}
	}
}

// TestParseLang_AgreesWithParseDhnt: the two parsers must converge on
// the same AST — Layer 1 (English) and Layer 1.5 are two faces of one
// identity.
func TestParseLang_AgreesWithParseDhnt(t *testing.T) {
	g := loadSeedGlossary(t)
	s := sampleSkill()

	dh, err := LineariseDhnt(s)
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	fromDhnt, err := ParseDhnt(dh)
	if err != nil {
		t.Fatalf("ParseDhnt: %v", err)
	}

	en, err := LineariseLang(s, g, "en")
	if err != nil {
		t.Fatalf("LineariseLang(en): %v", err)
	}
	fromLang, err := ParseLang(en, g, "en")
	if err != nil {
		t.Fatalf("ParseLang(en): %v", err)
	}

	if !reflect.DeepEqual(normaliseSkill(fromDhnt), normaliseSkill(fromLang)) {
		t.Errorf("ParseDhnt and ParseLang disagree:\n dhnt %#v\n lang %#v",
			normaliseSkill(fromDhnt), normaliseSkill(fromLang))
	}
}

func TestParseLang_NumeralAndForeignAtom(t *testing.T) {
	g := loadSeedGlossary(t)
	// "log number 2018": decimal numeral round-trips; "core" foreign atom
	s := Skill{
		Name: "minili",
		Caps: []string{"core"},
		Steps: []Step{
			{Name: "alile", Primitive: "loge", Args: []Arg{{Name: "numibero", Value: NewNumber(2018)}}},
		},
	}
	en, _ := LineariseLang(s, g, "en")
	got, err := ParseLang(en, g, "en")
	if err != nil {
		t.Fatalf("ParseLang: %v", err)
	}
	if !reflect.DeepEqual(normaliseSkill(got), normaliseSkill(s)) {
		t.Errorf("numeral/atom roundtrip mismatch\n want %#v\n  got %#v\n  en %q",
			normaliseSkill(s), normaliseSkill(got), en)
	}
}

func TestParseLang_RejectsUnknownToken(t *testing.T) {
	g := loadSeedGlossary(t)
	// "Skill" capitalised is not a label and not canonical dhnt.
	if _, err := ParseLang("skill rilease ensure Nonexistent", g, "en"); err == nil {
		t.Errorf("ParseLang accepted an unknown token")
	}
}
