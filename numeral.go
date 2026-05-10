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

// EncodeDecimal encodes a non-negative decimal integer into the dhnt
// numeral form with the "ju" prefix. The digit mapping follows the
// dhnt spec: 1→a, 2→b, 3→c, 4→d, 5→e, 6→f, 7→g, 8→h, 9→i, 0→j. Each
// digit letter is paired with its row vowel ("a"→"a" since 'a' is a
// vowel itself, "b"→"ba", ..., "j"→"ji") to produce the full form.
//
// Examples:
//
//	0     → "juji"           (full form of ju + j + j's row vowel i)
//	12    → "juaba"          (ju + a + ba)
//	2018  → "jubajiahe"      (ju + ba + ji + a + he)
func EncodeDecimal(n uint64) string {
	if n == 0 {
		return "ju" + decimalDigit('0')
	}
	// build digits high-to-low
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	var b strings.Builder
	b.WriteString("ju")
	for _, d := range digits {
		b.WriteString(decimalDigit(d))
	}
	return b.String()
}

// decimalDigit returns the dhnt full-form syllable for a single ASCII
// decimal digit byte ('0'..'9').
func decimalDigit(d byte) string {
	letter := digitLetter(d)
	if isVowel(letter) {
		return string([]byte{letter})
	}
	v := rowVowel(letter)
	return string([]byte{letter, v})
}

func digitLetter(d byte) byte {
	switch d {
	case '1':
		return 'a'
	case '2':
		return 'b'
	case '3':
		return 'c'
	case '4':
		return 'd'
	case '5':
		return 'e'
	case '6':
		return 'f'
	case '7':
		return 'g'
	case '8':
		return 'h'
	case '9':
		return 'i'
	case '0':
		return 'j'
	}
	return 0
}

// DecodeDecimal parses a dhnt decimal numeral (with required "ju"
// prefix in this alpha) and returns its uint64 value. It accepts both
// full and contracted forms by virtue of mapping each digit letter
// independently. Returns an error on malformed input.
func DecodeDecimal(s string) (uint64, error) {
	if !strings.HasPrefix(s, "ju") {
		return 0, fmt.Errorf("dhnt: decimal numeral must start with %q, got %q", "ju", s)
	}
	rest := s[2:]
	if rest == "" {
		return 0, fmt.Errorf("dhnt: decimal numeral has no digits: %q", s)
	}
	var n uint64
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		d, ok := letterDigit(c)
		if !ok {
			return 0, fmt.Errorf("dhnt: not a digit letter %q in numeral %q", c, s)
		}
		n = n*10 + uint64(d)
		// skip a trailing row vowel if present (full form)
		if !isVowel(c) && i+1 < len(rest) && rest[i+1] == rowVowel(c) {
			i++
		}
	}
	return n, nil
}

func letterDigit(c byte) (uint8, bool) {
	switch c {
	case 'a':
		return 1, true
	case 'b':
		return 2, true
	case 'c':
		return 3, true
	case 'd':
		return 4, true
	case 'e':
		return 5, true
	case 'f':
		return 6, true
	case 'g':
		return 7, true
	case 'h':
		return 8, true
	case 'i':
		return 9, true
	case 'j':
		return 0, true
	}
	return 0, false
}
