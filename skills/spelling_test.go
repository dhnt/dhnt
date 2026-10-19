// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"strings"
	"testing"
)

// TestParseDhnt_AlternativeSpellings: ParseDhnt reads a token by its
// rule-derived expansion N(token) = dhnt.EncodeWord(token), and the AST
// (hence LineariseDhnt and Identity) only ever holds canonical forms.
func TestParseDhnt_AlternativeSpellings(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"ls as skill name", "sokilili ls fini", "sokilili liso fini"},
		{"liso unchanged", "sokilili liso fini", "sokilili liso fini"},
		{"bash", "sokilili bash fini", "sokilili basohe fini"},
		{"bsh", "sokilili bsh fini", "sokilili basohe fini"},
		{"basohe unchanged", "sokilili basohe fini", "sokilili basohe fini"},
		{"hello", "sokilili hello fini", "sokilili helilo fini"},
		{"hllo", "sokilili hllo fini", "sokilili helilo fini"},
		{"helilo unchanged", "sokilili helilo fini", "sokilili helilo fini"},
		{"aaa unchanged", "sokilili aaa fini", "sokilili aaa fini"},
		{"conventional keywords", "skill ls fin", "sokilili liso fini"},
		{"contracted keywords", "skll ls fin", "sokilili liso fini"},
		{
			"step name, primitive, arg name and ref value",
			"skill ls needs core fin step bash hllo value ls fin fin",
			"sokilili liso needaso core fini sotepo basohe helilo value liso fini fini",
		},
		{
			"numeral value is read as before",
			"skill ls step bash hllo numibero jubajiahe fin fin",
			"sokilili liso sotepo basohe helilo numibero jubajiahe fini fini",
		},
		{
			"effect cap, contract, latitude, branch, on-fail",
			"skill ls effect read write fin ensure exitzero fin " +
				"step bash hllo latitude judge fin " +
				"when exitzero step bsh ls fin else step bsh hllo fin fin " +
				"onfail retry fin fin",
			"sokilili liso efefecato reada wurite fini enisure exitozero fini " +
				"sotepo basohe helilo latitude judage fini " +
				"wuheni exitozero sotepo basohe liso fini elise sotepo basohe helilo fini fini " +
				"onifaili retoroyu fini fini",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := ParseDhnt(c.src)
			if err != nil {
				t.Fatalf("ParseDhnt(%q): %v", c.src, err)
			}
			got, err := LineariseDhnt(s)
			if err != nil {
				t.Fatalf("LineariseDhnt: %v", err)
			}
			if got != c.want {
				t.Fatalf("Linearise(ParseDhnt(%q))\n got  %q\n want %q", c.src, got, c.want)
			}
			canon, err := ParseDhnt(c.want)
			if err != nil {
				t.Fatalf("ParseDhnt(canonical %q): %v", c.want, err)
			}
			idGot, _ := Identity(s)
			idWant, _ := Identity(canon)
			if idGot != idWant {
				t.Fatalf("Identity differs: conventional %s, canonical %s", idGot, idWant)
			}
		})
	}
}

// TestParseDhnt_UnknownTokenShowsExpansion: a token with no
// rule-derived match keeps its error and also names N(token).
func TestParseDhnt_UnknownTokenShowsExpansion(t *testing.T) {
	cases := []struct {
		src  string
		want []string
	}{
		{"sokilili ls bogus fini", []string{`"bogus"`, `"boguso"`, "skill body"}},
		{"sokilili ls efefecato rd fini fini", []string{`"rd"`, `"roda"`, "effect atom"}},
		{"sokilili ls onifaili gvp fini fini", []string{`"gvp"`, `"gevupo"`, "on-fail policy"}},
		{"sokilili ls sotepo a b latitude lz fini fini", []string{`"lz"`, `"lizu"`, "latitude"}},
		{"skl ls fini", []string{`"skl"`, `"sokili"`, `"sokilili"`}},
	}
	for _, c := range cases {
		_, err := ParseDhnt(c.src)
		if err == nil {
			t.Errorf("ParseDhnt(%q) accepted an unknown token", c.src)
			continue
		}
		for _, w := range c.want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("ParseDhnt(%q) error %q does not mention %s", c.src, err, w)
			}
		}
	}
}

// TestResolveSense_Ambiguity: when N(token) matches more than one
// distinct entry in a single grammatical position, resolution fails
// with an error naming every candidate — it never guesses.
func TestResolveSense_Ambiguity(t *testing.T) {
	senses := []sense{
		{label: "bash (shell)", canonical: "basohe"},
		{label: "bash (party)", canonical: "basohe"},
		{label: "list", canonical: "liso"},
	}
	for _, tok := range []string{"bash", "bsh", "basohe"} {
		_, err := resolveSense("test position", tok, senses)
		if err == nil {
			t.Fatalf("resolveSense(%q) guessed among homonyms", tok)
		}
		for _, w := range []string{"ambiguous", "bash (shell)", "bash (party)", `"basohe"`, "test position"} {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("resolveSense(%q) error %q does not mention %q", tok, err, w)
			}
		}
	}
	got, err := resolveSense("test position", "ls", senses)
	if err != nil || got.label != "list" {
		t.Fatalf("resolveSense(ls) = %v, %v; want list", got, err)
	}
	if _, err := resolveSense("test position", "zzz", senses); err == nil ||
		!strings.Contains(err.Error(), `"zuzuzu"`) {
		t.Fatalf("resolveSense(zzz) error = %v; want unknown with expansion", err)
	}
}

// TestParserSenseTables_NoHomonyms: every closed set ParseDhnt resolves
// against is homonym-free today, so no existing input becomes ambiguous.
func TestParserSenseTables_NoHomonyms(t *testing.T) {
	for name, table := range map[string][]sense{
		"skill body": skillBodySenses,
		"effect":     effectSenses(),
		"policy":     policySenses,
		"latitude":   latitudeSenses,
	} {
		seen := map[string]string{}
		for _, s := range table {
			if prev, dup := seen[s.canonical]; dup {
				t.Errorf("%s: %q and %q share canonical form %q", name, prev, s.label, s.canonical)
			}
			seen[s.canonical] = s.label
		}
	}
}

// TestSpelling_IdentityStableOverFixtures pins the Identity of every
// named test skill (captured before alternative-spelling input landed)
// and checks the L1.5 roundtrip is still byte-identical.
func TestSpelling_IdentityStableOverFixtures(t *testing.T) {
	cases := []struct {
		name string
		s    Skill
		id   string
	}{
		{"sample", sampleSkill(), "h293c8e3be6719d1a6fdb74600283dfd2e06b2d8b2392668908549fc8e33de2d1"},
		{"broken", brokenSkill(), "h3eec33ab61c02b8ed866eb3d0945d5cd5bd4849af48dfea18fe38beececb8ea9"},
		{"contracted", contractedSkill(), "h344f0e0b17e404613d719a8ae5ae5f8775664c7dee266a98e43a7a843bdca268"},
		{"onfail abort", onFailSkill(PolicyAbort), "hecd82f18182e288ad8116b59c0c86b474161cf82fc8d6f7555fea929a4615a91"},
		{"onfail retry", onFailSkill(PolicyRetry), "h2251015a943aa9046094be65e0ba906d020988d2cacfaddec6ed582adab892b2"},
		{"onfail blockers", onFailSkill(PolicyBlockers), "h3dea58b131ab033ea7e1facdb3940f56470c8ce648847514c1a486b6832a0c39"},
	}
	for _, c := range cases {
		assertSpellingStable(t, c.name, c.s, c.id)
	}
}

// assertSpellingStable checks that s has the pinned Identity and that
// Linearise(ParseDhnt(Linearise(s))) is byte-identical to Linearise(s).
func assertSpellingStable(t *testing.T, name string, s Skill, wantID string) {
	t.Helper()
	enc, err := LineariseDhnt(s)
	if err != nil {
		t.Fatalf("%s: LineariseDhnt: %v", name, err)
	}
	if id, _ := Identity(s); id != wantID {
		t.Errorf("%s: Identity = %s, want %s", name, id, wantID)
	}
	parsed, err := ParseDhnt(enc)
	if err != nil {
		t.Fatalf("%s: ParseDhnt: %v", name, err)
	}
	again, err := LineariseDhnt(parsed)
	if err != nil {
		t.Fatalf("%s: re-linearise: %v", name, err)
	}
	if again != enc {
		t.Errorf("%s: roundtrip not byte-identical\n first  %s\n second %s", name, enc, again)
	}
}
