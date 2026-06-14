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
	keywordSkill  = "sokilili"  // dhnt of "skill"
	keywordNeeds  = "needaso"   // dhnt of "needs"
	keywordEffect = "efefecato" // dhnt of "effect" — effect-cap block
	keywordEnsure = "enisure"   // dhnt of "ensure" — a contract clause
	keywordStep   = "sotepo"    // dhnt of "step"
	keywordWhen   = "wuheni"    // dhnt of "when" — a flow-control branch
	keywordElse   = "elise"     // dhnt of "else" — the else arm of a branch
	keywordEnd    = "fini"      // dhnt of "fin" — block terminator

	latKeyword = "latitude" // reserved step arg name for the P4 dial
	latExact   = "exacato"  // dhnt of "exact"
	latJudge   = "judage"   // dhnt of "judge"
)

// latAtom maps a Latitude to its Layer 1.5 atom. Only non-default
// (LatJudge) is ever emitted.
func latAtom(l Latitude) string {
	switch l {
	case LatJudge:
		return latJudge
	default:
		return latExact
	}
}

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

	if len(s.EffectCap) > 0 {
		b.WriteByte(' ')
		b.WriteString(keywordEffect)
		for _, e := range s.EffectCap {
			atom := e.Dhnt()
			if atom == "" {
				return "", fmt.Errorf("skills: effect cap contains invalid effect %d", int(e))
			}
			b.WriteByte(' ')
			b.WriteString(atom)
		}
		b.WriteByte(' ')
		b.WriteString(keywordEnd)
	}

	for i := range s.Contract {
		c := &s.Contract[i]
		if !dhnt.IsCanonical(c.Predicate) {
			return "", fmt.Errorf("skills: contract %d predicate %q is not canonical dhnt", i, c.Predicate)
		}
		b.WriteByte(' ')
		b.WriteString(keywordEnsure)
		b.WriteByte(' ')
		b.WriteString(c.Predicate)
		for j := range c.Args {
			a := &c.Args[j]
			if !dhnt.IsCanonical(a.Name) {
				return "", fmt.Errorf("skills: contract %q arg %d name %q is not canonical dhnt", c.Predicate, j, a.Name)
			}
			b.WriteByte(' ')
			b.WriteString(a.Name)
			b.WriteByte(' ')
			val, err := lineariseExpr(a.Value)
			if err != nil {
				return "", fmt.Errorf("skills: contract %q arg %q: %w", c.Predicate, a.Name, err)
			}
			b.WriteString(val)
		}
		b.WriteByte(' ')
		b.WriteString(keywordEnd)
	}

	if err := lineariseStepsDhnt(&b, s.Steps); err != nil {
		return "", err
	}

	b.WriteByte(' ')
	b.WriteString(keywordEnd)
	return b.String(), nil
}

// lineariseArgsDhnt emits ` name value` pairs for an arg list.
func lineariseArgsDhnt(b *strings.Builder, who string, args []Arg) error {
	for j := range args {
		a := &args[j]
		if !dhnt.IsCanonical(a.Name) {
			return fmt.Errorf("skills: %s arg %d name %q is not canonical dhnt", who, j, a.Name)
		}
		b.WriteByte(' ')
		b.WriteString(a.Name)
		b.WriteByte(' ')
		val, err := lineariseExpr(a.Value)
		if err != nil {
			return fmt.Errorf("skills: %s arg %q: %w", who, a.Name, err)
		}
		b.WriteString(val)
	}
	return nil
}

// lineariseStepsDhnt renders a step list, recursing into branches.
func lineariseStepsDhnt(b *strings.Builder, steps []Step) error {
	for i := range steps {
		st := &steps[i]
		if st.Branch != nil {
			if err := lineariseBranchDhnt(b, st.Branch); err != nil {
				return err
			}
			continue
		}
		if !dhnt.IsCanonical(st.Name) {
			return fmt.Errorf("skills: step %d name %q is not canonical dhnt", i, st.Name)
		}
		if !dhnt.IsCanonical(st.Primitive) {
			return fmt.Errorf("skills: step %q primitive %q is not canonical dhnt", st.Name, st.Primitive)
		}
		b.WriteByte(' ')
		b.WriteString(keywordStep)
		b.WriteByte(' ')
		b.WriteString(st.Name)
		b.WriteByte(' ')
		b.WriteString(st.Primitive)
		if st.Latitude != LatExact {
			b.WriteByte(' ')
			b.WriteString(latKeyword)
			b.WriteByte(' ')
			b.WriteString(latAtom(st.Latitude))
		}
		if err := lineariseArgsDhnt(b, "step "+st.Name, st.Args); err != nil {
			return err
		}
		b.WriteByte(' ')
		b.WriteString(keywordEnd)
	}
	return nil
}

// lineariseBranchDhnt renders `wuheni <cond> <then> [elise <else>] fini`.
func lineariseBranchDhnt(b *strings.Builder, br *Branch) error {
	if !dhnt.IsCanonical(br.Cond.Predicate) {
		return fmt.Errorf("skills: branch condition %q is not canonical dhnt", br.Cond.Predicate)
	}
	b.WriteByte(' ')
	b.WriteString(keywordWhen)
	b.WriteByte(' ')
	b.WriteString(br.Cond.Predicate)
	if err := lineariseArgsDhnt(b, "branch "+br.Cond.Predicate, br.Cond.Args); err != nil {
		return err
	}
	if err := lineariseStepsDhnt(b, br.Then); err != nil {
		return err
	}
	if len(br.Else) > 0 {
		b.WriteByte(' ')
		b.WriteString(keywordElse)
		if err := lineariseStepsDhnt(b, br.Else); err != nil {
			return err
		}
	}
	b.WriteByte(' ')
	b.WriteString(keywordEnd)
	return nil
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

	if len(s.EffectCap) > 0 {
		b.WriteByte(' ')
		b.WriteString(resolve(keywordEffect))
		for _, e := range s.EffectCap {
			b.WriteByte(' ')
			b.WriteString(resolve(e.Dhnt()))
		}
	}

	for i := range s.Contract {
		c := &s.Contract[i]
		b.WriteByte(' ')
		b.WriteString(resolve(keywordEnsure))
		b.WriteByte(' ')
		b.WriteString(resolve(c.Predicate))
		if err := lineariseArgsLang(&b, resolve, "contract "+c.Predicate, c.Args); err != nil {
			return "", err
		}
	}

	if err := lineariseStepsLang(&b, resolve, s.Steps); err != nil {
		return "", err
	}
	return b.String(), nil
}

// lineariseArgsLang emits ` label value` pairs in the target language.
func lineariseArgsLang(b *strings.Builder, resolve func(string) string, who string, args []Arg) error {
	for j := range args {
		a := &args[j]
		b.WriteByte(' ')
		b.WriteString(resolve(a.Name))
		b.WriteByte(' ')
		switch a.Value.Kind {
		case ExprRef:
			b.WriteString(resolve(a.Value.Ref))
		case ExprNumber:
			fmt.Fprintf(b, "%d", a.Value.Number)
		default:
			return fmt.Errorf("%s arg %q: invalid expr kind %d", who, a.Name, a.Value.Kind)
		}
	}
	return nil
}

// lineariseStepsLang renders a step list in the target language,
// recursing into branches. Leaf steps carry no terminator (Layer 1
// convention), but a branch is closed by `fini` so it stays unambiguous
// when re-parsed by ParseLang.
func lineariseStepsLang(b *strings.Builder, resolve func(string) string, steps []Step) error {
	for i := range steps {
		st := &steps[i]
		if st.Branch != nil {
			b.WriteByte(' ')
			b.WriteString(resolve(keywordWhen))
			b.WriteByte(' ')
			b.WriteString(resolve(st.Branch.Cond.Predicate))
			if err := lineariseArgsLang(b, resolve, "branch", st.Branch.Cond.Args); err != nil {
				return err
			}
			if err := lineariseStepsLang(b, resolve, st.Branch.Then); err != nil {
				return err
			}
			if len(st.Branch.Else) > 0 {
				b.WriteByte(' ')
				b.WriteString(resolve(keywordElse))
				if err := lineariseStepsLang(b, resolve, st.Branch.Else); err != nil {
					return err
				}
			}
			b.WriteByte(' ')
			b.WriteString(resolve(keywordEnd))
			continue
		}
		b.WriteByte(' ')
		b.WriteString(resolve(keywordStep))
		b.WriteByte(' ')
		b.WriteString(resolve(st.Name))
		b.WriteByte(' ')
		b.WriteString(resolve(st.Primitive))
		if st.Latitude != LatExact {
			b.WriteByte(' ')
			b.WriteString(resolve(latKeyword))
			b.WriteByte(' ')
			b.WriteString(resolve(latAtom(st.Latitude)))
		}
		if err := lineariseArgsLang(b, resolve, "step "+st.Name, st.Args); err != nil {
			return err
		}
	}
	return nil
}
