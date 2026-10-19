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

// SpellingTasks returns the paired task suite for the spelling eval. Every
// task runs under both arms with the same checks; only the requested surface
// spelling differs. The suite covers the README hypothesis checklist:
// shared expansions (bash/bsh), an unseen nonce word (wug), a homonym case
// (liso as predicate vs step), aaa (canonical as-is), plus the liso/cada
// basics and a rewriting task with exact literal preservation.
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
	const shape = "Emit only the skill in Layer 1.5 form " +
		"(lowercase a-z words separated by single spaces), no explanation."
	return []SpellingTask{
		{
			Name: "list-ls",
			PromptCanonical: "Spelling arm: canonical.\nTask: list-ls.\n" +
				"Author a dhnt skill named `taso` with one listing step named `liso` " +
				"(canonical spelling) invoking primitive `loge`. " + shape,
			PromptConventional: "Spelling arm: conventional.\nTask: list-ls.\n" +
				"Author a dhnt skill named `taso` with one listing step named `ls` " +
				"(conventional spelling, accepted as-is) invoking primitive `loge`. " + shape,
			Intended:     []string{"liso"},
			Conventional: []string{"ls"},
			Expected:     listSkill,
		},
		{
			Name: "shell-bash",
			PromptCanonical: "Spelling arm: canonical.\nTask: shell-bash.\n" +
				"Author a dhnt skill named `taso` with one shell step named `basohe` " +
				"(canonical spelling) invoking primitive `loge`. " + shape,
			PromptConventional: "Spelling arm: conventional.\nTask: shell-bash.\n" +
				"Author a dhnt skill named `taso` with one shell step named `bash` " +
				"(conventional spelling, accepted as-is) invoking primitive `loge`. " + shape,
			Intended:     []string{"basohe"},
			Conventional: []string{"bash"},
			Expected:     shellSkill,
		},
		{
			// Shared expansion: the contraction `bsh` expands to `basohe`
			// exactly like `bash` does (dhnt.md); no registry picks one.
			Name: "shell-bsh",
			PromptCanonical: "Spelling arm: canonical.\nTask: shell-bsh.\n" +
				"Author a dhnt skill named `taso` with one shell step named `basohe` " +
				"(canonical spelling) invoking primitive `loge`. " + shape,
			PromptConventional: "Spelling arm: conventional.\nTask: shell-bsh.\n" +
				"Author a dhnt skill named `taso` with one shell step named `bsh` " +
				"(conventional contraction, accepted as-is) invoking primitive `loge`. " + shape,
			Intended:     []string{"basohe"},
			Conventional: []string{"bsh"},
			Expected:     shellSkill,
		},
		{
			// Unseen word: `wug` has no glossary entry; the encoder still
			// expands it deterministically to `wuge`.
			Name: "unseen-wug",
			PromptCanonical: "Spelling arm: canonical.\nTask: unseen-wug.\n" +
				"Author a dhnt skill named `taso` with one step named `wuge` " +
				"(canonical spelling of a new nonce word) invoking primitive `loge`. " + shape,
			PromptConventional: "Spelling arm: conventional.\nTask: unseen-wug.\n" +
				"Author a dhnt skill named `taso` with one step named `wug` " +
				"(the nonce word as-is) invoking primitive `loge`. " + shape,
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
			PromptCanonical: "Spelling arm: canonical.\nTask: homonym-liso.\n" +
				"Author a dhnt skill named `taso` whose contract is one ensure clause " +
				"with predicate `liso` (canonical spelling; here it means " +
				"\"the listing is sorted\", not the listing step) and one leaf step " +
				"`alile` invoking primitive `loge`. " + shape,
			PromptConventional: "Spelling arm: conventional.\nTask: homonym-liso.\n" +
				"Author a dhnt skill named `taso` whose contract is one ensure clause " +
				"with predicate `ls` (conventional spelling; here it means " +
				"\"the listing is sorted\", not the listing step) and one leaf step " +
				"`alile` invoking primitive `loge`. " + shape,
			Intended:     []string{"liso"},
			Conventional: []string{"ls"},
			Expected:     homonymSkill,
		},
		{
			// Accepted canonical as-is: `aaa` is already three vowel
			// syllables (a/a/a), so both arms emit the same form.
			Name: "aaa-as-is",
			PromptCanonical: "Spelling arm: canonical.\nTask: aaa-as-is.\n" +
				"Author a dhnt skill named `aaa` (already canonical: three vowel " +
				"syllables a/a/a, accepted as-is) with one leaf step `alile` " +
				"invoking primitive `loge`. " + shape,
			PromptConventional: "Spelling arm: conventional.\nTask: aaa-as-is.\n" +
				"Author a dhnt skill named `aaa` (already canonical: three vowel " +
				"syllables a/a/a, accepted as-is) with one leaf step `alile` " +
				"invoking primitive `loge`. " + shape,
			Intended:     []string{"aaa"},
			Conventional: []string{"aaa"},
			Expected:     aaaSkill,
		},
		{
			Name: "cd-cada",
			PromptCanonical: "Spelling arm: canonical.\nTask: cd-cada.\n" +
				"Author a dhnt skill named `taso` with one step named `cada` " +
				"(canonical spelling) invoking primitive `loge`. " + shape,
			PromptConventional: "Spelling arm: conventional.\nTask: cd-cada.\n" +
				"Author a dhnt skill named `taso` with one step named `cd` " +
				"(conventional spelling, accepted as-is) invoking primitive `loge`. " + shape,
			Intended:     []string{"cada"},
			Conventional: []string{"cd"},
			Expected:     cdSkill,
		},
		{
			// Rewriting task: re-spell a conventional draft without touching
			// non-vocabulary material. dhnt.md requires that rewriting a
			// word never rewrites flags, paths, or quoted text.
			Name: "preserve-rewrite",
			PromptCanonical: "Spelling arm: canonical.\nTask: preserve-rewrite.\n" +
				"Rewrite this conventional draft into canonical dhnt spellings " +
				"(`ls` becomes `liso`), preserving EXACTLY the flag `-l`, the path " +
				"./notes/ls, and the quoted text \"My Files\":\n" +
				"sokilili taso sotepo ls loge fini fini\n" +
				"Put the rewritten skill in a ``` fenced block, then an `args:` line " +
				"carrying the flag, path, and quoted text unchanged.",
			PromptConventional: "Spelling arm: conventional.\nTask: preserve-rewrite.\n" +
				"Rewrite this conventional draft keeping conventional spellings " +
				"(`ls` stays `ls`), preserving EXACTLY the flag `-l`, the path " +
				"./notes/ls, and the quoted text \"My Files\":\n" +
				"sokilili taso sotepo ls loge fini fini\n" +
				"Put the rewritten skill in a ``` fenced block, then an `args:` line " +
				"carrying the flag, path, and quoted text unchanged.",
			Intended:     []string{"liso"},
			Conventional: []string{"ls"},
			Expected:     listSkill,
			Preserve:     []string{"-l", "./notes/ls", `"My Files"`},
		},
	}
}
