// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"fmt"
	"strings"
	"testing"
)

// TestNormalise_FakeModel exercises the L0→L2 pipeline with a
// deterministic completer (no LLM): prose → CNL → parsed Skill, with
// loan-word encoding of the free-form name.
func TestNormalise_FakeModel(t *testing.T) {
	g := loadSeedGlossary(t)
	fake := func(prompt string) (string, error) {
		// tool chatter around the required marker block, with a free-form
		// skill/step name that must be loan-encoded.
		return "sure! here you go:\n<dhnt> skill greet needs core step say print value text </dhnt>\nhope that helps", nil
	}
	skill, cnl, err := Normalise("Print the text. Needs core.", g, "en", fake, 1)
	if err != nil {
		t.Fatalf("Normalise: %v", err)
	}
	if skill.Name == "" || len(skill.Steps) != 1 {
		t.Fatalf("unexpected skill: %#v (cnl=%q)", skill, cnl)
	}
	// the parsed skill must itself be valid (re-linearises cleanly)
	if _, err := LineariseDhnt(skill); err != nil {
		t.Errorf("normalised skill does not linearise: %v", err)
	}
	if skill.Steps[0].Primitive != "porinito" { // "print"
		t.Errorf("expected print primitive, got %q", skill.Steps[0].Primitive)
	}
}

// TestNormalise_RetriesOnBadOutput: a first unparseable reply is fed back
// and the second, valid reply is accepted.
func TestNormalise_RetriesOnBadOutput(t *testing.T) {
	g := loadSeedGlossary(t)
	var n int
	fake := func(prompt string) (string, error) {
		n++
		if n == 1 {
			return "I cannot help with that.", nil // no <dhnt> block
		}
		return "<dhnt>skill greet step say log numibero 7</dhnt>", nil
	}
	skill, _, err := Normalise("log the number 7", g, "en", fake, 2)
	if err != nil {
		t.Fatalf("Normalise: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 completer calls, got %d", n)
	}
	if len(skill.Steps) != 1 || skill.Steps[0].Primitive != "loge" {
		t.Errorf("unexpected skill after retry: %#v", skill)
	}
}

// TestNormalise_FailsAfterRetries: persistently bad output errors out.
func TestNormalise_FailsAfterRetries(t *testing.T) {
	g := loadSeedGlossary(t)
	fake := func(prompt string) (string, error) { return "nope", nil }
	if _, _, err := Normalise("whatever", g, "en", fake, 1); err == nil {
		t.Errorf("expected failure after retries")
	}
}

// TestNormalisePrompt_ListsVocabulary: the prompt must surface the glossary
// words the model is allowed to use.
func TestNormalisePrompt_ListsVocabulary(t *testing.T) {
	g := loadSeedGlossary(t)
	p := NormalisePrompt(g, "en")
	for _, w := range []string{"skill", "step", "ensure", "print", "green", "read"} {
		if !strings.Contains(p, w) {
			t.Errorf("prompt missing vocabulary %q", w)
		}
	}
	_ = fmt.Sprint(p)
}
