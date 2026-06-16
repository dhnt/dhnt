// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command conductor demonstrates the goal-oriented orchestrator dhnt
// skill. By default it runs hermetically against stubbed phases (POSIX
// true) so you can see a valid+consistent attestation without a real
// fleet; it also prints the canonical form and the real `ycode weave`
// orchestration surface the Spec binds to.
//
//	go run .                       # hermetic stub run (valid=true)
//	go run . --real "ship X"       # real run: drive `ycode weave` for the goal
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
)

func main() {
	skill := dev.ConductorSkill()
	id, _ := skills.Identity(skill)
	canon, _ := skills.LineariseDhnt(skill)
	fmt.Printf("conductor (%s…)\n", id[:14])
	fmt.Printf("canonical: %s\n\n", canon)

	real := len(os.Args) > 1 && os.Args[1] == "--real"
	var spec dev.Spec
	tier := "conductor-stub"
	if real {
		goal := strings.TrimSpace(strings.Join(os.Args[2:], " "))
		if goal == "" {
			fmt.Fprintln(os.Stderr, "usage: go run . --real \"<goal>\"")
			os.Exit(2)
		}
		spec = dev.ConductorSpec(".", goal, []string{"claude"}, nil, 3)
		tier = "conductor"
		fmt.Printf("driving the fleet for goal: %q\n", goal)
	} else {
		// Hermetic: stub every phase + the goal/converged gates with `true`,
		// and keep the goal simple (RESEARCH branch skipped).
		t := dev.Command{Argv: []string{"true"}, Effects: []skills.Effect{skills.EffRead}}
		f := dev.Command{Argv: []string{"false"}, Effects: []skills.Effect{skills.EffRead}}
		spec = dev.Spec{Commands: map[string]dev.Command{
			"pa": t, "bo": f, "re": t, "fa": t, "wo": t, "vo": t, "ru": t, "go": t, "cu": t,
		}}
		fmt.Println("hermetic stub run (no fleet):")
	}

	env, _, err := dev.NewEnv(spec)
	if err != nil {
		panic(err)
	}
	att, runErr := skills.Run(skill, env, tier)
	if runErr != nil {
		fmt.Printf("run error: %v\n", runErr)
		os.Exit(1)
	}
	fmt.Printf("valid=%v consistent=%v passed=%v failed=%v effects=%v\n",
		att.Valid, att.Consistent(skill), att.Passed, att.Failed, att.Effects)
	if !att.Valid {
		os.Exit(1)
	}
}
