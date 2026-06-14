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

// ParseDhnt parses a Layer 1.5 dhnt source string into a Skill AST.
// The input must be strictly [a-z] characters with words separated
// by ASCII whitespace. Any other character is rejected.
//
// Validity is exactly "this transpiles": ParseDhnt accepts iff the
// canonical encoder can have produced this input from some Skill.
func ParseDhnt(src string) (Skill, error) {
	if err := validateLayer15Charset(src); err != nil {
		return Skill{}, err
	}
	tokens := strings.Fields(src)
	p := &parser{tokens: tokens}
	return p.parseSkill()
}

func validateLayer15Charset(src string) error {
	for i := 0; i < len(src); i++ {
		c := src[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		if c < 'a' || c > 'z' {
			return fmt.Errorf("skills: Layer 1.5 character %q at position %d is not a-z or whitespace", c, i)
		}
	}
	return nil
}

type parser struct {
	tokens []string
	pos    int
}

func (p *parser) peek() (string, bool) {
	if p.pos >= len(p.tokens) {
		return "", false
	}
	return p.tokens[p.pos], true
}

func (p *parser) next() (string, bool) {
	tok, ok := p.peek()
	if ok {
		p.pos++
	}
	return tok, ok
}

func (p *parser) expect(literal string) error {
	tok, ok := p.next()
	if !ok {
		return fmt.Errorf("skills: expected %q, got end-of-input", literal)
	}
	if tok != literal {
		return fmt.Errorf("skills: expected %q at token %d, got %q", literal, p.pos-1, tok)
	}
	return nil
}

func (p *parser) parseSkill() (Skill, error) {
	if err := p.expect(keywordSkill); err != nil {
		return Skill{}, err
	}
	name, ok := p.next()
	if !ok {
		return Skill{}, fmt.Errorf("skills: expected skill name after %q", keywordSkill)
	}
	if !dhnt.IsCanonical(name) {
		return Skill{}, fmt.Errorf("skills: skill name %q is not canonical dhnt", name)
	}
	skill := Skill{Name: name}
	for {
		tok, ok := p.peek()
		if !ok {
			return Skill{}, fmt.Errorf("skills: expected %q, got end-of-input", keywordEnd)
		}
		switch tok {
		case keywordEnd:
			p.pos++
			return skill, nil
		case keywordNeeds:
			caps, err := p.parseNeeds()
			if err != nil {
				return Skill{}, err
			}
			skill.Caps = append(skill.Caps, caps...)
		case keywordEffect:
			effs, err := p.parseEffectCap()
			if err != nil {
				return Skill{}, err
			}
			skill.EffectCap = append(skill.EffectCap, effs...)
		case keywordEnsure:
			chk, err := p.parseEnsure()
			if err != nil {
				return Skill{}, err
			}
			skill.Contract = append(skill.Contract, chk)
		case keywordStep:
			step, err := p.parseStep()
			if err != nil {
				return Skill{}, err
			}
			skill.Steps = append(skill.Steps, step)
		default:
			return Skill{}, fmt.Errorf("skills: unexpected token %q in skill body", tok)
		}
	}
}

func (p *parser) parseNeeds() ([]string, error) {
	if err := p.expect(keywordNeeds); err != nil {
		return nil, err
	}
	var caps []string
	for {
		tok, ok := p.next()
		if !ok {
			return nil, fmt.Errorf("skills: expected %q in needs block", keywordEnd)
		}
		if tok == keywordEnd {
			break
		}
		if !dhnt.IsCanonical(tok) {
			return nil, fmt.Errorf("skills: capability %q is not canonical dhnt", tok)
		}
		caps = append(caps, tok)
	}
	if len(caps) == 0 {
		return nil, fmt.Errorf("skills: empty needs block")
	}
	return caps, nil
}

// parseEffectCap parses an effect-cap block (pillar P3):
//
//	efefecato <effect-atom>+ fini
//
// Each atom must be a known effect of the closed lattice; an unknown
// atom is rejected (the cap is a closed set, not free text).
func (p *parser) parseEffectCap() ([]Effect, error) {
	if err := p.expect(keywordEffect); err != nil {
		return nil, err
	}
	var effs []Effect
	for {
		tok, ok := p.next()
		if !ok {
			return nil, fmt.Errorf("skills: expected %q in effect block", keywordEnd)
		}
		if tok == keywordEnd {
			break
		}
		e, ok := effectByDhntAtom(tok)
		if !ok {
			return nil, fmt.Errorf("skills: %q is not a known effect atom", tok)
		}
		effs = append(effs, e)
	}
	if len(effs) == 0 {
		return nil, fmt.Errorf("skills: empty effect block")
	}
	return effs, nil
}

// parseEnsure parses one contract clause:
//
//	enisure <predicate> (<arg-name> <value>)* fini
//
// The predicate is a canonical glossary id (kind "predicate"); the
// argument loop mirrors parseStep and stops at the block terminator.
func (p *parser) parseEnsure() (Check, error) {
	if err := p.expect(keywordEnsure); err != nil {
		return Check{}, err
	}
	pred, ok := p.next()
	if !ok {
		return Check{}, fmt.Errorf("skills: expected predicate after %q", keywordEnsure)
	}
	if !dhnt.IsCanonical(pred) {
		return Check{}, fmt.Errorf("skills: contract predicate %q is not canonical dhnt", pred)
	}
	chk := Check{Predicate: pred}
	for {
		tok, ok := p.next()
		if !ok {
			return Check{}, fmt.Errorf("skills: contract %q expected %q", pred, keywordEnd)
		}
		if tok == keywordEnd {
			break
		}
		argName := tok
		if !dhnt.IsCanonical(argName) {
			return Check{}, fmt.Errorf("skills: contract %q arg name %q is not canonical dhnt", pred, argName)
		}
		valueTok, ok := p.next()
		if !ok {
			return Check{}, fmt.Errorf("skills: contract %q arg %q has no value", pred, argName)
		}
		val, err := parseValue(valueTok)
		if err != nil {
			return Check{}, fmt.Errorf("skills: contract %q arg %q: %w", pred, argName, err)
		}
		chk.Args = append(chk.Args, Arg{Name: argName, Value: val})
	}
	return chk, nil
}

func (p *parser) parseStep() (Step, error) {
	if err := p.expect(keywordStep); err != nil {
		return Step{}, err
	}
	name, ok := p.next()
	if !ok {
		return Step{}, fmt.Errorf("skills: expected step name after %q", keywordStep)
	}
	if !dhnt.IsCanonical(name) {
		return Step{}, fmt.Errorf("skills: step name %q is not canonical dhnt", name)
	}
	primitive, ok := p.next()
	if !ok {
		return Step{}, fmt.Errorf("skills: step %q expected primitive name", name)
	}
	if !dhnt.IsCanonical(primitive) {
		return Step{}, fmt.Errorf("skills: step %q primitive %q is not canonical dhnt", name, primitive)
	}
	step := Step{Name: name, Primitive: primitive}
	for {
		tok, ok := p.next()
		if !ok {
			return Step{}, fmt.Errorf("skills: step %q expected %q", name, keywordEnd)
		}
		if tok == keywordEnd {
			break
		}
		// argument: this token is the arg name; the next token is the value.
		argName := tok
		if !dhnt.IsCanonical(argName) {
			return Step{}, fmt.Errorf("skills: step %q arg name %q is not canonical dhnt", name, argName)
		}
		valueTok, ok := p.next()
		if !ok {
			return Step{}, fmt.Errorf("skills: step %q arg %q has no value", name, argName)
		}
		val, err := parseValue(valueTok)
		if err != nil {
			return Step{}, fmt.Errorf("skills: step %q arg %q: %w", name, argName, err)
		}
		step.Args = append(step.Args, Arg{Name: argName, Value: val})
	}
	return step, nil
}

// parseValue interprets a single value token. A token is a numeral if
// it starts with "ju" and the body decodes cleanly per the dhnt
// numeral rules; otherwise it is a canonical-dhnt reference.
func parseValue(tok string) (Expr, error) {
	if strings.HasPrefix(tok, "ju") && len(tok) > 2 {
		if n, err := dhnt.DecodeDecimal(tok); err == nil {
			return NewNumber(n), nil
		}
		// fall through to ref interpretation
	}
	if !dhnt.IsCanonical(tok) {
		return Expr{}, fmt.Errorf("value %q is not canonical dhnt", tok)
	}
	return NewRef(tok), nil
}
