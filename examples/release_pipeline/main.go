// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command release_pipeline demonstrates the full skill-CNL roundtrip:
// build a Skill in Go, linearise it to Layer 1.5 (dhnt canonical
// form, a-z + spaces), parse it back to AST (proving validity), and
// linearise the same AST into human-readable English and Chinese
// using the seed glossary.
//
// The same identity (the AST) projects deterministically into each
// language; English and Chinese readers see different text but
// reason about the same skill.
//
// Run from this directory with:
//
//	go run .
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/dhnt/dhnt/skills"
)

func main() {
	g, err := skills.LoadGlossary(seedGlossaryPath())
	if err != nil {
		log.Fatalf("LoadGlossary: %v", err)
	}

	original := skills.Skill{
		Name: "salutoyu",
		Caps: []string{"core"},
		// EffectCap is the blast radius (P3): this run may read and
		// write, but may NOT spend, destroy, or touch the network —
		// enforceable no matter which executor tier runs the steps.
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite},
		// Contract is the spine (P1): a run is valid iff both checks
		// hold, regardless of which executor tier ran the steps.
		Contract: []skills.Check{
			{Predicate: "gereeni"},  // tests-green
			{Predicate: "sigeneda"}, // tag-signed
		},
		Steps: []skills.Step{
			{
				Name:      "feritisitu",
				Primitive: "porinito",
				Args: []skills.Arg{
					{Name: "value", Value: skills.NewRef("texuto")},
				},
			},
			{
				Name:      "secunido",
				Primitive: "loge",
				Args: []skills.Arg{
					{Name: "numibero", Value: skills.NewNumber(2018)},
				},
			},
		},
	}

	dh, err := skills.LineariseDhnt(original)
	if err != nil {
		log.Fatalf("LineariseDhnt: %v", err)
	}
	fmt.Println("Layer 1.5 (dhnt canonical machine form, a-z only):")
	fmt.Println("  " + dh)

	parsed, err := skills.ParseDhnt(dh)
	if err != nil {
		log.Fatalf("ParseDhnt: %v", err)
	}
	if !reflect.DeepEqual(parsed, original) {
		log.Fatalf("AST roundtrip mismatch")
	}
	fmt.Println("\nRoundtrip OK: parse(linearise(skill)) == skill")

	en, err := skills.LineariseLang(original, g, "en")
	if err != nil {
		log.Fatalf("LineariseLang(en): %v", err)
	}
	fmt.Println("\nLayer 1 (English projection):")
	fmt.Println("  " + en)

	zh, err := skills.LineariseLang(original, g, "zh")
	if err != nil {
		log.Fatalf("LineariseLang(zh): %v", err)
	}
	fmt.Println("\nLayer 1 (Chinese projection):")
	fmt.Println("  " + zh)

	// P0 identity + P5 attestation: a run by any executor tier emits a
	// portable receipt; Valid is computed from the contract (P1) and the
	// effect cap (P3), not asserted by the executor.
	id, err := skills.Identity(original)
	if err != nil {
		log.Fatalf("Identity: %v", err)
	}
	fmt.Println("\nIdentity (content address):")
	fmt.Println("  " + id)

	att, err := skills.Attest(original, "claude",
		map[string]bool{"gereeni": true, "sigeneda": true},
		[]skills.Effect{skills.EffRead, skills.EffWrite}, "tagged v1.2.3")
	if err != nil {
		log.Fatalf("Attest: %v", err)
	}
	fmt.Printf("\nAttestation: tier=%s valid=%v passed=%v failed=%v effects=%v\n",
		att.Tier, att.Valid, att.Passed, att.Failed, att.Effects)
	fmt.Printf("  re-checkable by anyone holding the skill: Consistent=%v\n", att.Consistent(original))

	// A forged verdict is caught on re-check.
	forged := att
	forged.Valid = true
	forged.Failed = []string{"sigeneda"}
	fmt.Printf("  forged receipt (Valid flipped) caught: Consistent=%v\n", forged.Consistent(original))
}

// seedGlossaryPath locates the seed glossary YAML relative to this
// example's source file. Works whether the example is run via
// `go run .` from the source dir or via `go run ./examples/...` from
// the module root.
func seedGlossaryPath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	exampleDir := filepath.Dir(thisFile)
	return filepath.Join(exampleDir, "..", "..", "skills", "testdata", "glossary.yaml")
}
