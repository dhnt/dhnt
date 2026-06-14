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
