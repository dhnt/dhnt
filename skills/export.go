// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteBundle writes a drop-in skill folder: <dir>/SKILL.md (for any
// Skills-capable tool) and <dir>/skill.dhnt (the canonical form, for a
// dhnt-aware runtime). dir is created if needed.
func WriteBundle(dir string, s Skill, g *Glossary, meta SkillMeta) error {
	md, err := ExportSkillMD(s, g, meta)
	if err != nil {
		return err
	}
	canon, err := LineariseDhnt(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(md), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "skill.dhnt"), []byte(canon+"\n"), 0o644)
}

// SkillMeta carries the human-facing frontmatter for a SKILL.md export.
// Name and Description are required (Description is what a Skills-capable
// tool reads to decide relevance).
type SkillMeta struct {
	Name          string
	Description   string
	Phase         string
	UserInvocable bool
	Aliases       []string
}

// ExportSkillMD renders a dhnt skill as a standard Anthropic-style
// SKILL.md so that ANY Skills-capable tool can use it with no extra
// config: the body is self-contained prose (steps + success criteria +
// effect constraints) that a generic model follows directly, while
// `executor: cnl` and the embedded canonical form let dhnt-aware
// runtimes execute it deterministically and attest the result. A tool
// that doesn't know dhnt simply ignores the unknown executor and follows
// the prose. g may be nil (identifiers then render in dhnt form).
func ExportSkillMD(s Skill, g *Glossary, meta SkillMeta) (string, error) {
	if meta.Name == "" {
		return "", fmt.Errorf("skills: SkillMeta.Name is required")
	}
	if meta.Description == "" {
		return "", fmt.Errorf("skills: SkillMeta.Description is required")
	}
	canon, err := LineariseDhnt(s)
	if err != nil {
		return "", err
	}
	id, err := Identity(s)
	if err != nil {
		return "", err
	}
	resolve := func(dh string) string {
		if g != nil {
			if e := g.LookupDhnt(dh); e != nil {
				if lbl := e.PrimaryLabel("en"); lbl != "" {
					return lbl
				}
			}
		}
		return dh
	}

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", meta.Name)
	fmt.Fprintf(&b, "description: %s\n", meta.Description)
	if meta.Phase != "" {
		fmt.Fprintf(&b, "phase: %s\n", meta.Phase)
	}
	b.WriteString("executor: cnl\n")
	if meta.UserInvocable {
		b.WriteString("user_invocable: true\n")
	}
	if len(meta.Aliases) > 0 {
		fmt.Fprintf(&b, "aliases: [%s]\n", strings.Join(meta.Aliases, ", "))
	}
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "# %s\n\n%s\n\n", meta.Name, meta.Description)
	b.WriteString("> This is a **dhnt skill** — a machine-checkable procedure. " +
		"The steps and success criteria below are self-contained: any tool can " +
		"follow them. A dhnt-aware runtime can instead execute the canonical " +
		"form at the end and emit a verifiable attestation; the result is the " +
		"same either way.\n\n")

	b.WriteString("## Constraints\n\n")
	b.WriteString(renderEffects(s.EffectCap))
	b.WriteString("\n\n")

	b.WriteString("## Steps\n\n")
	lines := renderSteps(s.Steps, resolve, "")
	if len(lines) == 0 {
		b.WriteString("_Declarative: there are no prescribed steps. Reach the success criteria below by any means allowed by the constraints._\n")
	} else {
		b.WriteString(strings.Join(lines, "\n") + "\n")
	}
	b.WriteString("\n")

	b.WriteString("## Success criteria\n\n")
	if len(s.Contract) == 0 {
		b.WriteString("_No explicit success criteria declared._\n\n")
	} else {
		b.WriteString("The run is successful only if **all** of these hold:\n\n")
		for i := range s.Contract {
			fmt.Fprintf(&b, "- %s\n", renderCheck(&s.Contract[i], resolve))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Canonical form (dhnt)\n\n")
	b.WriteString("For dhnt-aware runtimes — execute this; it re-parses to the same skill:\n\n")
	fmt.Fprintf(&b, "```\n%s\n```\n\n", canon)
	fmt.Fprintf(&b, "identity: `%s`\n", id)
	return b.String(), nil
}

// renderEffects describes the allowed and forbidden effects in prose.
func renderEffects(cap []Effect) string {
	if len(cap) == 0 {
		return "_No effect bound declared (the procedure may cause any effect)._"
	}
	allowed := map[Effect]bool{}
	var allowNames []string
	for _, e := range cap {
		if !allowed[e] {
			allowed[e] = true
			allowNames = append(allowNames, e.String())
		}
	}
	var forbid []string
	for e := EffRead; e <= EffTime; e++ {
		if !allowed[e] {
			forbid = append(forbid, e.String())
		}
	}
	out := "This procedure may only: **" + strings.Join(allowNames, ", ") + "**."
	if len(forbid) > 0 {
		out += " It must not: " + strings.Join(forbid, ", ") + "."
	}
	return out
}

// renderSteps renders a step list as a markdown procedure, recursing into
// branches (rendered as if/otherwise).
func renderSteps(steps []Step, resolve func(string) string, indent string) []string {
	var out []string
	n := 0
	for i := range steps {
		st := &steps[i]
		if st.Branch != nil {
			out = append(out, fmt.Sprintf("%s- If %s:", indent, renderCheck(&st.Branch.Cond, resolve)))
			out = append(out, renderSteps(st.Branch.Then, resolve, indent+"    ")...)
			if len(st.Branch.Else) > 0 {
				out = append(out, fmt.Sprintf("%s- Otherwise:", indent))
				out = append(out, renderSteps(st.Branch.Else, resolve, indent+"    ")...)
			}
			continue
		}
		n++
		line := fmt.Sprintf("%s%d. Run **%s**", indent, n, resolve(st.Primitive))
		if extra := renderArgs(st.Args, resolve); extra != "" {
			line += " " + extra
		}
		if st.Latitude == LatJudge {
			line += " _(judgement allowed)_"
		}
		out = append(out, line)
	}
	return out
}

func renderCheck(c *Check, resolve func(string) string) string {
	s := "**" + resolve(c.Predicate) + "**"
	if extra := renderArgs(c.Args, resolve); extra != "" {
		s += " " + extra
	}
	return s
}

func renderArgs(args []Arg, resolve func(string) string) string {
	if len(args) == 0 {
		return ""
	}
	var parts []string
	for i := range args {
		a := &args[i]
		var val string
		switch a.Value.Kind {
		case ExprRef:
			val = resolve(a.Value.Ref)
		case ExprNumber:
			val = fmt.Sprintf("%d", a.Value.Number)
		default:
			val = "?"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", resolve(a.Name), val))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
