// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command tui_drive is an applied demo: one tool-agnostic dhnt skill
// drives a real terminal program over a pseudo-terminal (the expect(1)
// idiom), and the contract verifies the interaction succeeded.
//
// The same skill drives any TUI — cat, sh, or an agentic CLI (claude /
// codex / gemini / aider / opencode) — by swapping the Spec's argv,
// patterns, and inputs (see agentCLISpecExample below).
//
//	go run .
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

// driveSkill: spawn → expect READY → send INPUT → expect RESULT → quit,
// with the contract "the tool exited cleanly". Tool-agnostic.
func driveSkill() skills.Skill {
	ref := func(k string) []skills.Arg {
		return []skills.Arg{{Name: "value", Value: skills.NewRef(k)}}
	}
	return skills.Skill{
		Name:      "darive",
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffTime},
		Steps: []skills.Step{
			{Name: "sa", Primitive: tui.PrimSpawn},
			{Name: "se", Primitive: tui.PrimExpect, Args: ref("readayu")},
			{Name: "si", Primitive: tui.PrimSend, Args: ref("iniputo")},
			{Name: "so", Primitive: tui.PrimExpect, Args: ref("resulito")},
			{Name: "su", Primitive: tui.PrimQuit},
		},
		Contract: []skills.Check{{Predicate: tui.PredClean}},
	}
}

func main() {
	// Spec for /bin/sh: run `echo READY`, confirm the output, then exit.
	shSpec := tui.Spec{
		Argv:     []string{"sh"},
		Patterns: map[string]string{"readayu": "", "resulito": "READY"},
		Inputs:   map[string]string{"iniputo": "echo READY\n"},
		Quit:     "exit\n",
		Timeout:  5 * time.Second,
	}

	env, sess, err := tui.NewEnv(shSpec)
	if err != nil {
		log.Fatalf("NewEnv: %v", err)
	}
	defer sess.Close()

	skill := driveSkill()
	id, _ := skills.Identity(skill)
	fmt.Printf("dhnt skill %q  identity=%s\n", skill.Name, id[:18]+"…")

	att, err := skills.Run(skill, env, "sh-driver")
	if err != nil {
		log.Fatalf("Run: %v\noutput:\n%s", err, sess.Output())
	}
	fmt.Printf("driving /bin/sh over a real PTY:\n")
	fmt.Printf("  valid=%v passed=%v effects=%v consistent=%v\n",
		att.Valid, att.Passed, att.Effects, att.Consistent(skill))
	fmt.Printf("  pty output: %q\n", oneLine(sess.Output()))

	fmt.Println("\nSame skill, swap the Spec to drive an agentic CLI:")
	fmt.Println(agentCLISpecExample)
}

// agentCLISpecExample shows how the identical driveSkill() drives an
// agentic coding CLI — only the Spec changes.
const agentCLISpecExample = `  tui.Spec{
    Argv:     []string{"claude"},                 // or codex / gemini / aider / opencode
    Patterns: map[string]string{
      "readayu":  ` + "`" + `(?s)(Welcome|>|›)` + "`" + `,    // tool is ready for input
      "resulito": ` + "`" + `(?s)(Done|✓|\$ $)` + "`" + `,        // task acknowledged / complete
    },
    Inputs: map[string]string{"iniputo": "summarize the README\n"},
    Quit:   "/exit\n",
    Timeout: 60 * time.Second,
  }`

func oneLine(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\n' || r == '\r' {
			out = append(out, '⏎')
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
