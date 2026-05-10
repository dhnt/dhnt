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

	"github.com/dhnt/dhnt"
)

// Structural keywords used as block markers in Layer 1.5. These are
// also valid canonical dhnt words so they go through the parser as
// regular tokens, but the parser/lineariser recognises them by
// literal value.
//
// The choice is deliberate: every structural marker is itself a
// well-formed dhnt word, so Layer 1.5 has no out-of-band syntax —
// only [a-z] tokens separated by whitespace.
const (
	keywordSkill = "sokilili" // dhnt of "skill"
	keywordNeeds = "needaso"  // dhnt of "needs"
	keywordStep  = "sotepo"   // dhnt of "step"
	keywordEnd   = "fini"     // dhnt of "fin" — block terminator
)

// LineariseDhnt renders a Skill into Layer 1.5 (dhnt canonical form).
// The output is strictly [a-z] characters and ASCII space. It is the
// machine-internal form used for hashing, diffing, and the
// transpilability validation rule (skill is valid iff this function
// produces output that re-parses to an equivalent Skill).
func LineariseDhnt(s Skill) (string, error) {
	if s.Name == "" {
		return "", fmt.Errorf("skills: skill has no name")
	}
	if !dhnt.IsCanonical(s.Name) {
		return "", fmt.Errorf("skills: skill name %q is not canonical dhnt", s.Name)
	}
	var b strings.Builder
	b.WriteString(keywordSkill)
	b.WriteByte(' ')
	b.WriteString(s.Name)

	if len(s.Caps) > 0 {
		b.WriteByte(' ')
		b.WriteString(keywordNeeds)
		for _, cap := range s.Caps {
			if !dhnt.IsCanonical(cap) {
				return "", fmt.Errorf("skills: capability %q is not canonical dhnt", cap)
			}
			b.WriteByte(' ')
			b.WriteString(cap)
		}
		b.WriteByte(' ')
		b.WriteString(keywordEnd)
	}

	for i := range s.Steps {
		st := &s.Steps[i]
		if !dhnt.IsCanonical(st.Name) {
			return "", fmt.Errorf("skills: step %d name %q is not canonical dhnt", i, st.Name)
		}
		if !dhnt.IsCanonical(st.Primitive) {
			return "", fmt.Errorf("skills: step %q primitive %q is not canonical dhnt", st.Name, st.Primitive)
		}
		b.WriteByte(' ')
		b.WriteString(keywordStep)
		b.WriteByte(' ')
		b.WriteString(st.Name)
		b.WriteByte(' ')
		b.WriteString(st.Primitive)
		for j := range st.Args {
			a := &st.Args[j]
			if !dhnt.IsCanonical(a.Name) {
				return "", fmt.Errorf("skills: step %q arg %d name %q is not canonical dhnt", st.Name, j, a.Name)
			}
			b.WriteByte(' ')
			b.WriteString(a.Name)
			b.WriteByte(' ')
			val, err := lineariseExpr(a.Value)
			if err != nil {
				return "", fmt.Errorf("skills: step %q arg %q: %w", st.Name, a.Name, err)
			}
			b.WriteString(val)
		}
		b.WriteByte(' ')
		b.WriteString(keywordEnd)
	}

	b.WriteByte(' ')
	b.WriteString(keywordEnd)
	return b.String(), nil
}

func lineariseExpr(e Expr) (string, error) {
	switch e.Kind {
	case ExprRef:
		if !dhnt.IsCanonical(e.Ref) {
			return "", fmt.Errorf("ref %q is not canonical dhnt", e.Ref)
		}
		return e.Ref, nil
	case ExprNumber:
		return dhnt.EncodeDecimal(e.Number), nil
	default:
		return "", fmt.Errorf("invalid expr kind %d", e.Kind)
	}
}

// LineariseLang renders a Skill into Layer 1 in the given language,
// using glossary labels for display. Identifier-shaped tokens that
// have no per-language label fall back to the LangAll label, then to
// the dhnt form itself. The output is human-readable but structural
// keywords still appear in dhnt form (for parser compatibility — the
// LLM normaliser inverts the structural keyword set per language at
// authoring time).
//
// For the alpha, identifier display in Layer 1 uses each glossary
// entry's primary label for the language. Numerals are rendered in
// decimal ASCII for readability.
func LineariseLang(s Skill, g *Glossary, lang string) (string, error) {
	if g == nil {
		return "", fmt.Errorf("skills: nil glossary")
	}
	resolve := func(dh string) string {
		if e := g.LookupDhnt(dh); e != nil {
			if lbl := e.PrimaryLabel(lang); lbl != "" {
				return lbl
			}
		}
		return dh
	}
	var b strings.Builder
	b.WriteString(resolve(keywordSkill))
	b.WriteByte(' ')
	b.WriteString(resolve(s.Name))

	if len(s.Caps) > 0 {
		b.WriteByte(' ')
		b.WriteString(resolve(keywordNeeds))
		for _, cap := range s.Caps {
			b.WriteByte(' ')
			b.WriteString(resolve(cap))
		}
	}

	for i := range s.Steps {
		st := &s.Steps[i]
		b.WriteByte(' ')
		b.WriteString(resolve(keywordStep))
		b.WriteByte(' ')
		b.WriteString(resolve(st.Name))
		b.WriteByte(' ')
		b.WriteString(resolve(st.Primitive))
		for j := range st.Args {
			a := &st.Args[j]
			b.WriteByte(' ')
			b.WriteString(resolve(a.Name))
			b.WriteByte(' ')
			switch a.Value.Kind {
			case ExprRef:
				b.WriteString(resolve(a.Value.Ref))
			case ExprNumber:
				fmt.Fprintf(&b, "%d", a.Value.Number)
			default:
				return "", fmt.Errorf("step %q arg %q: invalid expr kind %d", st.Name, a.Name, a.Value.Kind)
			}
		}
	}
	return b.String(), nil
}
