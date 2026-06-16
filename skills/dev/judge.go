// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package dev

import (
	"fmt"
	"strings"

	"github.com/dhnt/dhnt/skills"
)

// PredJudge is the model-judged goal-met predicate (= dhnt.EncodeWord
// ("met")). Unlike PredExit (a deterministic exit code), it gathers
// evidence by running its referenced command and asks a Spec.Judge
// completer whether the goal was achieved — the contract clause for
// outcomes that have no clean pass/fail verifier (e.g. "the docs now
// explain X", "the refactor is cleaner"). It is bound only when
// Spec.Judge is set, so the deterministic leaf is unchanged by default.
const PredJudge = "meto"

// judged runs the referenced evidence command and asks the Spec.Judge
// completer whether Spec.Goal has been achieved, given that evidence. It
// is conservative: any answer other than an explicit YES is false. The
// effects reflect a model call (read evidence, reach the network, spend
// tokens, burn time).
func (s *Session) judged(args []skills.Arg) (bool, []skills.Effect, error) {
	effects := []skills.Effect{skills.EffRead, skills.EffNet, skills.EffSpend, skills.EffTime}
	if s.spec.Judge == nil {
		return false, nil, fmt.Errorf("dev: judged predicate needs a Spec.Judge completer")
	}
	ref, err := argRef(args)
	if err != nil {
		return false, nil, err
	}
	evidence := s.exec(ref).out
	out, err := s.spec.Judge(judgePrompt(s.spec.Goal, evidence))
	if err != nil {
		return false, effects, fmt.Errorf("dev: judge: %w", err)
	}
	return parseVerdict(out), effects, nil
}

// judgePrompt asks the model for a strict YES/NO verdict on whether goal
// is achieved, citing only the supplied evidence and defaulting to NO.
func judgePrompt(goal, evidence string) string {
	goal = strings.TrimSpace(goal)
	if goal == "" {
		goal = "(no goal stated)"
	}
	if strings.TrimSpace(evidence) == "" {
		evidence = "(no evidence captured)"
	}
	return "You are a strict verifier. Decide whether the GOAL has been FULLY achieved, " +
		"using ONLY the EVIDENCE below — do not assume work not shown.\n\n" +
		"GOAL:\n" + goal + "\n\nEVIDENCE:\n" + evidence + "\n\n" +
		"Reply with exactly one tag and nothing else: <verdict>YES</verdict> if the goal is " +
		"fully achieved, or <verdict>NO</verdict> otherwise (default to NO when unsure)."
}

// parseVerdict extracts the model's YES/NO judgement, tolerating chatter
// and code fences. It accepts an explicit <verdict>…</verdict> tag or a
// bare one-word reply; anything not clearly YES is false.
func parseVerdict(out string) bool {
	lo := strings.ToLower(out)
	if i := strings.Index(lo, "<verdict>"); i >= 0 {
		if j := strings.Index(lo[i:], "</verdict>"); j >= 0 {
			return strings.TrimSpace(lo[i+len("<verdict>"):i+j]) == "yes"
		}
	}
	return strings.TrimSpace(lo) == "yes"
}
