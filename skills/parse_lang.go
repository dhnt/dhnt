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
	"strconv"
	"strings"

	"github.com/dhnt/dhnt"
)

// ParseLang is the inverse of LineariseLang (pillar P7): it parses a
// Layer 1 string written in `lang` back into the canonical Skill AST.
// A human may therefore edit any natural-language projection and have
// it re-normalise to the one identity:
//
//	ParseLang(LineariseLang(s, g, lang), g, lang) == s
//
// for any skill whose non-keyword identifiers are either glossary
// entries or already canonical dhnt atoms.
//
// Resolution is per token: a glossary label in `lang` (falling back to
// the language-neutral `all` labels) becomes its dhnt id; an ASCII
// decimal numeral becomes a dhnt numeral; anything else must already be
// canonical dhnt (skill/step names carry through verbatim). Unlike
// Layer 1.5, Layer 1 has no `fini` terminators, so blocks run until the
// next structural keyword.
func ParseLang(src string, g *Glossary, lang string) (Skill, error) {
	if g == nil {
		return Skill{}, fmt.Errorf("skills: nil glossary")
	}
	fields := strings.Fields(src)
	toks := make([]string, len(fields))
	for i, f := range fields {
		dh, err := resolveToDhnt(g, lang, f)
		if err != nil {
			return Skill{}, err
		}
		toks[i] = dh
	}
	lp := &langParser{toks: toks}
	return lp.parse()
}

// resolveToDhnt maps one Layer 1 surface token to its canonical dhnt
// form.
func resolveToDhnt(g *Glossary, lang, tok string) (string, error) {
	if e := g.LookupLabel(lang, tok); e != nil {
		return e.Dhnt, nil
	}
	if isASCIIDecimal(tok) {
		n, err := strconv.ParseUint(tok, 10, 64)
		if err != nil {
			return "", fmt.Errorf("skills: numeral %q: %w", tok, err)
		}
		return dhnt.EncodeDecimal(n), nil
	}
	if dhnt.IsCanonical(tok) {
		return tok, nil
	}
	return "", fmt.Errorf("skills: token %q is neither a known %q label, a numeral, nor canonical dhnt", tok, lang)
}

func isASCIIDecimal(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

var langKeyword = map[string]bool{
	keywordSkill:  true,
	keywordNeeds:  true,
	keywordEffect: true,
	keywordEnsure: true,
	keywordStep:   true,
}

func isLangKeyword(t string) bool { return langKeyword[t] }

// langParser parses a keyword-delimited dhnt token stream (Layer 1
// with surface tokens already resolved to dhnt) into a Skill.
type langParser struct {
	toks []string
	pos  int
}

func (p *langParser) peek() (string, bool) {
	if p.pos >= len(p.toks) {
		return "", false
	}
	return p.toks[p.pos], true
}

func (p *langParser) next() (string, bool) {
	t, ok := p.peek()
	if ok {
		p.pos++
	}
	return t, ok
}

func (p *langParser) parse() (Skill, error) {
	first, ok := p.next()
	if !ok || first != keywordSkill {
		return Skill{}, fmt.Errorf("skills: Layer 1 input must start with the skill keyword")
	}
	name, ok := p.next()
	if !ok {
		return Skill{}, fmt.Errorf("skills: expected skill name")
	}
	if isLangKeyword(name) || !dhnt.IsCanonical(name) {
		return Skill{}, fmt.Errorf("skills: invalid skill name %q", name)
	}
	s := Skill{Name: name}
	for {
		kw, ok := p.peek()
		if !ok {
			break
		}
		switch kw {
		case keywordNeeds:
			p.pos++
			caps := p.collectIdents()
			if len(caps) == 0 {
				return Skill{}, fmt.Errorf("skills: empty needs block")
			}
			s.Caps = append(s.Caps, caps...)
		case keywordEffect:
			p.pos++
			atoms := p.collectIdents()
			if len(atoms) == 0 {
				return Skill{}, fmt.Errorf("skills: empty effect block")
			}
			for _, a := range atoms {
				e, ok := effectByDhntAtom(a)
				if !ok {
					return Skill{}, fmt.Errorf("skills: %q is not a known effect atom", a)
				}
				s.EffectCap = append(s.EffectCap, e)
			}
		case keywordEnsure:
			p.pos++
			pred, ok := p.next()
			if !ok || isLangKeyword(pred) || !dhnt.IsCanonical(pred) {
				return Skill{}, fmt.Errorf("skills: ensure expects a predicate")
			}
			_, args, err := p.collectArgs()
			if err != nil {
				return Skill{}, err
			}
			s.Contract = append(s.Contract, Check{Predicate: pred, Args: args})
		case keywordStep:
			p.pos++
			nm, ok := p.next()
			if !ok || isLangKeyword(nm) || !dhnt.IsCanonical(nm) {
				return Skill{}, fmt.Errorf("skills: step expects a name")
			}
			prim, ok := p.next()
			if !ok || isLangKeyword(prim) || !dhnt.IsCanonical(prim) {
				return Skill{}, fmt.Errorf("skills: step %q expects a primitive", nm)
			}
			lat, args, err := p.collectArgs()
			if err != nil {
				return Skill{}, err
			}
			s.Steps = append(s.Steps, Step{Name: nm, Primitive: prim, Latitude: lat, Args: args})
		default:
			return Skill{}, fmt.Errorf("skills: unexpected token %q in Layer 1 body", kw)
		}
	}
	return s, nil
}

// collectIdents consumes canonical tokens until the next keyword or
// end-of-input.
func (p *langParser) collectIdents() []string {
	var out []string
	for {
		t, ok := p.peek()
		if !ok || isLangKeyword(t) {
			break
		}
		p.pos++
		out = append(out, t)
	}
	return out
}

// collectArgs consumes an optional leading `latitude <atom>` dial
// followed by (name, value) pairs, until the next keyword or
// end-of-input. The returned Latitude defaults to LatExact.
func (p *langParser) collectArgs() (Latitude, []Arg, error) {
	lat := LatExact
	var args []Arg
	first := true
	for {
		name, ok := p.peek()
		if !ok || isLangKeyword(name) {
			break
		}
		p.pos++
		if first && name == latKeyword {
			first = false
			atom, ok := p.next()
			if !ok || isLangKeyword(atom) {
				return 0, nil, fmt.Errorf("skills: latitude has no value")
			}
			lv, err := parseLatitudeAtom(atom)
			if err != nil {
				return 0, nil, err
			}
			lat = lv
			continue
		}
		first = false
		if !dhnt.IsCanonical(name) {
			return 0, nil, fmt.Errorf("skills: arg name %q is not canonical dhnt", name)
		}
		valTok, ok := p.peek()
		if !ok || isLangKeyword(valTok) {
			return 0, nil, fmt.Errorf("skills: arg %q has no value", name)
		}
		p.pos++
		val, err := parseValue(valTok)
		if err != nil {
			return 0, nil, fmt.Errorf("skills: arg %q: %w", name, err)
		}
		args = append(args, Arg{Name: name, Value: val})
	}
	return lat, args, nil
}
