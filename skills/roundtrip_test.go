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
	if len(s.EffectCap) == 0 {
		s.EffectCap = nil
	}
	if len(s.Contract) == 0 {
		s.Contract = nil
	}
	if len(s.Steps) == 0 {
		s.Steps = nil
	}
	for i := range s.Contract {
		if len(s.Contract[i].Args) == 0 {
			s.Contract[i].Args = nil
		}
	}
	s.Steps = normaliseSteps(s.Steps)
	return s
}

func normaliseSteps(steps []Step) []Step {
	if len(steps) == 0 {
		return nil
	}
	for i := range steps {
		if steps[i].Branch != nil {
			if len(steps[i].Branch.Cond.Args) == 0 {
				steps[i].Branch.Cond.Args = nil
			}
			steps[i].Branch.Then = normaliseSteps(steps[i].Branch.Then)
			steps[i].Branch.Else = normaliseSteps(steps[i].Branch.Else)
			continue
		}
		if len(steps[i].Args) == 0 {
			steps[i].Args = nil
		}
	}
	return steps
}

// TestRoundtrip_Contract exercises pillar P1: a skill whose spine is a
// contract (with and without steps) must linearise to a-z-clean L1.5,
// re-parse to the identical AST, and project into multiple languages.
func TestRoundtrip_Contract(t *testing.T) {
	shapes := []Skill{
		// contract-only (an SOP: policy, no steps)
		{
			Name:     "rilease",
			Contract: []Check{{Predicate: "gereeni"}, {Predicate: "sigeneda"}},
		},
		// contract + a predicate arg + a step (a runbook)
		{
			Name: "rilease",
			Caps: []string{"core"},
			Contract: []Check{
				{Predicate: "gereeni"},
				{Predicate: "lesoso", Args: []Arg{{Name: "numibero", Value: NewNumber(5)}}},
			},
			Steps: []Step{{Name: "alile", Primitive: "loge"}},
		},
	}
	for i, s := range shapes {
		enc, err := LineariseDhnt(s)
		if err != nil {
			t.Fatalf("shape %d: LineariseDhnt: %v", i, err)
		}
		if err := validateLayer15Charset(enc); err != nil {
			t.Fatalf("shape %d: non-Layer-1.5 output: %v\n%s", i, err, enc)
		}
		if !strings.Contains(enc, keywordEnsure) {
			t.Errorf("shape %d: expected %q in L1.5:\n%s", i, keywordEnsure, enc)
		}
		parsed, err := ParseDhnt(enc)
		if err != nil {
			t.Fatalf("shape %d: ParseDhnt(%q): %v", i, enc, err)
		}
		if !reflect.DeepEqual(normaliseSkill(parsed), normaliseSkill(s)) {
			t.Errorf("shape %d: roundtrip mismatch\n want %#v\n  got %#v\n  enc %s",
				i, normaliseSkill(s), normaliseSkill(parsed), enc)
		}
	}
}

// TestRoundtrip_ContractDefaultOmitted verifies back-compat (invariant
// #4): a skill with no contract linearises byte-identically to the
// pre-contract behaviour — i.e. no stray keyword leaks in.
func TestRoundtrip_ContractDefaultOmitted(t *testing.T) {
	enc, err := LineariseDhnt(sampleSkill())
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	if strings.Contains(enc, keywordEnsure) || strings.Contains(enc, keywordEffect) {
		t.Errorf("contract/effect-free skill leaked a keyword into L1.5:\n%s", enc)
	}
}

// TestRoundtrip_EffectCap exercises pillar P3: an effect-cap block must
// linearise to a-z-clean L1.5 with the effect atoms and re-parse to the
// identical effect set; unknown atoms are rejected.
func TestRoundtrip_EffectCap(t *testing.T) {
	s := Skill{
		Name:      "rilease",
		EffectCap: []Effect{EffRead, EffWrite},
		Contract:  []Check{{Predicate: "gereeni"}},
		Steps:     []Step{{Name: "alile", Primitive: "loge"}},
	}
	enc, err := LineariseDhnt(s)
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	if err := validateLayer15Charset(enc); err != nil {
		t.Fatalf("non-Layer-1.5 output: %v\n%s", err, enc)
	}
	if !strings.Contains(enc, keywordEffect) || !strings.Contains(enc, "reada") || !strings.Contains(enc, "wurite") {
		t.Errorf("expected effect block + atoms in L1.5:\n%s", enc)
	}
	parsed, err := ParseDhnt(enc)
	if err != nil {
		t.Fatalf("ParseDhnt(%q): %v", enc, err)
	}
	if !reflect.DeepEqual(normaliseSkill(parsed), normaliseSkill(s)) {
		t.Errorf("roundtrip mismatch\n want %#v\n  got %#v\n  enc %s",
			normaliseSkill(s), normaliseSkill(parsed), enc)
	}

	// unknown effect atom is rejected
	if _, err := ParseDhnt("sokilili rilease efefecato boguso fini fini"); err == nil {
		t.Errorf("ParseDhnt accepted unknown effect atom")
	}
}

// TestRoundtrip_Latitude exercises pillar P4: a non-default latitude
// round-trips through L1.5 and through every language projection, while
// a default-latitude step stays byte-identical (no `latitude` token).
func TestRoundtrip_Latitude(t *testing.T) {
	g := loadSeedGlossary(t)
	s := Skill{
		Name: "rilease",
		Steps: []Step{
			{Name: "alile", Primitive: "loge", Latitude: LatJudge,
				Args: []Arg{{Name: "numibero", Value: NewNumber(7)}}},
			{Name: "balile", Primitive: "porinito"}, // default LatExact
		},
	}
	enc, err := LineariseDhnt(s)
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	if !strings.Contains(enc, latKeyword+" "+latJudge) {
		t.Errorf("expected latitude marker in L1.5:\n%s", enc)
	}
	// the default-latitude step must not emit a latitude token of its own
	if strings.Count(enc, latKeyword) != 1 {
		t.Errorf("default latitude leaked into L1.5:\n%s", enc)
	}
	parsed, err := ParseDhnt(enc)
	if err != nil {
		t.Fatalf("ParseDhnt: %v", err)
	}
	if !reflect.DeepEqual(normaliseSkill(parsed), normaliseSkill(s)) {
		t.Errorf("L1.5 latitude roundtrip mismatch\n want %#v\n got %#v", s, parsed)
	}
	for _, lang := range []string{"en", "zh"} {
		l1, err := LineariseLang(s, g, lang)
		if err != nil {
			t.Fatalf("LineariseLang(%s): %v", lang, err)
		}
		got, err := ParseLang(l1, g, lang)
		if err != nil {
			t.Fatalf("ParseLang(%s): %v", lang, err)
		}
		if !reflect.DeepEqual(normaliseSkill(got), normaliseSkill(s)) {
			t.Errorf("%s latitude roundtrip mismatch\n want %#v\n got %#v\n l1 %q",
				lang, normaliseSkill(s), normaliseSkill(got), l1)
		}
	}
}

// TestRoundtrip_Branch exercises flow control: a skill with a nested
// when/else branch round-trips through L1.5 and through en/zh.
func TestRoundtrip_Branch(t *testing.T) {
	g := loadSeedGlossary(t)
	s := Skill{
		Name: "darive",
		Steps: []Step{
			{Name: "sa", Primitive: "loge"},
			{Branch: &Branch{
				Cond: Check{Predicate: "gereeni"},
				Then: []Step{
					{Name: "se", Primitive: "porinito", Args: []Arg{{Name: "value", Value: NewRef("texuto")}}},
					{Branch: &Branch{ // nested
						Cond: Check{Predicate: "sigeneda"},
						Then: []Step{{Name: "si", Primitive: "loge"}},
					}},
				},
				Else: []Step{{Name: "so", Primitive: "loge", Args: []Arg{{Name: "numibero", Value: NewNumber(9)}}}},
			}},
		},
		Contract: []Check{{Predicate: "gereeni"}},
	}
	enc, err := LineariseDhnt(s)
	if err != nil {
		t.Fatalf("LineariseDhnt: %v", err)
	}
	if err := validateLayer15Charset(enc); err != nil {
		t.Fatalf("non-Layer-1.5 output: %v\n%s", err, enc)
	}
	if !strings.Contains(enc, keywordWhen) || !strings.Contains(enc, keywordElse) {
		t.Errorf("expected when/else in L1.5:\n%s", enc)
	}
	parsed, err := ParseDhnt(enc)
	if err != nil {
		t.Fatalf("ParseDhnt(%q): %v", enc, err)
	}
	if !reflect.DeepEqual(normaliseSkill(parsed), normaliseSkill(s)) {
		t.Errorf("L1.5 branch roundtrip mismatch\n want %#v\n  got %#v\n  enc %s",
			normaliseSkill(s), normaliseSkill(parsed), enc)
	}
	for _, lang := range []string{"en", "zh"} {
		l1, err := LineariseLang(s, g, lang)
		if err != nil {
			t.Fatalf("LineariseLang(%s): %v", lang, err)
		}
		got, err := ParseLang(l1, g, lang)
		if err != nil {
			t.Fatalf("ParseLang(%s, %q): %v", lang, l1, err)
		}
		if !reflect.DeepEqual(normaliseSkill(got), normaliseSkill(s)) {
			t.Errorf("%s branch roundtrip mismatch\n want %#v\n  got %#v\n  l1 %q",
				lang, normaliseSkill(s), normaliseSkill(got), l1)
		}
	}
}

// TestLineariseLang_Contract checks the contract projects into per-
// language surface forms (pillar P7).
func TestLineariseLang_Contract(t *testing.T) {
	g := loadSeedGlossary(t)
	s := Skill{Name: "rilease", Contract: []Check{{Predicate: "gereeni"}}}
	en, err := LineariseLang(s, g, "en")
	if err != nil {
		t.Fatalf("LineariseLang(en): %v", err)
	}
	zh, err := LineariseLang(s, g, "zh")
	if err != nil {
		t.Fatalf("LineariseLang(zh): %v", err)
	}
	if !strings.Contains(en, "ensure") || !strings.Contains(en, "green") {
		t.Errorf("English contract projection missing terms:\n%s", en)
	}
	if !strings.Contains(zh, "确保") || !strings.Contains(zh, "测试通过") {
		t.Errorf("Chinese contract projection missing terms:\n%s", zh)
	}
}
