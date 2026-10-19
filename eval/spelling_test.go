// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package eval

import (
	"strings"
	"testing"
)

// Spelling equivalence is defined by equal EncodeWord expansions
// (dhnt.md): ls/liso, bash/bsh/basohe, cd/cada, wug/wuge, AAA/aaa.
func TestSpelling_Equivalence(t *testing.T) {
	cases := []struct {
		output   string
		intended []string
		want     bool
	}{
		{"sotepo ls loge", []string{"liso"}, true},
		{"sotepo liso loge", []string{"liso"}, true},
		{"sotepo LS loge", []string{"liso"}, true}, // as-is, lowercased
		{"sotepo bash loge", []string{"basohe"}, true},
		{"sotepo bsh loge", []string{"basohe"}, true}, // shared expansion
		{"sotepo basohe loge", []string{"basohe"}, true},
		{"sotepo wug loge", []string{"wuge"}, true}, // unseen word
		{"sokilili AAA sotepo alile loge fini fini", []string{"aaa"}, true},
		{"sotepo cd loge", []string{"cada"}, true},
		{"sotepo liso loge", []string{"cada"}, false},
		{"sotepo loge fini", []string{"liso"}, false}, // intended form absent
	}
	for _, tc := range cases {
		if got := CheckSpellingEquivalence(tc.output, tc.intended); got != tc.want {
			t.Errorf("CheckSpellingEquivalence(%q, %v) = %v, want %v",
				tc.output, tc.intended, got, tc.want)
		}
	}
}

// Surface is arm-specific and exact: arm A (canonical) fails when the model
// writes the conventional form, even though equivalence still holds.
func TestSpelling_Surface(t *testing.T) {
	cases := []struct {
		output   string
		expected []string
		want     bool
	}{
		{"sotepo liso loge", []string{"liso"}, true},
		{"sotepo ls loge", []string{"liso"}, false}, // wrong surface, still equivalent
		{"sotepo ls loge", []string{"ls"}, true},
		{"sotepo liso loge", []string{"ls"}, false},
		{"sotepo LISO loge", []string{"liso"}, true}, // case-insensitive exact
		{"sotepo bsh loge", []string{"bsh"}, true},
		{"sotepo bash loge", []string{"bsh"}, false}, // shared expansion, other surface
		{"sokilili AAA sotepo alile loge fini fini", []string{"aaa"}, true},
		{"sotepo loge fini", []string{"liso"}, false},
	}
	for _, tc := range cases {
		if got := CheckSurfaceSpelling(tc.output, tc.expected); got != tc.want {
			t.Errorf("CheckSurfaceSpelling(%q, %v) = %v, want %v",
				tc.output, tc.expected, got, tc.want)
		}
	}
}

// The steering case: arm A asked for liso answers with ls — surface fails,
// equivalence passes. The columns tell the arms apart.
func TestSpelling_SurfaceVsEquivalent(t *testing.T) {
	out := "sokilili taso sotepo ls loge fini fini"
	if !CheckSpellingEquivalence(out, []string{"liso"}) {
		t.Fatalf("ls should be equivalent to liso")
	}
	if CheckSurfaceSpelling(out, []string{"liso"}) {
		t.Fatalf("ls should fail canonical surface liso")
	}
	if !CheckSurfaceSpelling(out, []string{"ls"}) {
		t.Fatalf("ls should pass conventional surface ls")
	}
}

func TestSpelling_Preservation(t *testing.T) {
	out := "```\nsokilili taso sotepo liso loge fini fini\n```\nargs: -l ./notes/ls \"My Files\""
	lits := []string{"-l", "./notes/ls", `"My Files"`}
	if !CheckPreservation(out, lits) {
		t.Errorf("exact literals not recognised as preserved")
	}
	if CheckPreservation(out, []string{"-L"}) {
		t.Errorf("case-sensitive flag mismatch accepted as preserved")
	}
	if CheckPreservation(out, []string{"./notes/Ls"}) {
		t.Errorf("rewritten path accepted as preserved")
	}
	if !CheckPreservation(out, nil) {
		t.Errorf("empty literal set should be vacuously preserved")
	}
}

func TestSpelling_ExtractCandidate(t *testing.T) {
	bare := "sokilili taso sotepo liso loge fini fini"
	if got := ExtractDhntCandidate(bare); got != bare {
		t.Errorf("bare output candidate = %q, want whole output", got)
	}
	fenced := "rewritten skill:\n```\nsokilili taso sotepo liso loge fini fini\n```\nargs: -l"
	if got := ExtractDhntCandidate(fenced); got != bare {
		t.Errorf("fenced candidate = %q, want %q", got, bare)
	}
}

// stubSpellingModel replays recorded transcripts keyed by the arm and task
// marker lines in each prompt. No network, fully deterministic.
func stubSpellingModel(t *testing.T) func(string) (string, error) {
	t.Helper()
	canonical := map[string]string{
		"list-ls":          "sokilili taso sotepo liso loge fini fini",
		"shell-bash":       "sokilili taso sotepo basohe loge fini fini",
		"shell-bsh":        "sokilili taso sotepo basohe loge fini fini",
		"unseen-wug":       "sokilili taso sotepo wuge loge fini fini",
		"homonym-liso":     "sokilili taso enisure liso fini sotepo alile loge fini fini",
		"aaa-as-is":        "sokilili aaa sotepo alile loge fini fini",
		"cd-cada":          "sokilili taso sotepo cada loge fini fini",
		"preserve-rewrite": "```\nsokilili taso sotepo liso loge fini fini\n```\nargs: -l ./notes/ls \"My Files\"",
	}
	conventional := map[string]string{
		"list-ls":          "sokilili taso sotepo ls loge fini fini",
		"shell-bash":       "sokilili taso sotepo bash loge fini fini",
		"shell-bsh":        "sokilili taso sotepo bsh loge fini fini",
		"unseen-wug":       "sokilili taso sotepo wug loge fini fini",
		"homonym-liso":     "sokilili taso enisure ls fini sotepo alile loge fini fini",
		"aaa-as-is":        "sokilili aaa sotepo alile loge fini fini",
		"cd-cada":          "sokilili taso sotepo cd loge fini fini",
		"preserve-rewrite": "```\nsokilili taso sotepo ls loge fini fini\n```\nargs: -l ./notes/ls \"My Files\"",
	}
	return func(prompt string) (string, error) {
		var table map[string]string
		var arm string
		switch {
		case strings.Contains(prompt, "Spelling arm: canonical."):
			table, arm = canonical, "canonical"
		case strings.Contains(prompt, "Spelling arm: conventional."):
			table, arm = conventional, "conventional"
		default:
			t.Fatalf("prompt carries no arm marker:\n%s", prompt)
		}
		for task, out := range table {
			if strings.Contains(prompt, "Task: "+task+".") {
				return out, nil
			}
		}
		t.Fatalf("prompt carries no known task marker (%s arm):\n%s", arm, prompt)
		return "", nil
	}
}

// The canonical arm is fully green: exact surface, equivalent,
// parser-accepted, correct, preserving. The conventional arm carries its own
// exact surface and is spelling-equivalent (by design) but today's parser
// rejects it — the measured gap S2 will close.
func TestSpelling_PairedStubRun(t *testing.T) {
	tasks := SpellingTasks()
	verdicts, sum, err := RunSpelling(stubSpellingModel(t), tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(verdicts) != 2*len(tasks) {
		t.Fatalf("got %d verdicts, want %d", len(verdicts), 2*len(tasks))
	}
	byKey := map[string]SpellingVerdict{}
	for _, v := range verdicts {
		byKey[v.Task+"/"+string(v.Arm)] = v
	}
	for _, task := range tasks {
		canon := byKey[task.Name+"/"+string(ArmCanonical)]
		if !canon.Pass() {
			t.Errorf("canonical %s not fully green: %+v parse-err=%q",
				task.Name, canon, canon.ParseErr)
		}
		conv := byKey[task.Name+"/"+string(ArmConventional)]
		if !conv.Surface {
			t.Errorf("conventional %s should carry its exact surface: %+v", task.Name, conv)
		}
		if !conv.Equivalent {
			t.Errorf("conventional %s should be spelling-equivalent: %+v", task.Name, conv)
		}
		if !conv.Preserved {
			t.Errorf("conventional %s should preserve literals: %+v", task.Name, conv)
		}
		if task.Name == "aaa-as-is" {
			// Already-canonical form: both arms agree end to end.
			if !conv.Pass() {
				t.Errorf("conventional aaa-as-is should be fully green: %+v", conv)
			}
			continue
		}
		if conv.Parses || conv.Correct {
			t.Errorf("conventional %s unexpectedly accepted by today's parser: %+v",
				task.Name, conv)
		}
	}
	if sum.Canonical.Tasks != len(tasks) || sum.Conventional.Tasks != len(tasks) {
		t.Errorf("summary task counts wrong: %+v", sum)
	}
	if sum.Canonical.Surface != len(tasks) || sum.Canonical.Equivalent != len(tasks) ||
		sum.Canonical.Parses != len(tasks) ||
		sum.Canonical.Correct != len(tasks) || sum.Canonical.Preserved != len(tasks) {
		t.Errorf("canonical summary should be perfect: %+v", sum.Canonical)
	}
	if sum.Conventional.Surface != len(tasks) {
		t.Errorf("conventional summary should be fully on-surface: %+v", sum.Conventional)
	}
	if sum.Conventional.Equivalent != len(tasks) {
		t.Errorf("conventional summary should be fully equivalent: %+v", sum.Conventional)
	}
	if sum.Conventional.Parses != 1 || sum.Conventional.Correct != 1 {
		t.Errorf("only aaa-as-is should parse today: %+v", sum.Conventional)
	}
	t.Logf("\n%s", RenderSpellingMarkdown(verdicts, sum))
}

// A canonical-arm answer in the conventional surface (ls for liso) fails
// surface but passes equivalence — and so fails Pass. This is the column
// split the rework asked for: the arms are told apart even when the
// encoder accepts both.
func TestSpelling_WrongSurfaceArmA(t *testing.T) {
	tasks := SpellingTasks()[:1] // list-ls: canonical liso, conventional ls
	model := func(prompt string) (string, error) {
		return "sokilili taso sotepo ls loge fini fini", nil
	}
	verdicts, _, err := RunSpelling(model, tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(verdicts) != 2 {
		t.Fatalf("got %d verdicts, want 2", len(verdicts))
	}
	canon, conv := verdicts[0], verdicts[1]
	if canon.Arm != ArmCanonical || conv.Arm != ArmConventional {
		t.Fatalf("unexpected arm order: %+v", verdicts)
	}
	if canon.Surface {
		t.Errorf("canonical arm answering ls should fail surface: %+v", canon)
	}
	if !canon.Equivalent {
		t.Errorf("canonical arm answering ls should pass equivalence: %+v", canon)
	}
	if canon.Pass() {
		t.Errorf("canonical arm answering ls should not pass overall: %+v", canon)
	}
	if !conv.Surface || !conv.Equivalent {
		t.Errorf("conventional arm answering ls should be on-surface and equivalent: %+v", conv)
	}
}

// A model failure is recorded on that arm's verdict; the paired run still
// completes the remaining arms and tasks.
func TestSpelling_ModelError(t *testing.T) {
	tasks := SpellingTasks()[:1]
	model := func(prompt string) (string, error) {
		if strings.Contains(prompt, "Spelling arm: conventional.") {
			return "", errFail
		}
		return "sokilili taso sotepo liso loge fini fini", nil
	}
	verdicts, sum, err := RunSpelling(model, tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(verdicts) != 2 {
		t.Fatalf("got %d verdicts, want 2", len(verdicts))
	}
	if verdicts[0].ModelErr != "" || !verdicts[0].Pass() {
		t.Errorf("canonical verdict wrong: %+v", verdicts[0])
	}
	if verdicts[1].ModelErr == "" || verdicts[1].Pass() {
		t.Errorf("conventional error not recorded: %+v", verdicts[1])
	}
	if sum.Canonical.Correct != 1 || sum.Conventional.Correct != 0 {
		t.Errorf("summary wrong after model error: %+v", sum)
	}
}
