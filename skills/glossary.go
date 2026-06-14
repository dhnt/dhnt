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
	"os"
	"strings"

	"github.com/dhnt/dhnt"
	"gopkg.in/yaml.v3"
)

// EntryKind classifies what role a glossary entry plays in skill
// expressions.
type EntryKind string

const (
	KindKeyword    EntryKind = "keyword"
	KindCapability EntryKind = "capability"
	KindType       EntryKind = "type"
	KindPrimitive  EntryKind = "primitive"
	// KindPredicate is a contract check (pillar P1): it is invoked
	// inside an `enisure` clause and its evaluation yields a dhnt
	// boolean (bua/bub).
	KindPredicate EntryKind = "predicate"
)

// LangAll is a special label key that means "the same surface form
// applies in every language" — used for foreign atoms (project names,
// command names like git/oauth) that should be preserved verbatim
// across all linearisations.
const LangAll = "all"

// Entry is one concept in the closed multilingual glossary. The Dhnt
// field is the canonical machine identifier — short, regular, a-z
// only, derived from the primary English label by the dhnt loan-word
// rule. Labels are display projections per language.
type Entry struct {
	Dhnt   string              `yaml:"dhnt"`
	Kind   EntryKind           `yaml:"kind"`
	Labels map[string][]string `yaml:"labels"`
}

// PrimaryLabel returns the first label registered for the given
// language tag, falling back to the LangAll label if no per-language
// label is set. Returns "" if neither is available.
func (e *Entry) PrimaryLabel(lang string) string {
	if labels, ok := e.Labels[lang]; ok && len(labels) > 0 {
		return labels[0]
	}
	if labels, ok := e.Labels[LangAll]; ok && len(labels) > 0 {
		return labels[0]
	}
	return ""
}

// Glossary is an immutable closed lexicon. Build it once via
// LoadGlossary, ParseGlossary, or NewGlossary; all lookups are O(1)
// map probes.
type Glossary struct {
	entries  []Entry
	byDhnt   map[string]*Entry
	byLabel  map[string]map[string]*Entry // lang → label → entry
	langSeen map[string]struct{}
}

// NewGlossary constructs a Glossary from a slice of entries, building
// the bidirectional lookup tables and rejecting any inconsistencies
// (duplicate dhnt keys, ill-formed dhnt forms, label collisions
// within a language).
func NewGlossary(entries []Entry) (*Glossary, error) {
	g := &Glossary{
		entries:  make([]Entry, len(entries)),
		byDhnt:   make(map[string]*Entry, len(entries)),
		byLabel:  make(map[string]map[string]*Entry),
		langSeen: make(map[string]struct{}),
	}
	copy(g.entries, entries)

	for i := range g.entries {
		e := &g.entries[i]
		if e.Dhnt == "" {
			return nil, fmt.Errorf("glossary: entry %d missing dhnt key", i)
		}
		if !dhnt.IsCanonical(e.Dhnt) {
			return nil, fmt.Errorf("glossary: entry %d dhnt key %q is not canonical dhnt", i, e.Dhnt)
		}
		if e.Kind == "" {
			return nil, fmt.Errorf("glossary: entry %q missing kind", e.Dhnt)
		}
		if _, dup := g.byDhnt[e.Dhnt]; dup {
			return nil, fmt.Errorf("glossary: duplicate dhnt key %q", e.Dhnt)
		}
		g.byDhnt[e.Dhnt] = e

		for lang, labels := range e.Labels {
			if lang == "" {
				return nil, fmt.Errorf("glossary: entry %q has empty language key", e.Dhnt)
			}
			g.langSeen[lang] = struct{}{}
			perLang, ok := g.byLabel[lang]
			if !ok {
				perLang = make(map[string]*Entry)
				g.byLabel[lang] = perLang
			}
			for _, label := range labels {
				if label == "" {
					return nil, fmt.Errorf("glossary: entry %q has empty label in lang %q", e.Dhnt, lang)
				}
				key := normaliseLabel(label)
				if existing, ok := perLang[key]; ok && existing != e {
					return nil, fmt.Errorf("glossary: label %q collides between entries %q and %q in lang %q",
						label, existing.Dhnt, e.Dhnt, lang)
				}
				perLang[key] = e
			}
		}
	}
	return g, nil
}

// LoadGlossary reads a YAML document from path and builds a Glossary.
// The YAML root may be either a list of entries or a map with an
// "entries" key (both shapes are accepted).
func LoadGlossary(path string) (*Glossary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("glossary: read %s: %w", path, err)
	}
	return ParseGlossary(data)
}

// ParseGlossary builds a Glossary from raw YAML bytes.
func ParseGlossary(data []byte) (*Glossary, error) {
	// Try the map-with-entries shape first.
	var wrapper struct {
		Entries []Entry `yaml:"entries"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err == nil && len(wrapper.Entries) > 0 {
		return NewGlossary(wrapper.Entries)
	}
	// Fall back to a top-level list of entries.
	var entries []Entry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("glossary: parse yaml: %w", err)
	}
	return NewGlossary(entries)
}

// Merge returns a new Glossary that contains every entry from g and
// every entry from other. If both glossaries define the same dhnt key
// the resulting glossary fails to construct (duplicate-key rejection).
// Use this to layer a domain-specific addendum on top of a base
// glossary without mutating either.
func (g *Glossary) Merge(other *Glossary) (*Glossary, error) {
	if g == nil {
		return other, nil
	}
	if other == nil {
		return g, nil
	}
	combined := make([]Entry, 0, len(g.entries)+len(other.entries))
	combined = append(combined, g.entries...)
	combined = append(combined, other.entries...)
	return NewGlossary(combined)
}

// LookupDhnt returns the entry with the given canonical dhnt key, or
// nil if not present.
func (g *Glossary) LookupDhnt(dhnt string) *Entry {
	return g.byDhnt[dhnt]
}

// LookupLabel returns the entry whose label in lang matches the given
// surface form (case-insensitive, whitespace-trimmed). Falls back to
// LangAll labels if no per-language match is found. Returns nil if
// nothing matches.
func (g *Glossary) LookupLabel(lang, label string) *Entry {
	key := normaliseLabel(label)
	if perLang, ok := g.byLabel[lang]; ok {
		if e, ok := perLang[key]; ok {
			return e
		}
	}
	if perLang, ok := g.byLabel[LangAll]; ok {
		if e, ok := perLang[key]; ok {
			return e
		}
	}
	return nil
}

// Languages returns the sorted list of language tags this glossary
// has at least one label for. The LangAll virtual tag is included if
// any entry uses it.
func (g *Glossary) Languages() []string {
	out := make([]string, 0, len(g.langSeen))
	for lang := range g.langSeen {
		out = append(out, lang)
	}
	sortStrings(out)
	return out
}

// Entries returns a snapshot of the underlying entry slice. Callers
// must not mutate.
func (g *Glossary) Entries() []Entry {
	return g.entries
}

// Len returns the number of entries.
func (g *Glossary) Len() int { return len(g.entries) }

func normaliseLabel(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// small inline sort to avoid pulling in sort just for one helper.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
