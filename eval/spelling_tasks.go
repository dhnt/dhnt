// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package eval

import (
	"github.com/dhnt/dhnt/skills"
)

// mustLineariseDhnt renders the worked example for the spelling prompts. It
// panics on error: the example AST is fixed, so failure is a programmer
// error, and every prompt must carry a guaranteed-parsing example.
func mustLineariseDhnt(s skills.Skill) string {
	line, err := skills.LineariseDhnt(s)
	if err != nil {
		panic("eval: spelling example skill must linearise: " + err.Error())
	}
	return line
}

// spellingExampleLine is the worked Layer 1.5 example embedded verbatim in
// every spelling prompt. It is rendered via LineariseDhnt from a real AST,
// so it parses by construction. The example skill (teto/mimo/runi) shares
// no vocabulary with any task surface form in either spelling.
var spellingExampleLine = mustLineariseDhnt(skills.Skill{
	Name:  "teto",
	Steps: []skills.Step{{Name: "mimo", Primitive: "runi"}},
})

// SpellingExampleLine returns the worked Layer 1.5 example embedded in
// every spelling prompt.
func SpellingExampleLine() string {
	return spellingExampleLine
}

// spellingTeachGrammar is the teaching block shared identically by both
// arms of every prompt: the Layer 1.5 grammar, the worked example, and the
// one-line step form. The task sentence follows it; only that sentence's
// requested surface spelling differs between arms.
func spellingTeachGrammar() string {
	return "Layer 1.5 grammar: a skill is one line of lowercase a-z words " +
		"separated by single spaces. Worked example (a skill named teto " +
		"with one step mimo invoking runi):\n" +
		spellingExampleLine + "\n" +
		"A step has the form `sotepo <name> <primitive>`.\n"
}

// Answer-format sentences. Every authoring task wants the bare skill line;
// only the rewriting task keeps its fenced block plus args line.
const (
	spellingBareAnswer  = "Answer with the bare skill line only, with no code fence and no prose."
	spellingFenceAnswer = "Put the rewritten skill in a ``` fenced block, then an `args:` line " +
		"carrying the flag, path, and quoted text unchanged."
)

// spellingPrompt assembles one arm's prompt: arm/task markers, the shared
// teaching block, the task sentence, and the answer format.
func spellingPrompt(arm, task, sentence, answer string) string {
	return "Spelling arm: " + arm + ".\nTask: " + task + ".\n" +
		spellingTeachGrammar() + sentence + " " + answer
}

// SpellingTasks returns the paired task suite for the spelling eval. Every
// task runs under both arms with the same checks; only the requested surface
// spelling differs. The suite covers the README hypothesis checklist:
// shared expansions (bash/bsh), an unseen nonce word (wug), a homonym case
// (liso as predicate vs step), aaa (canonical as-is), plus the liso/cada
// basics and a rewriting task with exact literal preservation.
//
// Every prompt teaches the Layer 1.5 format first (grammar, worked example,
// step form) so columns (c)/(d) measure spelling, not missing instruction:
// the one recorded live smoke run without that teaching parsed 1/16 and is
// INVALID, not a result.
func SpellingTasks() []SpellingTask {
	listSkill := skills.Skill{
		Name:  "taso",
		Steps: []skills.Step{{Name: "liso", Primitive: "loge"}},
	}
	shellSkill := skills.Skill{
		Name:  "taso",
		Steps: []skills.Step{{Name: "basohe", Primitive: "loge"}},
	}
	unseenSkill := skills.Skill{
		Name:  "taso",
		Steps: []skills.Step{{Name: "wuge", Primitive: "loge"}},
	}
	homonymSkill := skills.Skill{
		Name:     "taso",
		Contract: []skills.Check{{Predicate: "liso"}},
		Steps:    []skills.Step{{Name: "alile", Primitive: "loge"}},
	}
	aaaSkill := skills.Skill{
		Name:  "aaa",
		Steps: []skills.Step{{Name: "alile", Primitive: "loge"}},
	}
	cdSkill := skills.Skill{
		Name:  "taso",
		Steps: []skills.Step{{Name: "cada", Primitive: "loge"}},
	}
	const exact = " (use exactly this spelling)"
	return []SpellingTask{
		{
			Name: "list-ls",
			PromptCanonical: spellingPrompt("canonical", "list-ls",
				"Author a dhnt skill named `taso` with one listing step named `liso`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			PromptConventional: spellingPrompt("conventional", "list-ls",
				"Author a dhnt skill named `taso` with one listing step named `ls`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			Intended:     []string{"liso"},
			Conventional: []string{"ls"},
			Expected:     listSkill,
		},
		{
			Name: "shell-bash",
			PromptCanonical: spellingPrompt("canonical", "shell-bash",
				"Author a dhnt skill named `taso` with one shell step named `basohe`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			PromptConventional: spellingPrompt("conventional", "shell-bash",
				"Author a dhnt skill named `taso` with one shell step named `bash`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			Intended:     []string{"basohe"},
			Conventional: []string{"bash"},
			Expected:     shellSkill,
		},
		{
			// Shared expansion: the contraction `bsh` expands to `basohe`
			// exactly like `bash` does (dhnt.md); no registry picks one.
			Name: "shell-bsh",
			PromptCanonical: spellingPrompt("canonical", "shell-bsh",
				"Author a dhnt skill named `taso` with one shell step named `basohe`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			PromptConventional: spellingPrompt("conventional", "shell-bsh",
				"Author a dhnt skill named `taso` with one shell step named `bsh`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			Intended:     []string{"basohe"},
			Conventional: []string{"bsh"},
			Expected:     shellSkill,
		},
		{
			// Unseen word: `wug` has no glossary entry; the encoder still
			// expands it deterministically to `wuge`.
			Name: "unseen-wug",
			PromptCanonical: spellingPrompt("canonical", "unseen-wug",
				"Author a dhnt skill named `taso` with one step named `wuge`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			PromptConventional: spellingPrompt("conventional", "unseen-wug",
				"Author a dhnt skill named `taso` with one step named `wug`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			Intended:     []string{"wuge"},
			Conventional: []string{"wug"},
			Expected:     unseenSkill,
		},
		{
			// Homonym: the surface form `liso` here is a contract predicate
			// meaning "the listing is sorted" — a distinct sense from the
			// listing step of task list-ls. Same form, context carries
			// the meaning (dhnt.md, "Homonyms and context").
			Name: "homonym-liso",
			PromptCanonical: spellingPrompt("canonical", "homonym-liso",
				"Author a dhnt skill named `taso` whose contract is one ensure clause "+
					"with predicate `liso`"+exact+" (here it means \"the listing is sorted\", not the listing step) "+
					"and one leaf step `alile` invoking primitive `loge`.",
				spellingBareAnswer),
			PromptConventional: spellingPrompt("conventional", "homonym-liso",
				"Author a dhnt skill named `taso` whose contract is one ensure clause "+
					"with predicate `ls`"+exact+" (here it means \"the listing is sorted\", not the listing step) "+
					"and one leaf step `alile` invoking primitive `loge`.",
				spellingBareAnswer),
			Intended:     []string{"liso"},
			Conventional: []string{"ls"},
			Expected:     homonymSkill,
		},
		{
			// Accepted canonical as-is: `aaa` is already three vowel
			// syllables (a/a/a), so both arms emit the same form.
			Name: "aaa-as-is",
			PromptCanonical: spellingPrompt("canonical", "aaa-as-is",
				"Author a dhnt skill named `aaa` (already canonical: three vowel "+
					"syllables a/a/a, accepted as-is) with one leaf step `alile` invoking primitive `loge`.",
				spellingBareAnswer),
			PromptConventional: spellingPrompt("conventional", "aaa-as-is",
				"Author a dhnt skill named `aaa` (already canonical: three vowel "+
					"syllables a/a/a, accepted as-is) with one leaf step `alile` invoking primitive `loge`.",
				spellingBareAnswer),
			Intended:     []string{"aaa"},
			Conventional: []string{"aaa"},
			Expected:     aaaSkill,
		},
		{
			Name: "cd-cada",
			PromptCanonical: spellingPrompt("canonical", "cd-cada",
				"Author a dhnt skill named `taso` with one step named `cada`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			PromptConventional: spellingPrompt("conventional", "cd-cada",
				"Author a dhnt skill named `taso` with one step named `cd`"+exact+" invoking primitive `loge`.",
				spellingBareAnswer),
			Intended:     []string{"cada"},
			Conventional: []string{"cd"},
			Expected:     cdSkill,
		},
		{
			// Rewriting task: re-spell a conventional draft without touching
			// non-vocabulary material. dhnt.md requires that rewriting a
			// word never rewrites flags, paths, or quoted text.
			Name: "preserve-rewrite",
			PromptCanonical: spellingPrompt("canonical", "preserve-rewrite",
				"Rewrite this conventional draft using the spelling `liso`, preserving EXACTLY the flag `-l`, "+
					"the path ./notes/ls, and the quoted text \"My Files\":\n"+
					"sokilili taso sotepo ls loge fini fini",
				spellingFenceAnswer),
			PromptConventional: spellingPrompt("conventional", "preserve-rewrite",
				"Rewrite this conventional draft using the spelling `ls`, preserving EXACTLY the flag `-l`, "+
					"the path ./notes/ls, and the quoted text \"My Files\":\n"+
					"sokilili taso sotepo ls loge fini fini",
				spellingFenceAnswer),
			Intended:     []string{"liso"},
			Conventional: []string{"ls"},
			Expected:     listSkill,
			Preserve:     []string{"-l", "./notes/ls", `"My Files"`},
		},
	}
}
