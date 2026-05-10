// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dhnt

import "testing"

func TestDecimalNumeral(t *testing.T) {
	cases := []struct {
		n    uint64
		want string
	}{
		{0, "juji"},   // ju + j (digit 0) + i (j's row vowel)
		{1, "jua"},    // ju + a (digit 1, vowel itself)
		{9, "jui"},    // ju + i (digit 9, vowel itself)
		{10, "juaji"}, // ju + a (1) + ji (0)
		{12, "juaba"}, // ju + a (1) + ba (2)
		{18, "juahe"}, // ju + a (1) + he (8)
		{2018, "jubajiahe"},
	}
	for _, tc := range cases {
		got := EncodeDecimal(tc.n)
		if got != tc.want {
			t.Errorf("EncodeDecimal(%d) = %q, want %q", tc.n, got, tc.want)
		}
		back, err := DecodeDecimal(got)
		if err != nil {
			t.Errorf("DecodeDecimal(%q) error: %v", got, err)
			continue
		}
		if back != tc.n {
			t.Errorf("Decimal roundtrip: %d → %q → %d", tc.n, got, back)
		}
	}
}

func TestDecimalNumeral_RejectsBadInput(t *testing.T) {
	bad := []string{
		"abc",   // no ju prefix
		"juabz", // z is not a digit letter
		"ju",    // empty body
	}
	for _, s := range bad {
		if _, err := DecodeDecimal(s); err == nil {
			t.Errorf("DecodeDecimal(%q) accepted bad input", s)
		}
	}
}
