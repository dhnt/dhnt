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
	"strings"
	"testing"

	"github.com/dhnt/dhnt"
)

// sampleSkill is the canonical fixture used across roundtrip tests.
// Its dhnt identifiers all exist in testdata/glossary.yaml.
func sampleSkill() Skill {
	return Skill{
		Name: "salutoyu",
		Caps: []string{"core"},
		Steps: []Step{
			{
				Name:      "feritisitu",
				Primitive: "porinito",
				Args: []Arg{
					{Name: "value", Value: NewRef("texuto")},
				},
			},
			{
				Name:      "secunido",
				Primitive: "loge",
				Args: []Arg{
					{Name: "numibero", Value: NewNumber(2018)},
				},
			},
		},
	}
}

func TestLineariseDhnt_BasicShape(t *testing.T) {
	got, err := LineariseDhnt(sampleSkill())
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	if err := validateLayer15Charset(got); err != nil {
		t.Fatalf("LineariseDhnt produced non-Layer-1.5 output: %v\n%s", err, got)
	}
	for _, w := range strings.Fields(got) {
		if !dhnt.IsCanonical(w) {
			if _, err := dhnt.DecodeDecimal(w); err != nil {
				t.Errorf("output word %q is neither canonical dhnt nor a numeral", w)
			}
		}
	}
}

func TestRoundtrip_SampleSkill(t *testing.T) {
	original := sampleSkill()
	enc, err := LineariseDhnt(original)
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	parsed, err := ParseDhnt(enc)
	if err != nil {
		t.Fatalf("ParseDhnt(%q): %v", enc, err)
	}
	if !reflect.DeepEqual(parsed, original) {
		t.Errorf("roundtrip mismatch:\n original = %#v\n parsed   = %#v\n encoded  = %s",
			original, parsed, enc)
	}
}

func TestRoundtrip_Idempotent(t *testing.T) {
	original := sampleSkill()
	enc1, err := LineariseDhnt(original)
	if err != nil {
		t.Fatalf("LineariseDhnt 1: %v", err)
	}
	parsed, err := ParseDhnt(enc1)
	if err != nil {
		t.Fatalf("ParseDhnt: %v", err)
	}
	enc2, err := LineariseDhnt(parsed)
	if err != nil {
		t.Fatalf("LineariseDhnt 2: %v", err)
	}
	if enc1 != enc2 {
		t.Errorf("non-idempotent linearisation:\n  first  = %s\n  second = %s", enc1, enc2)
	}
}

func TestParseDhnt_RejectsNonAZ(t *testing.T) {
	bad := []string{
		"sokilili Skill fini",  // capital S
		"sokilili skill1 fini", // digit
		"sokilili skill, fini", // punctuation
	}
	for _, s := range bad {
		if _, err := ParseDhnt(s); err == nil {
			t.Errorf("ParseDhnt(%q) accepted non-Layer-1.5 input", s)
		}
	}
}

func TestParseDhnt_RejectsBadStructure(t *testing.T) {
	bad := []string{
		"",
		"sotepo foo bar fini fini",
		"sokilili skill",
		"sokilili rilease needaso fini fini",
		"sokilili rilease bogus fini",
	}
	for _, s := range bad {
		if _, err := ParseDhnt(s); err == nil {
			t.Errorf("ParseDhnt(%q) accepted malformed input", s)
		}
	}
}

func TestLineariseLang_EnglishAndChinese(t *testing.T) {
	g := loadSeedGlossary(t)
	skill := sampleSkill()

	en, err := LineariseLang(skill, g, "en")
	if err != nil {
		t.Fatalf("LineariseLang(en): %v", err)
	}
	zh, err := LineariseLang(skill, g, "zh")
	if err != nil {
		t.Fatalf("LineariseLang(zh): %v", err)
	}

	for _, want := range []string{"skill", "needs", "step", "core", "print", "log", "text", "number"} {
		if !strings.Contains(en, want) {
			t.Errorf("English linearisation missing %q:\n%s", want, en)
		}
	}
	for _, want := range []string{"技能", "需要", "步骤", "core", "打印", "日志", "文本", "数字"} {
		if !strings.Contains(zh, want) {
			t.Errorf("Chinese linearisation missing %q:\n%s", want, zh)
		}
	}
	if !strings.Contains(en, "2018") || !strings.Contains(zh, "2018") {
		t.Errorf("expected decimal numeral 2018 in both linearisations:\n en=%s\n zh=%s", en, zh)
	}
	if en == zh {
		t.Errorf("English and Chinese linearisations are identical: %s", en)
	}
}

// TestRoundtrip_ManyShapes generates a small set of valid skill
// shapes and verifies the roundtrip property holds for every one.
func TestRoundtrip_ManyShapes(t *testing.T) {
	shapes := []Skill{
		{Name: "minili", Caps: nil, Steps: nil},
		{Name: "minili", Caps: []string{"core"}, Steps: nil},
		{
			Name: "minili",
			Caps: []string{"core"},
			Steps: []Step{
				{Name: "alile", Primitive: "porinito", Args: nil},
			},
		},
		{
			Name: "minili",
			Steps: []Step{
				{
					Name: "alile", Primitive: "loge",
					Args: []Arg{
						{Name: "numibero", Value: NewNumber(0)},
						{Name: "value", Value: NewNumber(2018)},
					},
				},
			},
		},
	}
	for i, s := range shapes {
		enc, err := LineariseDhnt(s)
		if err != nil {
			t.Errorf("shape %d: LineariseDhnt: %v", i, err)
			continue
		}
		parsed, err := ParseDhnt(enc)
		if err != nil {
			t.Errorf("shape %d: ParseDhnt(%q): %v", i, enc, err)
			continue
		}
		want := normaliseSkill(s)
		got := normaliseSkill(parsed)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("shape %d: roundtrip mismatch\n want %#v\n  got %#v\n  enc %s",
				i, want, got, enc)
		}
	}
}

func normaliseSkill(s Skill) Skill {
	if len(s.Caps) == 0 {
		s.Caps = nil
	}
	if len(s.Steps) == 0 {
		s.Steps = nil
	}
	for i := range s.Steps {
		if len(s.Steps[i].Args) == 0 {
			s.Steps[i].Args = nil
		}
	}
	return s
}
