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
	"sort"
	"strings"
)

// Completer is a pluggable text model: given a prompt it returns the
// model's text. It is how the L0→L2 normaliser stays free of any LLM
// SDK — a CLI-backed implementation lives in skills/tui (Completer), and
// tests inject a deterministic fake.
type Completer func(prompt string) (string, error)

// Normalise turns free prose (Layer 0) into a Skill (Layer 2). It asks
// the model to emit the dhnt CNL (Layer 1) constrained to the glossary,
// wrapped in <dhnt>…</dhnt> markers, then parses it with the loan-word
// rule. Output that fails to parse is fed back for up to `retries`
// further attempts. Validity is exactly transpilability: a result is
// returned only if it parses cleanly. Returns the Skill and the accepted
// Layer 1 CNL string.
//
// This is the constrained-decoded slot-filler in spirit: the model is
// free-form, but only glossary-grounded, parseable output is accepted.
func Normalise(prose string, g *Glossary, lang string, complete Completer, retries int) (Skill, string, error) {
	if g == nil {
		return Skill{}, "", fmt.Errorf("skills: nil glossary")
	}
	if complete == nil {
		return Skill{}, "", fmt.Errorf("skills: nil completer")
	}
	sys := NormalisePrompt(g, lang)
	prompt := sys + "\n\nProcedure to translate:\n" + prose
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		out, err := complete(prompt)
		if err != nil {
			return Skill{}, "", fmt.Errorf("skills: completer: %w", err)
		}
		cnl, ok := extractCNL(out)
		if !ok {
			lastErr = fmt.Errorf("no <dhnt>…</dhnt> block in model output")
			prompt = sys + "\n\nProcedure to translate:\n" + prose +
				"\n\nYour previous reply had no <dhnt>…</dhnt> block. Output exactly one."
			continue
		}
		skill, err := parseLangWith(cnl, g, lang, true)
		if err != nil {
			lastErr = err
			prompt = sys + "\n\nProcedure to translate:\n" + prose +
				fmt.Sprintf("\n\nYour previous CNL %q failed to parse: %v. Use only the listed words and the grammar.", cnl, err)
			continue
		}
		return skill, cnl, nil
	}
	return Skill{}, "", fmt.Errorf("skills: normalise failed after %d attempts: %w", retries+1, lastErr)
}

// extractCNL pulls the text between the first <dhnt> and </dhnt> markers,
// tolerating surrounding tool chatter and code fences.
func extractCNL(s string) (string, bool) {
	i := strings.Index(s, "<dhnt>")
	if i < 0 {
		return "", false
	}
	j := strings.Index(s[i:], "</dhnt>")
	if j < 0 {
		return "", false
	}
	inner := s[i+len("<dhnt>") : i+j]
	inner = strings.ReplaceAll(inner, "`", " ")
	return strings.TrimSpace(inner), true
}

// NormalisePrompt builds the instruction given to the model: the dhnt CNL
// grammar plus the exact glossary vocabulary (in the target language), so
// the model can only assemble valid skills.
func NormalisePrompt(g *Glossary, lang string) string {
	var b strings.Builder
	b.WriteString("You translate a described procedure into the dhnt skill CNL.\n")
	b.WriteString("Output ONLY the CNL, on one line, wrapped in <dhnt> and </dhnt>.\n\n")
	b.WriteString("Grammar (square brackets optional, * repeats):\n")
	b.WriteString("  skill <name> [needs <cap>*] [effect <eff>*] [ensure <predicate> [<argname> <value>]*]*\n")
	b.WriteString("    [step <name> <primitive> [<argname> <value>]*]* [when <predicate> <step>* [else <step>*] fini]*\n\n")
	b.WriteString("Use ONLY these words for keywords, primitives, predicates, effects and types.\n")
	b.WriteString("Names (skill/step names) may be any plain word — they are encoded automatically.\n\n")

	for _, kind := range []EntryKind{KindKeyword, KindPrimitive, KindPredicate, KindEffect, KindType, KindCapability} {
		var words []string
		for _, e := range g.Entries() {
			if e.Kind != kind {
				continue
			}
			if lbl := e.PrimaryLabel(lang); lbl != "" {
				words = append(words, lbl)
			}
		}
		if len(words) == 0 {
			continue
		}
		sort.Strings(words)
		fmt.Fprintf(&b, "%ss: %s\n", kind, strings.Join(words, ", "))
	}
	b.WriteString("\nExample: <dhnt>skill greet needs core step say print value text</dhnt>")
	return b.String()
}
