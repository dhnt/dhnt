// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package skills implements the multilingual Controlled Natural
// Language for skill specifications, layered on top of the dhnt
// language primitives in github.com/dhnt/dhnt.
//
// The model is a four-layer pipeline:
//
//	[Layer 0]  free prose in any natural language    (author-facing)
//	   ↓ LLM normaliser  (out of scope for this package)
//	[Layer 1]  glossary-locked CNL in source language
//	   ↓ deterministic glossary lookup + dhnt encode
//	[Layer 1.5]  dhnt canonical form  (a-z + spaces only)
//	   ↓ regular parse
//	[Layer 2]  typed AST keyed by dhnt identifiers
//	   ↓ interpret + dispatch  (out of scope for this package)
//	[Layer 3]  Wasm leaves + AST orchestrator
//
// Validity is defined by transpilability: a skill expression is valid
// iff it transpiles cleanly into Layer 1.5. The dhnt encoder (in the
// parent package) is the validator. dhnt is purely machine-facing;
// humans never have to read it, but anyone who chooses to can run
// the transpiler to check.
//
// Phase 0 of this package ships:
//
//   - Glossary  — closed multilingual lexicon with bidirectional
//     label↔dhnt lookup; YAML loader.
//   - Skill / Step / Arg / Expr  — Layer 2 typed AST.
//   - LineariseDhnt  — AST → Layer 1.5 (a-z + spaces only).
//   - LineariseLang  — AST → Layer 1 in any registered language.
//   - ParseDhnt      — Layer 1.5 → AST.
//
// The LLM constrained-decoded slot-filler (Layer 0 → Layer 2) and the
// Wasm leaf adapter (Layer 3) are separable follow-ups.
package skills
