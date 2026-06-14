// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command export_skill prints a dhnt skill as a standard Anthropic
// SKILL.md — the artifact any Skills-capable tool can use with no extra
// config. Drop the output at ~/.claude/skills/<name>/SKILL.md (or a
// project .claude/skills dir) and a generic tool will discover and
// follow it; a dhnt-aware runtime executes the embedded canonical form.
//
//	go run . > SKILL.md
package main

import (
	"fmt"
	"os"

	"github.com/dhnt/dhnt/skills"
)

func main() {
	g, err := skills.SeedGlossary()
	if err != nil {
		panic(err)
	}
	// A small playbook: print, then branch on whether tests are green.
	s := skills.Skill{
		Name:      "rilease",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite},
		Contract:  []skills.Check{{Predicate: "gereeni"}, {Predicate: "sigeneda"}},
		Steps: []skills.Step{
			{Name: "alile", Primitive: "porinito", Args: []skills.Arg{{Name: "value", Value: skills.NewRef("texuto")}}},
			{Branch: &skills.Branch{
				Cond: skills.Check{Predicate: "gereeni"},
				Then: []skills.Step{{Name: "se", Primitive: "loge"}},
				Else: []skills.Step{{Name: "so", Primitive: "porinito"}},
			}},
		},
	}
	md, err := skills.ExportSkillMD(s, g, skills.SkillMeta{
		Name:          "release-pipeline",
		Description:   "Cut a release: announce, then on green tests log success, otherwise report.",
		Phase:         "release",
		UserInvocable: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(md)
}
