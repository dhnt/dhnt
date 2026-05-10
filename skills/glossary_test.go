// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dhnt/dhnt"
)

func loadSeedGlossary(t *testing.T) *Glossary {
	t.Helper()
	g, err := LoadGlossary(filepath.Join("testdata", "glossary.yaml"))
	if err != nil {
		t.Fatalf("LoadGlossary: %v", err)
	}
	return g
}

func TestLoadGlossary_SeedFile(t *testing.T) {
	g := loadSeedGlossary(t)
	if g.Len() == 0 {
		t.Fatal("seed glossary loaded as empty")
	}
	required := []string{"sokilili", "sotepo", "needaso", "porinito", "loge", "core"}
	for _, dh := range required {
		if g.LookupDhnt(dh) == nil {
			t.Errorf("seed glossary missing required entry %q", dh)
		}
	}
}

// TestGlossary_DhntFormsMatchEncoder enforces the central glossary
// invariant: every entry's dhnt key must equal the encoder's output
// for the entry's primary English label (or for the all[0] label
// when only an `all` foreign-atom label is set).
func TestGlossary_DhntFormsMatchEncoder(t *testing.T) {
	g := loadSeedGlossary(t)
	for _, e := range g.Entries() {
		enLabels, ok := e.Labels["en"]
		if !ok || len(enLabels) == 0 {
			if all, ok := e.Labels[LangAll]; ok && len(all) > 0 {
				want, err := dhnt.EncodeWord(strings.ToLower(all[0]))
				if err != nil {
					t.Errorf("entry %q: encode all[0]=%q: %v", e.Dhnt, all[0], err)
					continue
				}
				if e.Dhnt != want {
					t.Errorf("entry %q: dhnt does not match EncodeWord(all[0]=%q): got %q, want %q",
						e.Dhnt, all[0], e.Dhnt, want)
				}
				continue
			}
			t.Errorf("entry %q: no en or all label to derive dhnt from", e.Dhnt)
			continue
		}
		want, err := dhnt.EncodeWord(strings.ToLower(enLabels[0]))
		if err != nil {
			t.Errorf("entry %q: EncodeWord(en[0]=%q): %v", e.Dhnt, enLabels[0], err)
			continue
		}
		if e.Dhnt != want {
			t.Errorf("entry %q: dhnt does not match EncodeWord(en[0]=%q): got %q, want %q",
				e.Dhnt, enLabels[0], e.Dhnt, want)
		}
	}
}

func TestGlossary_BidirectionalLookup(t *testing.T) {
	g := loadSeedGlossary(t)
	cases := []struct {
		lang, label, wantDhnt string
	}{
		{"en", "step", "sotepo"},
		{"en", "STEP", "sotepo"},     // case-insensitive
		{"en", "  step  ", "sotepo"}, // trim
		{"en", "action", "sotepo"},   // synonym
		{"zh", "步骤", "sotepo"},       // Chinese label
		{"zh", "buzhou", "sotepo"},   // Pinyin label
		{"all", "core", "core"},      // foreign atom via all
		{"en", "core", "core"},       // foreign atom falls back to all
	}
	for _, tc := range cases {
		e := g.LookupLabel(tc.lang, tc.label)
		if e == nil {
			t.Errorf("LookupLabel(%q, %q) returned nil", tc.lang, tc.label)
			continue
		}
		if e.Dhnt != tc.wantDhnt {
			t.Errorf("LookupLabel(%q, %q) = %q, want %q", tc.lang, tc.label, e.Dhnt, tc.wantDhnt)
		}
	}
}

func TestGlossary_RejectsDuplicateDhnt(t *testing.T) {
	yaml := `entries:
  - dhnt: sotepo
    kind: keyword
    labels: {en: [step]}
  - dhnt: sotepo
    kind: keyword
    labels: {en: [stomp]}
`
	if _, err := ParseGlossary([]byte(yaml)); err == nil {
		t.Error("expected duplicate dhnt key to be rejected")
	}
}

func TestGlossary_RejectsNonCanonicalDhnt(t *testing.T) {
	yaml := `entries:
  - dhnt: GIT
    kind: capability
    labels: {all: [git]}
`
	if _, err := ParseGlossary([]byte(yaml)); err == nil {
		t.Error("expected non-canonical dhnt key to be rejected")
	}
}

func TestEntry_PrimaryLabel(t *testing.T) {
	e := &Entry{
		Dhnt: "sotepo",
		Kind: KindKeyword,
		Labels: map[string][]string{
			"en": {"step", "action"},
			"zh": {"步骤", "buzhou"},
		},
	}
	if got := e.PrimaryLabel("en"); got != "step" {
		t.Errorf("PrimaryLabel(en) = %q, want %q", got, "step")
	}
	if got := e.PrimaryLabel("zh"); got != "步骤" {
		t.Errorf("PrimaryLabel(zh) = %q, want %q", got, "步骤")
	}
	if got := e.PrimaryLabel("xx"); got != "" {
		t.Errorf("PrimaryLabel(xx) = %q, want empty", got)
	}
	e.Labels = map[string][]string{LangAll: {"core"}}
	if got := e.PrimaryLabel("en"); got != "core" {
		t.Errorf("PrimaryLabel(en) with all-only = %q, want %q", got, "core")
	}
}

// TestGlossary_Merge verifies that two non-overlapping glossaries
// combine into one without losing entries; overlapping ones reject.
func TestGlossary_Merge(t *testing.T) {
	a := mustParse(t, `entries:
  - {dhnt: sokilili, kind: keyword, labels: {en: [skill]}}
`)
	b := mustParse(t, `entries:
  - {dhnt: sotepo, kind: keyword, labels: {en: [step]}}
`)
	merged, err := a.Merge(b)
	if err != nil {
		t.Fatalf("Merge non-overlapping: %v", err)
	}
	if merged.Len() != 2 {
		t.Errorf("merged.Len() = %d, want 2", merged.Len())
	}

	c := mustParse(t, `entries:
  - {dhnt: sokilili, kind: keyword, labels: {en: [SKILL]}}
`)
	if _, err := a.Merge(c); err == nil {
		t.Error("expected overlapping merge to be rejected")
	}
}

func mustParse(t *testing.T, yaml string) *Glossary {
	t.Helper()
	g, err := ParseGlossary([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseGlossary: %v", err)
	}
	return g
}
