// Copyright 2026 The dhnt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Command dhnt is the reference CLI for dhnt skills.
//
//	dhnt export   — render a skill (canonical .dhnt) as a SKILL.md bundle
//	dhnt run      — execute a skill against a tool (the runner) and attest
//	dhnt normalise — turn plain-English prose into a dhnt skill via an agent
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dhnt/dhnt/skills"
	"github.com/dhnt/dhnt/skills/tui"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "export":
		err = cmdExport(os.Args[2:])
	case "run":
		err = cmdRun(os.Args[2:])
	case "normalise", "normalize":
		err = cmdNormalise(os.Args[2:])
	case "promote":
		err = cmdPromote(os.Args[2:])
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "dhnt:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `dhnt — dhnt skills CLI

  dhnt export   [--skill FILE] --name N --desc D [--phase P] [--out DIR]
                read a canonical .dhnt skill (or stdin) and write a
                SKILL.md bundle (stdout if no --out).

  dhnt run      (--agent NAME [--headless] [--message M]) | (--spec FILE [--skill FILE])
                execute a skill against a tool over a real PTY and print
                the verifiable attestation.

  dhnt normalise --agent NAME [--out DIR] [prose...]
                turn plain-English prose (args or stdin) into a dhnt skill
                using an agent CLI as the model; print canonical or write
                a bundle.

  dhnt promote  --skill FILE --name N --desc D [--parent ID] --out DIR
                render a host-learned (folded) skill version as a
                review-ready bundle for the catalog. Never auto-commits —
                a human reviews and merges.
`)
}

// --- export -----------------------------------------------------------

func cmdExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	skillFile := fs.String("skill", "", "canonical .dhnt skill file (default stdin)")
	name := fs.String("name", "", "SKILL.md name (required)")
	desc := fs.String("desc", "", "SKILL.md description (required)")
	phase := fs.String("phase", "", "lifecycle phase")
	aliases := fs.String("aliases", "", "comma-separated aliases")
	out := fs.String("out", "", "output dir for the bundle (default: SKILL.md to stdout)")
	_ = fs.Parse(args)

	canon, err := readSource(*skillFile)
	if err != nil {
		return err
	}
	skill, err := skills.ParseDhnt(strings.TrimSpace(canon))
	if err != nil {
		return fmt.Errorf("parse skill: %w", err)
	}
	g, err := skills.SeedGlossary()
	if err != nil {
		return err
	}
	meta := skills.SkillMeta{Name: *name, Description: *desc, Phase: *phase, UserInvocable: true}
	if *aliases != "" {
		meta.Aliases = strings.Split(*aliases, ",")
	}
	if *out == "" {
		md, err := skills.ExportSkillMD(skill, g, meta)
		if err != nil {
			return err
		}
		fmt.Print(md)
		return nil
	}
	if err := skills.WriteBundle(*out, skill, g, meta); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %s/SKILL.md and %s/skill.dhnt\n", *out, *out)
	return nil
}

// --- run (the runner) -------------------------------------------------

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	agent := fs.String("agent", "", "catalogued agent CLI (claude|codex|gemini|opencode|aider)")
	headless := fs.Bool("headless", false, "with --agent: one-shot model call instead of --version")
	message := fs.String("message", "Reply with exactly one word: PONG", "with --headless: the prompt")
	specFile := fs.String("spec", "", "tool Spec JSON (custom run)")
	skillFile := fs.String("skill", "", "canonical .dhnt skill (with --spec; default: DriveOnce)")
	adapt := fs.Bool("adapt", false, "self-healing: reuse/learn a host-local version of the skill")
	storeDir := fs.String("store", "", "version overlay dir (default ~/.dhnt/versions)")
	repairAgent := fs.String("repair-agent", "", "agent CLI used as the repair model (with --adapt)")
	_ = fs.Parse(args)

	var spec tui.Spec
	var skill skills.Skill
	var tier string

	switch {
	case *agent != "":
		var ok bool
		if *headless {
			spec, ok = tui.HeadlessSpec(*agent, *message, `(?i)PONG`, "")
		} else {
			spec, ok = tui.VersionSpec(*agent, "")
		}
		if !ok {
			return fmt.Errorf("unknown agent %q", *agent)
		}
		skill = tui.DriveOnceSkill()
		tier = *agent
	case *specFile != "":
		data, err := os.ReadFile(*specFile)
		if err != nil {
			return err
		}
		if spec, err = tui.SpecFromJSON(data); err != nil {
			return fmt.Errorf("spec json: %w", err)
		}
		if *skillFile != "" {
			canon, err := readSource(*skillFile)
			if err != nil {
				return err
			}
			if skill, err = skills.ParseDhnt(strings.TrimSpace(canon)); err != nil {
				return fmt.Errorf("parse skill: %w", err)
			}
		} else {
			skill = tui.DriveOnceSkill()
		}
		tier = "custom"
	default:
		return fmt.Errorf("run needs --agent or --spec")
	}

	env, sess, err := tui.NewEnv(spec)
	if err != nil {
		return err
	}
	defer sess.Close()

	if *adapt {
		return runAdaptive(skill, env, sess, tier, *storeDir, *repairAgent)
	}

	att, runErr := skills.Run(skill, env, tier)
	if tail := lastLines(sess.Output(), 8); tail != "" {
		fmt.Fprintf(os.Stderr, "--- output (tail) ---\n%s\n---------------------\n", tail)
	}
	if runErr != nil {
		return runErr
	}
	fmt.Printf("valid=%v consistent=%v passed=%v failed=%v effects=%v\n",
		att.Valid, att.Consistent(skill), att.Passed, att.Failed, att.Effects)
	if !att.Valid {
		os.Exit(1)
	}
	return nil
}

// runAdaptive wraps a run with the self-healing Runtime: it prefers a
// host-local learned version, and (with --repair-agent) repairs+folds on
// failure. Probes default to OS/arch.
func runAdaptive(skill skills.Skill, env skills.Env, sess *tui.Session, tier, storeDir, repairAgent string) error {
	if storeDir == "" {
		home, _ := os.UserHomeDir()
		storeDir = filepath.Join(home, ".dhnt", "versions")
	}
	g, err := skills.SeedGlossary()
	if err != nil {
		return err
	}
	rt := &skills.Runtime{
		Glossary: g, Lang: "en", Tier: tier,
		Probes:   defaultProbes(),
		Versions: &skills.FileVersionStore{Dir: storeDir},
	}
	if repairAgent != "" {
		if c, ok := tui.Completer(repairAgent, "", 120*time.Second); ok {
			rt.Repair = &skills.Repairer{Complete: c, Glossary: g, Lang: "en", MaxAttempts: 2}
		}
	}
	att, outcome, err := rt.Run(skill, env)
	if tail := lastLines(sess.Output(), 8); tail != "" {
		fmt.Fprintf(os.Stderr, "--- output (tail) ---\n%s\n---------------------\n", tail)
	}
	if err != nil {
		return err
	}
	fmt.Printf("outcome=%s valid=%v consistent=%v passed=%v effects=%v\n",
		outcome, att.Valid, att.Consistent(skill), att.Passed, att.Effects)
	if !att.Valid {
		os.Exit(1)
	}
	return nil
}

func defaultProbes() []skills.EnvProbe {
	return []skills.EnvProbe{{Name: "os", Value: runtime.GOOS}, {Name: "arch", Value: runtime.GOARCH}}
}

// --- normalise --------------------------------------------------------

func cmdNormalise(args []string) error {
	fs := flag.NewFlagSet("normalise", flag.ExitOnError)
	agent := fs.String("agent", "gemini", "agent CLI used as the model")
	out := fs.String("out", "", "write a bundle to this dir (else print canonical)")
	name := fs.String("name", "normalised-skill", "SKILL.md name when --out is set")
	_ = fs.Parse(args)

	prose := strings.Join(fs.Args(), " ")
	if strings.TrimSpace(prose) == "" {
		b, _ := io.ReadAll(os.Stdin)
		prose = strings.TrimSpace(string(b))
	}
	if prose == "" {
		return fmt.Errorf("no prose given (args or stdin)")
	}
	g, err := skills.SeedGlossary()
	if err != nil {
		return err
	}
	complete, ok := tui.Completer(*agent, "", 120*time.Second)
	if !ok {
		return fmt.Errorf("unknown agent %q", *agent)
	}
	skill, cnl, err := skills.Normalise(prose, g, "en", complete, 2)
	if err != nil {
		return err
	}
	if *out == "" {
		fmt.Println(cnl)
		return nil
	}
	meta := skills.SkillMeta{Name: *name, Description: prose, UserInvocable: true}
	if err := skills.WriteBundle(*out, skill, g, meta); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %s/SKILL.md and %s/skill.dhnt\n", *out, *out)
	return nil
}

// --- promote ----------------------------------------------------------

func cmdPromote(args []string) error {
	fs := flag.NewFlagSet("promote", flag.ExitOnError)
	skillFile := fs.String("skill", "", "folded/derived canonical .dhnt skill (default stdin)")
	name := fs.String("name", "", "SKILL.md name (required)")
	desc := fs.String("desc", "", "SKILL.md description (required)")
	parent := fs.String("parent", "", "parent skill identity this was derived from")
	out := fs.String("out", "", "output dir for the review bundle (required)")
	_ = fs.Parse(args)
	if *out == "" {
		return fmt.Errorf("promote needs --out")
	}
	canon, err := readSource(*skillFile)
	if err != nil {
		return err
	}
	skill, err := skills.ParseDhnt(strings.TrimSpace(canon))
	if err != nil {
		return fmt.Errorf("parse skill: %w", err)
	}
	g, err := skills.SeedGlossary()
	if err != nil {
		return err
	}
	if err := skills.WriteBundle(*out, skill, g, skills.SkillMeta{Name: *name, Description: *desc, UserInvocable: true}); err != nil {
		return err
	}
	id, _ := skills.Identity(skill)
	note := fmt.Sprintf(`# Promotion candidate: %s

This skill version was **learned on a host** (a contract-verified
adaptation) and is proposed for the catalog. It is NOT auto-merged.

- derived identity: %s
- parent identity:  %s

## Reviewer checklist
- [ ] the change is a generalisation (e.g. an added environment branch),
      not a weakening of the contract or effect cap;
- [ ] re-verify in the catalog's target environments;
- [ ] no host-specific paths/secrets leaked into steps;
- [ ] effects stay within the original cap.
`, *name, id, orEmpty(*parent))
	if err := os.WriteFile(filepath.Join(*out, "PROMOTION.md"), []byte(note), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote review bundle to %s (SKILL.md, skill.dhnt, PROMOTION.md)\n", *out)
	return nil
}

func orEmpty(s string) string {
	if s == "" {
		return "(unspecified)"
	}
	return s
}

// --- helpers ----------------------------------------------------------

func readSource(path string) (string, error) {
	if path == "" || path == "-" {
		b, err := io.ReadAll(os.Stdin)
		return string(b), err
	}
	b, err := os.ReadFile(path)
	return string(b), err
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
