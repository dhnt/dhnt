// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev_test

import (
	"strings"
	"testing"

	"github.com/dhnt/dhnt"
	"github.com/dhnt/dhnt/catalog"
	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
	"github.com/dhnt/dhnt/skills/tui"
)

// TestSpelling_IdentityStableOverShippedSkills pins the Identity of
// every shipped Skill (captured before ParseDhnt accepted rule-derived
// alternative spellings) and checks the L1.5 roundtrip is byte-identical.
func TestSpelling_IdentityStableOverShippedSkills(t *testing.T) {
	cases := []struct {
		name string
		s    skills.Skill
		id   string
	}{
		{"conductor", dev.ConductorSkill(), "hdff537944f4158711267c9403cf454ff3f809e8faf5a9496a5389f0ba5656b54"},
		{"conductor judge", dev.ConductorJudgeSkill(), "h50529b826ebc75af7413dc94fde5c9fb85b1adc59f4178dd8a3bb7e7592c7091"},
		{"conductor task", dev.ConductorTaskSkill(), "h861d4231b9250ecbc4ba2adac8ce9d58d6e5d457c0bdffca40d09100663a7d3a"},
		{"conductor composed", dev.ConductorComposedSkill(), "h1ff16abcf96af8f7b8bdce1b3d9f483bac5a8897841ebc16445d0e18fe3e046d"},
		{"safe commit", dev.SafeCommitSkill(), "h6a1682e7b2db41232b7f46d9119943b413b4851d18b809ac0d0bd30089e8d1a4"},
		{"pin bump", dev.PinBumpSkill(), "hb1d205048fad87fbb1e3fbefda58f49b11fccf512eb1dde71a3d120457f88586"},
		{"go verify", dev.GoVerifySkill(), "hf30ae6419f45ff4ee3c58779b9cf556b057b60698eb857f67b49ecd6f4b1f398"},
		{"go check", dev.GoCheckSkill(), "h5abc56b2c00d9e25fb487efcb57f7f8eeb518871432617146725a30a68c59796"},
		{"tui drive once", tui.DriveOnceSkill(), "h5afb7a47a2f3d9c62c4ca07cb6197289be4d40a43b3039c25edc06788f5d8841"},
	}
	for _, c := range cases {
		enc, err := skills.LineariseDhnt(c.s)
		if err != nil {
			t.Fatalf("%s: LineariseDhnt: %v", c.name, err)
		}
		if id, _ := skills.Identity(c.s); id != c.id {
			t.Errorf("%s: Identity = %s, want %s", c.name, id, c.id)
		}
		parsed, err := skills.ParseDhnt(enc)
		if err != nil {
			t.Fatalf("%s: ParseDhnt: %v", c.name, err)
		}
		if again, _ := skills.LineariseDhnt(parsed); again != enc {
			t.Errorf("%s: roundtrip not byte-identical\n first  %s\n second %s", c.name, enc, again)
		}
	}
}

// TestSpelling_CatalogNames: every catalog skill slug, written as its
// familiar a-z words, is accepted as a skill name and stored (and
// re-emitted) only in canonical form.
func TestSpelling_CatalogNames(t *testing.T) {
	all := catalog.All()
	if len(all) == 0 {
		t.Fatal("catalog is empty")
	}
	for _, c := range all {
		for _, word := range strings.Split(c.Name, "-") {
			if word == "" || strings.Trim(word, "abcdefghijklmnopqrstuvwxyz") != "" {
				continue
			}
			canon, err := dhnt.EncodeWord(word)
			if err != nil {
				t.Fatalf("%s: EncodeWord(%q): %v", c.Name, word, err)
			}
			s, err := skills.ParseDhnt("skill " + word + " fin")
			if err != nil {
				t.Fatalf("%s: ParseDhnt(skill %s fin): %v", c.Name, word, err)
			}
			if s.Name != canon {
				t.Errorf("%s: name %q stored as %q, want %q", c.Name, word, s.Name, canon)
			}
			want := "sokilili " + canon + " fini"
			if got, _ := skills.LineariseDhnt(s); got != want {
				t.Errorf("%s: Linearise = %q, want %q", c.Name, got, want)
			}
			if a, _ := skills.Identity(s); a != mustIdentity(t, want) {
				t.Errorf("%s: Identity of %q differs from canonical", c.Name, word)
			}
		}
	}
}

func mustIdentity(t *testing.T, canon string) string {
	t.Helper()
	s, err := skills.ParseDhnt(canon)
	if err != nil {
		t.Fatalf("ParseDhnt(%q): %v", canon, err)
	}
	id, err := skills.Identity(s)
	if err != nil {
		t.Fatalf("Identity: %v", err)
	}
	return id
}
