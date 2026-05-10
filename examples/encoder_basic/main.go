// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command encoder_basic demonstrates the dhnt language primitives:
// the vowel-insertion rule applied to a-z input and the ju-prefixed
// decimal numeral encoding.
//
// Run with:
//
//	go run github.com/dhnt/dhnt/examples/encoder_basic
package main

import (
	"fmt"
	"log"

	"github.com/dhnt/dhnt"
)

func main() {
	// Encode each English word into its canonical dhnt full form.
	for _, w := range []string{"hello", "world", "git", "github", "step", "skill"} {
		out, err := dhnt.EncodeWord(w)
		if err != nil {
			log.Fatalf("EncodeWord(%q): %v", w, err)
		}
		fmt.Printf("%-10s → %s\n", w, out)
	}

	// EncodePhrase applies the rule to each whitespace-separated word.
	phrase, err := dhnt.EncodePhrase("how are you today")
	if err != nil {
		log.Fatalf("EncodePhrase: %v", err)
	}
	fmt.Printf("\nphrase: %s\n", phrase)

	// Numerals roundtrip cleanly through the ju-prefixed decimal form.
	for _, n := range []uint64{0, 1, 9, 12, 2018, 1234567} {
		enc := dhnt.EncodeDecimal(n)
		back, err := dhnt.DecodeDecimal(enc)
		if err != nil {
			log.Fatalf("DecodeDecimal(%q): %v", enc, err)
		}
		fmt.Printf("%-10d → %-15s → %d\n", n, enc, back)
	}
}
