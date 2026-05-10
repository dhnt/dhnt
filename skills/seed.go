// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import _ "embed"

//go:embed testdata/glossary.yaml
var seedGlossaryYAML []byte

// SeedGlossaryYAML returns the raw YAML bytes of the minimal generic
// glossary that ships with this package. It covers the structural
// keywords needed by ParseDhnt/LineariseDhnt (skill, step, needs,
// fini), a handful of generic primitives (print, log, format), the
// generic types (text, number), and a single "core" capability
// foreign atom.
//
// Consumers typically call SeedGlossary() instead; this accessor
// exists for callers that want to inspect the raw bytes (e.g. to
// merge with another YAML source before parsing).
func SeedGlossaryYAML() []byte {
	out := make([]byte, len(seedGlossaryYAML))
	copy(out, seedGlossaryYAML)
	return out
}

// SeedGlossary returns a freshly-built Glossary containing the
// minimal generic seed entries shipped with this package. Use
// Glossary.Merge to layer a domain-specific glossary on top.
//
// Returns the same content as LoadGlossary("testdata/glossary.yaml")
// but does not require the testdata directory to exist on disk —
// the data is embedded into the binary.
func SeedGlossary() (*Glossary, error) {
	return ParseGlossary(seedGlossaryYAML)
}
