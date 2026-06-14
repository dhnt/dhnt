// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportSkillMD(t *testing.T) {
	g := loadSeedGlossary(t)
	s := Skill{
		Name:      "rilease",
		EffectCap: []Effect{EffRead, EffWrite},
		Contract:  []Check{{Predicate: "gereeni"}, {Predicate: "sigeneda"}},
		Steps: []Step{
			{Name: "alile", Primitive: "porinito", Args: []Arg{{Name: "value", Value: NewRef("texuto")}}},
			{Branch: &Branch{
				Cond: Check{Predicate: "gereeni"},
				Then: []Step{{Name: "se", Primitive: "loge"}},
				Else: []Step{{Name: "so", Primitive: "porinito"}},
			}},
		},
	}
	md, err := ExportSkillMD(s, g, SkillMeta{
		Name:          "release-pipeline",
		Description:   "Cut a release: print, branch on tests, log.",
		Phase:         "release",
		UserInvocable: true,
	})
	if err != nil {
		t.Fatalf("ExportSkillMD: %v", err)
	}

	// frontmatter a generic Skills tool needs
	for _, want := range []string{
		"---", "name: release-pipeline", "description: Cut a release",
		"executor: cnl", "user_invocable: true", "phase: release",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("frontmatter missing %q\n%s", want, md)
		}
	}
	// readable body rendered from the AST (English labels via glossary)
	for _, want := range []string{
		"## Constraints", "may only: **read, write**", "must not: net, spend, destroy, time",
		"## Steps", "Run **print**", "If **green", "Otherwise", "Run **log**",
		"## Success criteria", "**green**", "**signed**",
		"## Canonical form", "sokilili rilease", "identity: `h",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("body missing %q\n%s", want, md)
		}
	}

	// the embedded canonical form must re-parse to the same skill
	canonStart := strings.Index(md, "```\n") + 4
	canonEnd := strings.Index(md[canonStart:], "\n```")
	canon := md[canonStart : canonStart+canonEnd]
	parsed, err := ParseDhnt(canon)
	if err != nil {
		t.Fatalf("embedded canonical form does not parse: %v\n%q", err, canon)
	}
	if !reflect_DeepEqualSkill(parsed, s) {
		t.Errorf("embedded canonical form != source skill")
	}
}

// reflect_DeepEqualSkill compares via re-linearisation (slice nil/empty
// differences aside).
func reflect_DeepEqualSkill(a, b Skill) bool {
	la, ea := LineariseDhnt(a)
	lb, eb := LineariseDhnt(b)
	return ea == nil && eb == nil && la == lb
}

func TestWriteBundle(t *testing.T) {
	g := loadSeedGlossary(t)
	s := Skill{Name: "rilease", Contract: []Check{{Predicate: "gereeni"}}, Steps: []Step{{Name: "alile", Primitive: "loge"}}}
	dir := t.TempDir()
	if err := WriteBundle(dir, s, g, SkillMeta{Name: "release", Description: "Cut a release"}); err != nil {
		t.Fatalf("WriteBundle: %v", err)
	}
	md, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil || !strings.Contains(string(md), "name: release") {
		t.Fatalf("SKILL.md bad: %v\n%s", err, md)
	}
	canon, err := os.ReadFile(filepath.Join(dir, "skill.dhnt"))
	if err != nil {
		t.Fatalf("skill.dhnt: %v", err)
	}
	parsed, err := ParseDhnt(strings.TrimSpace(string(canon)))
	if err != nil {
		t.Fatalf("skill.dhnt does not parse: %v", err)
	}
	if !reflect_DeepEqualSkill(parsed, s) {
		t.Errorf("skill.dhnt != source skill")
	}
}

func TestExportSkillMD_RequiresNameAndDescription(t *testing.T) {
	g := loadSeedGlossary(t)
	s := Skill{Name: "rilease"}
	if _, err := ExportSkillMD(s, g, SkillMeta{Description: "x"}); err == nil {
		t.Errorf("expected error for missing name")
	}
	if _, err := ExportSkillMD(s, g, SkillMeta{Name: "x"}); err == nil {
		t.Errorf("expected error for missing description")
	}
}
