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

	"github.com/dhnt/dhnt/skills"
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

// Both arms are fully green with the S2 parser: exact surface, equivalent,
// parser-accepted, correct, preserving. The conventional arm carries its own
// exact surface and is spelling-equivalent (by design), and the S2 parser
// accepts the rule-derived alternative spellings, so the stub run shows
// the conventional arm 8/8 on every column — the expected result.
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
		// With the S2 parser both arms parse and are correct for every
		// stub task (aaa-as-is included: both arms agree end to end).
		if !conv.Parses || !conv.Correct {
			t.Errorf("conventional %s should parse and be correct with the S2 parser: %+v",
				task.Name, conv)
		}
		if !conv.Pass() {
			t.Errorf("conventional %s should be fully green: %+v", task.Name, conv)
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
	if sum.Conventional.Parses != len(tasks) || sum.Conventional.Correct != len(tasks) {
		t.Errorf("conventional summary should fully parse and be correct with the S2 parser: %+v", sum.Conventional)
	}
	if sum.Conventional.Preserved != len(tasks) {
		t.Errorf("conventional summary should fully preserve literals: %+v", sum.Conventional)
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

// Every spelling prompt teaches the Layer 1.5 format: each prompt carries
// the worked example line (rendered via LineariseDhnt from a real AST), the
// example parses and shares no word with any task surface form, and the two
// arms differ only in the task's surface words (modulo the arm marker).
func TestSpelling_PromptsTeachGrammar(t *testing.T) {
	tasks := SpellingTasks()
	example := SpellingExampleLine()
	if _, err := skills.ParseDhnt(example); err != nil {
		t.Fatalf("worked example does not parse: %v (%q)", err, example)
	}
	exampleID, err := skills.Identity(mustParseDhnt(t, example))
	if err != nil {
		t.Fatalf("worked example has no identity: %v", err)
	}
	for _, task := range tasks {
		wantID, err := skills.Identity(task.Expected)
		if err != nil {
			t.Fatalf("%s: expected AST has no identity: %v", task.Name, err)
		}
		if wantID == exampleID {
			t.Errorf("%s: expected AST equals the worked example; the example must stay unrelated", task.Name)
		}
	}
	forbidden := map[string]bool{}
	for _, task := range tasks {
		for _, w := range append(append([]string{}, task.Intended...), task.Conventional...) {
			forbidden[strings.ToLower(w)] = true
		}
	}
	for _, w := range spellingWords(example) {
		if forbidden[w] {
			t.Errorf("worked example %q shares task word %q", example, w)
		}
	}
	for _, task := range tasks {
		prompts := []string{task.PromptCanonical, task.PromptConventional}
		for i, p := range prompts {
			if !strings.Contains(p, example) {
				t.Errorf("%s arm %d: prompt missing worked example %q", task.Name, i, example)
			}
			if !strings.Contains(p, "A step has the form") {
				t.Errorf("%s arm %d: prompt missing the step-form statement", task.Name, i)
			}
		}
		if task.Name == "preserve-rewrite" {
			for _, p := range prompts {
				if !strings.Contains(p, "fenced block") || !strings.Contains(p, "args:") {
					t.Errorf("%s: rewrite prompt must keep the fence plus args line", task.Name)
				}
				if strings.Contains(p, "no code fence") {
					t.Errorf("%s: rewrite prompt must not ask for the bare skill line", task.Name)
				}
			}
		} else {
			for _, p := range prompts {
				if !strings.Contains(p, "no code fence and no prose") {
					t.Errorf("%s: prompt must ask for the bare skill line", task.Name)
				}
			}
		}
		// The arms differ only in the task's surface words: normalise the
		// arm marker, mask every surface word, and require equality. The
		// marker itself must still name its arm (the stub routes on it).
		if !strings.Contains(task.PromptCanonical, "Spelling arm: canonical.") {
			t.Errorf("%s: canonical prompt missing its arm marker", task.Name)
		}
		if !strings.Contains(task.PromptConventional, "Spelling arm: conventional.") {
			t.Errorf("%s: conventional prompt missing its arm marker", task.Name)
		}
		norm := func(p string) string {
			p = strings.ReplaceAll(p, "Spelling arm: canonical.", "Spelling arm: ARM.")
			p = strings.ReplaceAll(p, "Spelling arm: conventional.", "Spelling arm: ARM.")
			return maskSurfaceWords(p, append(task.Intended, task.Conventional...))
		}
		if norm(task.PromptCanonical) != norm(task.PromptConventional) {
			t.Errorf("%s: arms differ beyond the task surface words:\ncanonical:\n%s\nconventional:\n%s",
				task.Name, task.PromptCanonical, task.PromptConventional)
		}
		if sameSurfaceSet(task.Intended, task.Conventional) {
			// Same surface words: only the arm marker may differ.
			if task.PromptCanonical != task.PromptConventional &&
				strings.ReplaceAll(task.PromptCanonical, "Spelling arm: canonical.", "Spelling arm: ARM.") !=
					strings.ReplaceAll(task.PromptConventional, "Spelling arm: conventional.", "Spelling arm: ARM.") {
				t.Errorf("%s: same surface words but prompts differ beyond the arm marker", task.Name)
			}
		} else if task.PromptCanonical == task.PromptConventional {
			t.Errorf("%s: different surface words but prompts identical", task.Name)
		}
	}
}

func mustParseDhnt(t *testing.T, src string) skills.Skill {
	t.Helper()
	s, err := skills.ParseDhnt(src)
	if err != nil {
		t.Fatalf("ParseDhnt(%q): %v", src, err)
	}
	return s
}

// maskSurfaceWords replaces every maximal ASCII-letter run equal
// (case-insensitively) to a surface word with a placeholder, so two prompts
// that differ only in surface words normalise to the same string.
// Whole-word runs avoid the wug/wuge prefix trap of substring replacement.
func maskSurfaceWords(s string, words []string) string {
	set := map[string]bool{}
	for _, w := range words {
		set[strings.ToLower(w)] = true
	}
	var b strings.Builder
	i := 0
	isLetter := func(c byte) bool {
		return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
	}
	for i < len(s) {
		if !isLetter(s[i]) {
			b.WriteByte(s[i])
			i++
			continue
		}
		j := i
		for j < len(s) && isLetter(s[j]) {
			j++
		}
		if set[strings.ToLower(s[i:j])] {
			b.WriteString("\x00")
		} else {
			b.WriteString(s[i:j])
		}
		i = j
	}
	return b.String()
}

// sameSurfaceSet reports whether two surface-word lists name the same set
// (case-insensitively).
func sameSurfaceSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := map[string]int{}
	for _, w := range a {
		set[strings.ToLower(w)]++
	}
	for _, w := range b {
		k := strings.ToLower(w)
		if set[k] == 0 {
			return false
		}
		set[k]--
	}
	return true
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
