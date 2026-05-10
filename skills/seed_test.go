// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import "testing"

// TestSeedGlossary verifies that the embedded seed produces a
// glossary equivalent to the file-loaded one. This guarantees
// dependents that consume SeedGlossary() get the same content
// that's in testdata/glossary.yaml.
func TestSeedGlossary(t *testing.T) {
	embedded, err := SeedGlossary()
	if err != nil {
		t.Fatalf("SeedGlossary: %v", err)
	}
	fileLoaded := loadSeedGlossary(t)
	if embedded.Len() != fileLoaded.Len() {
		t.Errorf("seed glossary entry count: embedded=%d, file=%d",
			embedded.Len(), fileLoaded.Len())
	}
	for _, want := range fileLoaded.Entries() {
		got := embedded.LookupDhnt(want.Dhnt)
		if got == nil {
			t.Errorf("embedded seed missing entry %q", want.Dhnt)
			continue
		}
		if got.Kind != want.Kind {
			t.Errorf("entry %q: embedded kind %q, file kind %q",
				want.Dhnt, got.Kind, want.Kind)
		}
	}
}

// TestSeedGlossaryYAML_RoundtripsViaParse verifies that the raw bytes
// accessor returns parseable YAML.
func TestSeedGlossaryYAML_RoundtripsViaParse(t *testing.T) {
	raw := SeedGlossaryYAML()
	if len(raw) == 0 {
		t.Fatal("SeedGlossaryYAML returned empty")
	}
	g, err := ParseGlossary(raw)
	if err != nil {
		t.Fatalf("ParseGlossary on raw seed: %v", err)
	}
	if g.LookupDhnt("sokilili") == nil {
		t.Error("parsed seed missing the sokilili structural keyword")
	}
}
