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

// Rule-derived alternative spellings on input (dhnt.md, "Canonical words
// and rule-derived alternative spellings"). ParseDhnt reads every word
// token by its expansion N(token) = dhnt.EncodeWord(token), so `ls`
// reads as `liso` and `bash`/`bsh` as `basohe`. N is idempotent, so a
// canonical token reads as itself. The AST — and therefore L1.5 output
// and Identity (P0) — only ever holds the canonical expansion.

// canonWord returns N(tok). Layer 1.5 input is charset-validated before
// tokenising, so EncodeWord cannot fail here; tok is returned unchanged
// if it ever does, leaving the caller's validation to reject it.
func canonWord(tok string) string {
	n, err := dhnt.EncodeWord(tok)
	if err != nil || n == "" {
		return tok
	}
	return n
}

// spelled renders a token with its expansion for error messages.
func spelled(tok string) string {
	return fmt.Sprintf("%q (rule-derived form %q)", tok, canonWord(tok))
}

// sense is one distinct entry a token may denote in a single
// grammatical position: a human-readable label and its canonical form.
type sense struct {
	label     string
	canonical string
}

// resolveSense matches N(tok) against the senses admissible in one
// grammatical position. More than one distinct match is a homonym the
// position cannot disambiguate: it is an error naming every candidate,
// never a guess. No match is an "unknown" error showing N(tok).
func resolveSense(position, tok string, senses []sense) (sense, error) {
	n := canonWord(tok)
	var hits []sense
	for _, s := range senses {
		if s.canonical == n {
			hits = append(hits, s)
		}
	}
	switch len(hits) {
	case 1:
		return hits[0], nil
	case 0:
		return sense{}, fmt.Errorf("unknown %s %s", position, spelled(tok))
	}
	labels := make([]string, len(hits))
	for i, h := range hits {
		labels[i] = h.label
	}
	return sense{}, fmt.Errorf("ambiguous %s %s: matches %s", position, spelled(tok), strings.Join(labels, ", "))
}

// Closed sets resolved by ParseDhnt, one per grammatical position.
var (
	skillBodySenses = []sense{
		{"fin", keywordEnd},
		{"needs", keywordNeeds},
		{"effect", keywordEffect},
		{"ensure", keywordEnsure},
		{"step", keywordStep},
		{"when", keywordWhen},
		{"on-fail", keywordOnFail},
	}
	policySenses = []sense{
		{"abort", policyAbort},
		{"retry", policyRetry},
		{"blockers", policyBlockers},
	}
	latitudeSenses = []sense{
		{"exact", latExact},
		{"judge", latJudge},
	}
)

// effectSenses lists the closed effect lattice (P3) in stable order.
func effectSenses() []sense {
	out := make([]sense, 0, len(effectDhnt))
	for e := EffRead; e.Dhnt() != ""; e++ {
		out = append(out, sense{e.String(), e.Dhnt()})
	}
	return out
}
