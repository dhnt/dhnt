// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package catalog is the curated, lifecycle-organised catalog of
// instruction-skills shipped with the dhnt module. It is sibling to
// (not nested in) the github.com/dhnt/dhnt/skills package: the
// skills package is the typed-AST machinery for programmatic
// skills; this package is the markdown catalog of LLM-instruction
// skills.
//
// Skills are organised by software-development-lifecycle phase
// (discover, plan, build, test, review, commit, integrate, release,
// deploy, operate, maintain, document, onboard). Each skill is a
// markdown file with YAML frontmatter declaring its name,
// description, phase, and execution mechanism. The full catalog is
// embedded into the binary via //go:embed; consumers do not need
// filesystem access at runtime.
//
// Typical use from an agent harness:
//
//	import "github.com/dhnt/dhnt/catalog"
//
//	for _, s := range catalog.ByPhase("review") {
//	    fmt.Println(s.Name, "—", s.Description)
//	}
//
//	if s, ok := catalog.Lookup("commit"); ok {
//	    // s.Body is the markdown instruction the LLM should follow.
//	    // s.Executor is one of: "markdown", "builtin", "cnl".
//	    runSkill(s)
//	}
//
// The frontmatter `executor` field signals to consumers how to
// dispatch the skill:
//
//   - "markdown" (default): hand the body to the LLM as instructions.
//   - "builtin":             consumer looks the name up in its own
//     builtin executor registry (e.g. ycode's
//     internal/runtime/builtin/GetSkillExecutor).
//     The body documents what the executor
//     does; the executor itself is not
//     shipped here.
//   - "cnl":                 consumer dispatches via the dhnt-CNL
//     typed-AST machinery in
//     github.com/dhnt/dhnt/skills. Reserved
//     for the future programmatic-skill
//     layer; consumers may reject this
//     executor today.
//
// Status: alpha. The set of skills, the frontmatter schema, and the
// API may change at minor versions until v1.0.
package catalog
