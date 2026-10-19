// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package eval

// Spelling eval — canonical vs conventional spelling (Sprint 476, story S1).
//
// Hypothesis under test (README.md, "Familiar spelling and LLMs"): by design
// every English word is already an accepted dhnt form as-is (lowercased), so
// conventional spellings (ls, bash, cd) are a valid authoring surface for
// LLMs, while canonical spellings (liso, basohe, cada) are the machine form.
// The paired design runs the SAME skill-authoring/rewriting task with the
// SAME checks under two arms — arm A prompted for canonical spellings, arm B
// for conventional ones — and scores each arm separately on four columns:
//
//	(a) surface: the output contains the exact words the arm asked for —
//	    arm A the canonical forms, arm B the conventional forms given in
//	    the prompt (lowercased exact-word match, so arm A answering "ls"
//	    fails surface while still passing equivalence below);
//	(b) equivalent: every intended canonical form is covered by an output
//	    word with an equal EncodeWord expansion (x ~ y iff N(x)=N(y));
//	(c) parser acceptance: the dhnt candidate parses via skills.ParseDhnt;
//	(d) semantic correctness: the parsed AST has the task's expected Identity;
//	(e) preservation: arguments, flags, paths and quoted text survive exactly.
//
// Column (c) is a MEASURED column, not a dependency: story S2 made
// ParseDhnt accept rule-derived conventional spellings on input, so with
// the S2 parser arm B reads Parses=true/Correct=true on equivalent input.
// Column (c) — like column (b) — stays measured: both arms are expected to
// pass every column on the stub transcripts, and the harness keeps scoring
// each column separately.
//
// Naming note: Sprint 476 tracks this as "E5". In this package the paper's
// E5 (interop, RunInterop in experiments.go) already holds that slot, so the
// spelling eval lives under Spelling* names to avoid the collision.

import (
	"strconv"
	"strings"

	"github.com/dhnt/dhnt"
	"github.com/dhnt/dhnt/skills"
)

// Arm is one side of a paired spelling task.
type Arm string

const (
	// ArmCanonical prompts for and expects canonical dhnt spellings
	// (liso, basohe, cada).
	ArmCanonical Arm = "canonical"
	// ArmConventional prompts for and expects conventional spellings
	// (ls, bash, cd) — the as-is authoring surface.
	ArmConventional Arm = "conventional"
)

// SpellingTask is one paired skill-authoring/rewriting task. Both arms
// describe the same task and are judged against the same checks; only the
// requested surface spelling differs.
type SpellingTask struct {
	Name string
	// PromptCanonical asks for canonical dhnt spellings.
	PromptCanonical string
	// PromptConventional asks for conventional spellings.
	PromptConventional string
	// Intended lists the canonical dhnt forms the output must cover, one
	// output word each, under EncodeWord equivalence. It is also arm A's
	// exact surface expectation.
	Intended []string
	// Conventional lists the exact conventional words arm B must emit
	// (the forms named in PromptConventional, e.g. ls, bash, bsh, cd).
	Conventional []string
	// Expected is the AST the parsed output must match (via Identity).
	Expected skills.Skill
	// Preserve lists literals (flags, paths, quoted text) that must
	// survive verbatim in the raw output, per dhnt.md: rewriting a word
	// never rewrites -l, paths, or quoted text.
	Preserve []string
}

// PromptFor returns the arm-specific prompt for a task.
func PromptFor(t SpellingTask, arm Arm) string {
	if arm == ArmConventional {
		return t.PromptConventional
	}
	return t.PromptCanonical
}

// SpellingVerdict scores one arm of one task.
type SpellingVerdict struct {
	Task       string
	Arm        Arm
	Output     string
	Surface    bool // (a) output carries the exact words the arm asked for
	Equivalent bool // (b) every intended form covered via EncodeWord equivalence
	Parses     bool // (c) dhnt candidate accepted by skills.ParseDhnt
	Correct    bool // (d) parsed AST Identity equals the task's expected AST
	Preserved  bool // (e) every Preserve literal present verbatim
	ParseErr   string
	ModelErr   string
}

// Pass reports whether the arm satisfied every measured column.
func (v SpellingVerdict) Pass() bool {
	return v.Surface && v.Equivalent && v.Parses && v.Correct && v.Preserved
}

// SpellingArmSummary aggregates one arm across tasks.
type SpellingArmSummary struct {
	Tasks      int
	Surface    int // (a) outputs carrying the arm's exact surface words
	Equivalent int // (b) spelling-equivalent outputs
	Parses     int // (c) parser-accepted outputs
	Correct    int // (d) semantically correct outputs
	Preserved  int // (e) outputs preserving every literal
}

// SpellingSummary aggregates both arms of a paired run.
type SpellingSummary struct {
	Canonical    SpellingArmSummary
	Conventional SpellingArmSummary
}

func (s *SpellingSummary) add(arm Arm, v SpellingVerdict) {
	dst := &s.Canonical
	if arm == ArmConventional {
		dst = &s.Conventional
	}
	dst.Tasks++
	if v.Surface {
		dst.Surface++
	}
	if v.Equivalent {
		dst.Equivalent++
	}
	if v.Parses {
		dst.Parses++
	}
	if v.Correct {
		dst.Correct++
	}
	if v.Preserved {
		dst.Preserved++
	}
}

// CheckSurfaceSpelling reports whether the output contains each expected
// surface word exactly: expected words are matched against the output's
// maximal lowercase ASCII-letter runs, so arm A (expecting "liso")
// fails when the model writes "ls", and arm B (expecting "ls") fails
// when the model writes "liso". Matching is case-insensitive ("LISO"
// satisfies surface "liso"); letter-runs containing more or fewer letters
// do not match.
func CheckSurfaceSpelling(output string, expected []string) bool {
	have := map[string]bool{}
	for _, w := range spellingWords(output) {
		have[w] = true
	}
	for _, e := range expected {
		if !have[strings.ToLower(e)] {
			return false
		}
	}
	return true
}

// SurfaceFor returns the exact surface words an arm must emit for a task:
// the canonical forms for arm A, the conventional forms for arm B.
func SurfaceFor(t SpellingTask, arm Arm) []string {
	if arm == ArmConventional {
		return t.Conventional
	}
	return t.Intended
}

// CheckSpellingEquivalence reports whether the output covers every intended
// canonical form: for each intended form c there is an output word w with
// EncodeWord(lower(w)) == EncodeWord(lower(c)). Output words are maximal
// ASCII-letter runs, lowercased first, so "LS" covers "liso" and the
// contraction "bsh" covers "basohe". Non-letter material (flags, paths,
// quoted text) carries no vocabulary words and is ignored here — it is
// column (e)'s job.
func CheckSpellingEquivalence(output string, intended []string) bool {
	have := map[string]bool{}
	for _, w := range spellingWords(output) {
		enc, err := dhnt.EncodeWord(w)
		if err != nil {
			continue
		}
		have[enc] = true
	}
	for _, c := range intended {
		enc, err := dhnt.EncodeWord(strings.ToLower(c))
		if err != nil || !have[enc] {
			return false
		}
	}
	return true
}

// CheckSpellingCompliance is a kept alias of CheckSpellingEquivalence.
func CheckSpellingCompliance(output string, intended []string) bool {
	return CheckSpellingEquivalence(output, intended)
}

// spellingWords splits text into maximal lowercase ASCII-letter runs.
func spellingWords(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return r < 'a' || r > 'z'
	})
}

// ExtractDhntCandidate returns the text column (b) parses: the first fenced
// block when the output carries one (the rewriting task emits dhnt plus a
// verbatim args line), otherwise the whole output, trimmed.
func ExtractDhntCandidate(output string) string {
	rest := output
	if i := strings.Index(rest, "```"); i >= 0 {
		rest = rest[i+3:]
		// Drop an optional info string on the opening fence line.
		if j := strings.Index(rest, "\n"); j >= 0 {
			rest = rest[j+1:]
		}
		if j := strings.Index(rest, "```"); j >= 0 {
			rest = rest[:j]
		}
	}
	return strings.TrimSpace(rest)
}

// CheckPreservation reports whether every literal survives verbatim
// (byte-exact, case-sensitive) in the raw output. Vacuously true when the
// task preserves nothing.
func CheckPreservation(output string, literals []string) bool {
	for _, lit := range literals {
		if !strings.Contains(output, lit) {
			return false
		}
	}
	return true
}

// scoreSpellingArm scores one model output against one task.
func scoreSpellingArm(taskName string, arm Arm, output string, t SpellingTask, wantID string) SpellingVerdict {
	v := SpellingVerdict{Task: taskName, Arm: arm, Output: output}
	v.Surface = CheckSurfaceSpelling(output, SurfaceFor(t, arm))
	v.Equivalent = CheckSpellingEquivalence(output, t.Intended)
	v.Preserved = CheckPreservation(output, t.Preserve)
	got, err := skills.ParseDhnt(ExtractDhntCandidate(output))
	if err != nil {
		v.ParseErr = err.Error()
		return v
	}
	v.Parses = true
	gotID, err := skills.Identity(got)
	if err != nil {
		v.ParseErr = err.Error()
		v.Parses = false
		return v
	}
	v.Correct = gotID == wantID
	return v
}

// RunSpelling runs every task under both arms against one model function
// (prompt -> text): the same task, the same checks, only the requested
// surface spelling differs. A model error is recorded on that arm's verdict,
// not fatal to the run; an unlinearisable Expected AST is a programmer
// error and fails the run.
func RunSpelling(model skills.Completer, tasks []SpellingTask) ([]SpellingVerdict, SpellingSummary, error) {
	want := make([]string, len(tasks))
	for i, t := range tasks {
		id, err := skills.Identity(t.Expected)
		if err != nil {
			return nil, SpellingSummary{}, err
		}
		want[i] = id
	}
	var verdicts []SpellingVerdict
	var sum SpellingSummary
	for i, t := range tasks {
		for _, arm := range []Arm{ArmCanonical, ArmConventional} {
			out, err := model(PromptFor(t, arm))
			if err != nil {
				v := SpellingVerdict{Task: t.Name, Arm: arm, ModelErr: err.Error()}
				verdicts = append(verdicts, v)
				sum.add(arm, v)
				continue
			}
			v := scoreSpellingArm(t.Name, arm, out, t, want[i])
			verdicts = append(verdicts, v)
			sum.add(arm, v)
		}
	}
	return verdicts, sum, nil
}

// RenderSpellingMarkdown renders verdicts and the summary as a Markdown
// table for the paper and the live smoke run.
func RenderSpellingMarkdown(verdicts []SpellingVerdict, sum SpellingSummary) string {
	var b strings.Builder
	b.WriteString("| task | arm | surface | equivalent | parses | correct | preserved |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	yn := func(ok bool) string {
		if ok {
			return "yes"
		}
		return "no"
	}
	for _, v := range verdicts {
		b.WriteString("| " + v.Task + " | " + string(v.Arm) + " | " +
			yn(v.Surface) + " | " + yn(v.Equivalent) + " | " +
			yn(v.Parses) + " | " + yn(v.Correct) + " | " +
			yn(v.Preserved) + " |\n")
	}
	b.WriteString("\n")
	b.WriteString("canonical: " + armLine(sum.Canonical) + "\n")
	b.WriteString("conventional: " + armLine(sum.Conventional) + "\n")
	return b.String()
}

func armLine(s SpellingArmSummary) string {
	return strconv.Itoa(s.Surface) + "/" + strconv.Itoa(s.Tasks) + " surface, " +
		strconv.Itoa(s.Equivalent) + "/" + strconv.Itoa(s.Tasks) + " equivalent, " +
		strconv.Itoa(s.Parses) + "/" + strconv.Itoa(s.Tasks) + " parse, " +
		strconv.Itoa(s.Correct) + "/" + strconv.Itoa(s.Tasks) + " correct, " +
		strconv.Itoa(s.Preserved) + "/" + strconv.Itoa(s.Tasks) + " preserved"
}
