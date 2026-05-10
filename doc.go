// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package dhnt is the Go reference implementation of the dhnt language
// specification (see dhnt.md in this repository).
//
// dhnt is a constructed interlingua designed to unify all languages —
// programming, constructed, natural — into a single normalised form.
// Its alphabet is the 26 lowercase ASCII letters a–z, organised into
// five vowel-group rows that drive a CV-only phonotactic rule. The
// language is TAM-less and inflection-free; loan words from any source
// language transform deterministically into dhnt-conformant strings.
//
// This root package implements the universal primitives:
//
//   - EncodeWord / EncodePhrase — apply the vowel-insertion rule to
//     lowercase a–z input and produce canonical dhnt full-form output.
//   - IsCanonical — validate that a string is well-formed dhnt
//     (no consonant clusters, no invalid characters).
//   - EncodeDecimal / DecodeDecimal — round-trip non-negative integers
//     through the ju-prefixed numeral encoding.
//
// Higher-level constructions (the multilingual Controlled Natural
// Language for skill specifications, including the closed glossary,
// the typed AST, and the bidirectional parser/lineariser) live in
// the github.com/dhnt/dhnt/skills subpackage.
//
// Status: alpha (v0.1.x). The API may change at minor versions until
// v1.0 commits to stability.
package dhnt
