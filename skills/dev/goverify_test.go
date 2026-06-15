// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/dev"
)

// fmtCleanSkill is a minimal skill that only asserts go-verify's fmt-clean
// check, so we can test the fmt scoping without needing a buildable module.
func fmtCleanSkill() skills.Skill {
	return skills.Skill{
		Name:      "fomato",
		EffectCap: []skills.Effect{skills.EffRead},
		Contract:  []skills.Check{{Predicate: dev.PredEmptyOut, Args: []skills.Arg{{Name: "value", Value: skills.NewRef("fi")}}}},
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fmtClean(t *testing.T, dir string) bool {
	t.Helper()
	env, _, err := dev.NewEnv(dev.GoVerifySpec(dir))
	if err != nil {
		t.Fatal(err)
	}
	att, err := skills.Run(fmtCleanSkill(), env, "fmt")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return att.Valid
}

const unformatted = "package x\nfunc  F(){}\n" // double space → gofmt would rewrite

// TestFmtScope_ExcludesVendoredTrees: an unformatted file under vendor/ or
// priorart/ must NOT fail fmt-clean (we don't own it); an unformatted file
// in the repo's own tree MUST.
func TestFmtScope_ExcludesVendoredTrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh/find/gofmt")
	}
	dir := t.TempDir()
	write(t, filepath.Join(dir, "good.go"), "package good\n")
	write(t, filepath.Join(dir, "vendor", "dep", "bad.go"), unformatted)
	write(t, filepath.Join(dir, "priorart", "u", "v", "vendor", "bad.go"), unformatted)
	if !fmtClean(t, dir) {
		t.Errorf("vendored/priorart unformatted files should be excluded from fmt-clean")
	}

	// now an unformatted file in our own tree must be caught
	write(t, filepath.Join(dir, "pkg", "mine.go"), unformatted)
	if fmtClean(t, dir) {
		t.Errorf("an unformatted file in the repo's own tree must fail fmt-clean")
	}
}
