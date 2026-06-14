// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command contract_run is an *applied* demonstration: it runs one skill
// through the Layer 3 executor against the real filesystem, via two
// different executor tiers, and shows that the single contract levels
// them — the diligent tier converges, the lazy tier is caught.
//
//	go run .
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dhnt/dhnt/skills"
)

func main() {
	dir, err := os.MkdirTemp("", "dhnt-run-*")
	if err != nil {
		log.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(dir)
	target := filepath.Join(dir, "release.txt")

	// One skill: "leave release.txt present, using only read+write".
	// Contract is the spine; the step is just one possible implementation.
	skill := skills.Skill{
		Name:      "rilease",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite},
		Contract:  []skills.Check{{Predicate: "poresenito"}}, // file present
		Steps:     []skills.Step{{Name: "wurite", Primitive: "wurite"}},
	}

	id, _ := skills.Identity(skill)
	fmt.Printf("skill identity: %s\n\n", id)

	// Real predicate, shared by every tier: does the file exist?
	present := func(args []skills.Arg) (bool, []skills.Effect, error) {
		_, err := os.Stat(target)
		return err == nil, []skills.Effect{skills.EffRead}, nil
	}

	run := func(label string, write skills.PrimitiveFn) {
		_ = os.Remove(target) // reset world-state between tiers
		env := skills.Env{
			Primitives: map[string]skills.PrimitiveFn{"wurite": write},
			Predicates: map[string]skills.PredicateFn{"poresenito": present},
		}
		att, err := skills.Run(skill, env, label)
		if err != nil {
			log.Fatalf("%s: %v", label, err)
		}
		fmt.Printf("tier=%-9s valid=%-5v passed=%v failed=%v effects=%v consistent=%v\n",
			att.Tier, att.Valid, att.Passed, att.Failed, att.Effects, att.Consistent(skill))
	}

	// Tier A — diligent: actually writes the file.
	run("diligent", func(args []skills.Arg) ([]skills.Effect, error) {
		return []skills.Effect{skills.EffWrite}, os.WriteFile(target, []byte("v1\n"), 0o644)
	})

	// Tier B — lazy: claims success but does nothing.
	run("lazy", func(args []skills.Arg) ([]skills.Effect, error) {
		return nil, nil
	})

	// Tier C — rogue: does the work but also deletes (destroy ∉ cap).
	run("rogue", func(args []skills.Arg) ([]skills.Effect, error) {
		if err := os.WriteFile(target, []byte("v1\n"), 0o644); err != nil {
			return nil, err
		}
		return []skills.Effect{skills.EffWrite, skills.EffDestroy}, nil
	})

	fmt.Println("\nOne contract, three executor tiers, one verdict each — convergence is enforced, not hoped for.")
}
