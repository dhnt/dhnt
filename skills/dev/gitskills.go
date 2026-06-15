// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev

import (
	"fmt"
	"time"

	"github.com/dhnt/dhnt/skills"
)

func ref(k string) []skills.Arg {
	return []skills.Arg{{Name: "value", Value: skills.NewRef(k)}}
}

// --- safe-commit ------------------------------------------------------

// Command ids for safe-commit.
const (
	cmdTestGate = "te"  // tests must pass before committing (reuses go-verify's id)
	cmdAdd      = "ada" // git add <files...> (by name, never -A)
	cmdCommit   = "ko"  // git commit -m <msg>
	cmdNoStaged = "no"  // git diff --cached --quiet (0 == nothing left staged)
)

// SafeCommitSkill encodes the "stage by name, gate, commit" procedure:
// run the tests (fail-fast — no commit on red), stage the named files,
// commit, and contract that the staged set was consumed (the commit
// landed). The named files + message live in the Spec, so the canonical
// skill never contains free text and `git add -A` is structurally
// impossible (only the listed files are staged).
func SafeCommitSkill() skills.Skill {
	return skills.Skill{
		Name:      "safekomito",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: PrimRun, Args: ref(cmdTestGate)}, // tests (fail-fast)
			{Name: "se", Primitive: PrimRun, Args: ref(cmdAdd)},      // git add by name
			{Name: "si", Primitive: PrimRun, Args: ref(cmdCommit)},   // git commit
		},
		Contract: []skills.Check{
			{Predicate: PredExit, Args: ref(cmdNoStaged)}, // commit consumed the staged set
		},
	}
}

// SafeCommitSpec binds safe-commit for a repo at dir: the exact files to
// stage (by name), the commit message, and the test command (default
// `go test -short ./...`).
func SafeCommitSpec(dir, message string, files []string, testArgv ...string) Spec {
	if len(testArgv) == 0 {
		testArgv = []string{"go", "test", "-short", "./..."}
	}
	add := append([]string{"git", "add", "--"}, files...)
	return Spec{
		Dir:     dir,
		Timeout: 10 * time.Minute,
		Commands: map[string]Command{
			cmdTestGate: {Argv: testArgv, Effects: []skills.Effect{skills.EffRead, skills.EffTime}},
			cmdAdd:      {Argv: add, Effects: []skills.Effect{skills.EffWrite}},
			cmdCommit:   {Argv: []string{"git", "commit", "-m", message}, Effects: []skills.Effect{skills.EffWrite}},
			cmdNoStaged: {Argv: []string{"git", "diff", "--cached", "--quiet"}, Effects: []skills.Effect{skills.EffRead}},
		},
	}
}

// --- submodule-pin-bump ----------------------------------------------

// Command ids for submodule-pin-bump.
const (
	cmdGuard   = "ga" // guard: submodule clean AND pushed (exit 0 iff safe)
	cmdAddSub  = "su" // git add <submodule>
	cmdBumpMsg = "bo" // git commit -m "sync: bump <sub> pin"
	cmdSubOK   = "ce" // re-check guard for the contract
)

// PinBumpSkill encodes the umbrella-side of the #1 submodule footgun.
// The FIRST step is a guard that exits non-zero if the submodule has
// uncommitted edits or unpushed commits — so the bump aborts BEFORE
// committing a pin that points at work nobody else can see (the classic
// "edits look lost on a fresh clone" mistake). Then it stages the
// submodule pointer and commits the canonical message, and the contract
// re-asserts the submodule is clean and pushed.
func PinBumpSkill() skills.Skill {
	return skills.Skill{
		Name:      "pinibumo",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: PrimRun, Args: ref(cmdGuard)},   // abort if sub dirty/unpushed
			{Name: "se", Primitive: PrimRun, Args: ref(cmdAddSub)},  // git add <sub>
			{Name: "si", Primitive: PrimRun, Args: ref(cmdBumpMsg)}, // git commit (canonical msg)
		},
		Contract: []skills.Check{
			{Predicate: PredExit, Args: ref(cmdSubOK)}, // submodule clean & pushed
		},
	}
}

// PinBumpSpec binds submodule-pin-bump for umbrella at dir bumping the
// submodule sub. It assumes the submodule's own commit was already pushed
// (the guard verifies it). The guard uses `sh -c` to combine the
// clean+pushed checks; this argv is runtime config, not part of the
// canonical skill.
func PinBumpSpec(dir, sub string) Spec {
	guard := fmt.Sprintf(
		`test -z "$(git -C %q status --porcelain)" && test -z "$(git -C %q log @{u}..HEAD --oneline 2>/dev/null)"`,
		sub, sub)
	return Spec{
		Dir:     dir,
		Timeout: 2 * time.Minute,
		Commands: map[string]Command{
			cmdGuard:   {Argv: []string{"sh", "-c", guard}, Effects: []skills.Effect{skills.EffRead}},
			cmdSubOK:   {Argv: []string{"sh", "-c", guard}, Effects: []skills.Effect{skills.EffRead}},
			cmdAddSub:  {Argv: []string{"git", "add", "--", sub}, Effects: []skills.Effect{skills.EffWrite}},
			cmdBumpMsg: {Argv: []string{"git", "commit", "-m", "sync: bump " + sub + " pin"}, Effects: []skills.Effect{skills.EffWrite}},
		},
	}
}
