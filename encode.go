// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dhnt

import (
	"fmt"
	"strings"
)

// rowVowel maps each lowercase consonant to its dhnt "row vowel" — the
// leading vowel of the row that consonant lives in:
//
//	row a: b c d
//	row e: f g h
//	row i: j k l m n
//	row o: p q r s t
//	row u: v w x y z
//
// Vowels themselves return 0; callers check isVowel first.
func rowVowel(c byte) byte {
	switch c {
	case 'b', 'c', 'd':
		return 'a'
	case 'f', 'g', 'h':
		return 'e'
	case 'j', 'k', 'l', 'm', 'n':
		return 'i'
	case 'p', 'q', 'r', 's', 't':
		return 'o'
	case 'v', 'w', 'x', 'y', 'z':
		return 'u'
	}
	return 0
}

func isVowel(c byte) bool {
	return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u'
}

func isLowerLetter(c byte) bool {
	return c >= 'a' && c <= 'z'
}

// EncodeWord applies the dhnt vowel-insertion rule to a single
// lowercase a-z word. The output is the canonical full form: a sequence
// of (V|CV) syllables with no consonant clusters and (apart from
// distinct adjacent single-vowel syllables) no vowel clusters either.
//
// Rules, applied character-by-character:
//   - A vowel emits itself.
//   - A consonant followed by a vowel emits the consonant plus that
//     vowel (the input vowel takes precedence over the row vowel).
//   - A consonant followed by a consonant or by end-of-word emits the
//     consonant plus its row vowel.
//
// Returns an error if the input contains any character outside [a-z].
func EncodeWord(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	var b strings.Builder
	b.Grow(len(s) * 2)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !isLowerLetter(c) {
			return "", fmt.Errorf("dhnt: input must be lowercase a-z, got %q at position %d in %q", c, i, s)
		}
		if isVowel(c) {
			b.WriteByte(c)
			continue
		}
		// consonant: write it, then write the next vowel (input or row).
		b.WriteByte(c)
		if i+1 < len(s) && isVowel(s[i+1]) {
			b.WriteByte(s[i+1])
			i++
			continue
		}
		v := rowVowel(c)
		if v == 0 {
			return "", fmt.Errorf("dhnt: no row vowel for byte %q", c)
		}
		b.WriteByte(v)
	}
	return b.String(), nil
}

// EncodePhrase splits on whitespace, encodes each word, and rejoins
// with single spaces. Empty input returns empty output.
func EncodePhrase(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	words := strings.Fields(strings.ToLower(s))
	out := make([]string, len(words))
	for i, w := range words {
		enc, err := EncodeWord(w)
		if err != nil {
			return "", err
		}
		out[i] = enc
	}
	return strings.Join(out, " "), nil
}

// IsCanonical reports whether s is a valid dhnt full-form word: a
// non-empty sequence of (V|CV) syllables in [a-z]+, with no consonant
// clusters. It is the parser-side validator for canonical word tokens.
func IsCanonical(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !isLowerLetter(c) {
			return false
		}
		if isVowel(c) {
			continue
		}
		// consonant must be followed by a vowel (no end-of-word
		// bare consonant in full form, and no consonant cluster).
		if i+1 >= len(s) || !isVowel(s[i+1]) {
			return false
		}
		i++ // consume the vowel of this CV syllable
	}
	return true
}
