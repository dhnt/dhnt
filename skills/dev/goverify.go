// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev

import (
	"time"

	"github.com/dhnt/dhnt/skills"
)

// Command ids for the go-verify skill (canonical dhnt; argv lives in the
// Spec so the test command can vary per repo and be learned as a branch).
const (
	cmdFmtWrite = "fa" // gofmt -w .   (the mutating step)
	cmdFmtList  = "fi" // gofmt -l .   (clean iff empty output)
	cmdVet      = "ve" // go vet ./...
	cmdTest     = "te" // the test command (repo-specific; self-healed)
)

// GoVerifySkill is the flagship dogfood skill: "leave the tree formatted,
// vetted, and green." It formats in place (a step) and contracts on
// fmt-clean ∧ vet-clean ∧ tests-green. The effect cap is {read, write,
// time}. The *test* command is abstract (cmdTest); GoVerifySpec supplies a
// default, and a repo where the default fails (e.g. cloudbox on darwin
// needs `make test`) is repaired and folded as an env-guarded branch.
func GoVerifySkill() skills.Skill {
	ref := func(k string) []skills.Arg {
		return []skills.Arg{{Name: "value", Value: skills.NewRef(k)}}
	}
	return skills.Skill{
		Name:      "goveri",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: PrimRun, Args: ref(cmdFmtWrite)}, // gofmt -w .
		},
		Contract: []skills.Check{
			{Predicate: PredEmptyOut, Args: ref(cmdFmtList)}, // fmt-clean
			{Predicate: PredExit, Args: ref(cmdVet)},         // vet-clean
			{Predicate: PredExit, Args: ref(cmdTest)},        // tests-green
		},
	}
}

// GoCheckSkill is the read-only sibling of go-verify: it omits the
// `gofmt -w` step (no mutation) but keeps the same contract — fmt-clean ∧
// vet-clean ∧ tests-green — and a read-only effect cap. Use it to verify a
// repo you do not want to modify (CI, or another project's working tree).
func GoCheckSkill() skills.Skill {
	ref := func(k string) []skills.Arg {
		return []skills.Arg{{Name: "value", Value: skills.NewRef(k)}}
	}
	return skills.Skill{
		Name:      "gocaki",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffTime},
		Contract: []skills.Check{
			{Predicate: PredEmptyOut, Args: ref(cmdFmtList)}, // fmt-clean
			{Predicate: PredExit, Args: ref(cmdVet)},         // vet-clean
			{Predicate: PredExit, Args: ref(cmdTest)},        // tests-green
		},
	}
}

// fmtFind lists the repo's own Go files, excluding vendored / upstream /
// generated trees that the project does not own or format (gofmt has no
// exclude flag; this is the portable equivalent). Patterns match at any
// depth (e.g. priorart/x/vendor).
const fmtFind = `find . -name '*.go' -not -path '*/vendor/*' -not -path '*/priorart/*' -not -path '*/external/*' -not -path '*/.git/*'`

// gofmtArgv runs `gofmt <flag>` over fmtFind's files; if there are none it
// exits 0 (avoids gofmt reading stdin and hanging).
func gofmtArgv(flag string) []string {
	return []string{"sh", "-c", `fs=$(` + fmtFind + `); [ -z "$fs" ] && exit 0; exec gofmt ` + flag + ` $fs`}
}

// GoVerifySpec returns the command table for a Go module at dir. testArgv
// overrides the test command (default: `go test -short ./...`); this is the
// knob a per-repo branch flips (e.g. to `make test`). The fmt commands
// exclude vendor/priorart/external so a repo with vendored trees isn't
// flagged for code it doesn't own.
func GoVerifySpec(dir string, testArgv ...string) Spec {
	if len(testArgv) == 0 {
		testArgv = []string{"go", "test", "-short", "./..."}
	}
	return Spec{
		Dir:     dir,
		Timeout: 10 * time.Minute,
		Commands: map[string]Command{
			cmdFmtWrite: {Argv: gofmtArgv("-w"), Effects: []skills.Effect{skills.EffWrite}},
			cmdFmtList:  {Argv: gofmtArgv("-l"), Effects: []skills.Effect{skills.EffRead}},
			cmdVet:      {Argv: []string{"go", "vet", "./..."}, Effects: []skills.Effect{skills.EffRead, skills.EffTime}},
			cmdTest:     {Argv: testArgv, Effects: []skills.Effect{skills.EffRead, skills.EffTime}},
		},
	}
}
