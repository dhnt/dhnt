// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command eval_convergence runs the E1 cross-model convergence experiment:
// one skill, several model tiers, one contract. It prints each tier's raw
// output and the contract verdict, then the with/without-contract summary.
//
//	go run .                       # hermetic fake tiers (strong/weak)
//	go run . --real gemini,aider   # live agent CLIs as tiers (spends tokens)
package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/dhnt/dhnt/eval"
	"github.com/dhnt/dhnt/skills/tui"
)

func main() {
	real := flag.String("real", "", "comma-separated agent CLIs to use as live tiers")
	prompt := flag.String("prompt", "What is 6 times 7? Reply with just the number.", "the task")
	want := flag.String("want", `\b42\b`, "regex defining a correct answer")
	flag.Parse()

	var tiers []eval.Tier
	if *real != "" {
		for _, name := range strings.Split(*real, ",") {
			name = strings.TrimSpace(name)
			c, ok := tui.Completer(name, "", 120*time.Second)
			if !ok {
				fmt.Printf("unknown agent %q; skipping\n", name)
				continue
			}
			tiers = append(tiers, eval.Tier{Name: name, Model: c})
		}
	} else {
		// hermetic fakes: different outputs, varying correctness.
		tiers = []eval.Tier{
			{Name: "strong", Model: fixed("The answer is 42.")},
			{Name: "weak", Model: fixed("Hmm, I think it's 7.")},
			{Name: "other", Model: fixed("It's 42, of course.")},
		}
	}

	results, err := eval.RunConvergence(*prompt, *want, tiers)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%-14s %-7s %s\n", "TIER", "VALID", "OUTPUT")
	for _, r := range results {
		out := r.Output
		if r.Err != "" {
			out = "ERR: " + r.Err
		}
		fmt.Printf("%-14s %-7v %s\n", r.Tier, r.Valid, oneLine(out, 60))
	}
	s := eval.Summarize(results)
	fmt.Printf("\ntiers=%d  distinct-outputs=%d  accepted: without-contract=%d  with-contract=%d\n",
		s.Tiers, s.DistinctOutputs, s.WithoutContract, s.WithContract)
	fmt.Println("→ outputs vary; a prose skill accepts all; the contract accepts only the correct — same verdict per provider.")
}

func fixed(s string) func(string) (string, error) {
	return func(string) (string, error) { return s, nil }
}

func oneLine(s string, n int) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", " ")
	if len(s) > n {
		s = s[:n] + "…"
	}
	return s
}
