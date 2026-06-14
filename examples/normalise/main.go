// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command normalise demonstrates the L0→L2 step: turn free prose into a
// validated dhnt skill, using a real agent CLI (default: gemini) as the
// model. The model emits the Layer-1 CNL constrained to the glossary;
// ParseLang is the validity gate; the result is a typed Skill AST.
//
//	go run . [agent] "<procedure in plain English>"
//
// e.g. (gemini needs a trusted workspace):
//
//	GEMINI_CLI_TRUST_WORKSPACE=true go run . gemini \
//	  "ensure the tests are green, then log the number 42"
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

func main() {
	agent := "gemini"
	prose := "ensure the tests are green, then log the number 42"
	args := os.Args[1:]
	if len(args) >= 1 {
		if _, ok := tui.Agents[args[0]]; ok {
			agent = args[0]
			args = args[1:]
		}
	}
	if len(args) >= 1 {
		prose = args[0]
	}

	g, err := skills.SeedGlossary()
	if err != nil {
		panic(err)
	}
	complete, ok := tui.Completer(agent, "", 120*time.Second)
	if !ok {
		fmt.Printf("unknown agent %q\n", agent)
		os.Exit(2)
	}

	fmt.Printf("L0 prose: %q\n", prose)
	fmt.Printf("model: %s (headless)\n\n", agent)

	skill, cnl, err := skills.Normalise(prose, g, "en", complete, 2)
	if err != nil {
		fmt.Printf("normalise failed: %v\n", err)
		os.Exit(1)
	}

	dh, _ := skills.LineariseDhnt(skill)
	en, _ := skills.LineariseLang(skill, g, "en")
	id, _ := skills.Identity(skill)

	fmt.Printf("L1 CNL (model output): %s\n\n", cnl)
	fmt.Printf("L1.5 canonical:        %s\n", dh)
	fmt.Printf("L1 (en projection):    %s\n", en)
	fmt.Printf("identity:              %s\n", id)
	fmt.Printf("AST: name=%q caps=%v effects=%v contract=%d steps=%d\n",
		skill.Name, skill.Caps, skill.EffectCap, len(skill.Contract), len(skill.Steps))
}
