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
	"time"

	"github.com/dhnt/dhnt/skills"
)

// --- conductor: a goal-oriented orchestrator -------------------------
//
// conductor is the generic director that sits ABOVE the single-purpose
// orchestration skills (weave's parallel fan-out, foreman's sequential
// dispatch, autopilot's inline loop) and drives a team of agent CLIs to
// achieve a goal. It unifies their shared spine — decompose, isolate,
// gate, converge — into one dhnt skill where:
//
//   - the GOAL is the Contract (pillar P1, outcome-first): a run is valid
//     iff a user-supplied verifier exits 0 AND all dispatched work merged;
//   - the phases are Steps (PLAN → optional RESEARCH → FAN-OUT → STEER →
//     CONVERGE → RETRO), with latitude=judge where the work is a bounded
//     judgement call rather than a deterministic command;
//   - the effect cap bounds the blast radius to {read, write, net, spend,
//     time} — agent CLIs reach the network and spend tokens, the fleet
//     writes, waiting burns wall-clock — but never `destroy`.
//
// "Goal-oriented until done" is encoded the dhnt way: not as a loop
// construct (the language has none), but as the goal Contract. A single
// Run executes the phases once and attests against the goal; the caller
// (or the self-healing Runtime, bounded by Repair.MaxAttempts) re-invokes
// until the contract holds. The convergence/orchestration surface — the
// concrete `ycode weave …` argv, the verifier, the roster — lives in the
// Spec, so the canonical skill stays portable and free of free text.

// Command ids for the conductor skill (canonical dhnt CV atoms; the
// concrete argv lives in ConductorSpec so the orchestration surface is
// runtime config, not part of the content-addressed skill).
const (
	cmdPlan      = "pa" // ycode weave add — decompose the goal into queued issues
	cmdComplex   = "bo" // is the goal complex? (exit 0 ⇒ yes ⇒ research first)
	cmdResearch  = "re" // research the goal via an agent CLI (only when complex)
	cmdFanout    = "fa" // ycode weave start — enlist one agent per open issue
	cmdSteer     = "wo" // ycode weave list — steering pass (say-on-demand is judgement)
	cmdConverge  = "vo" // ycode weave wait + pull verified work
	cmdRetro     = "ru" // ycode weave list --summary — tool report card / learn
	cmdGoalMet   = "go" // the user goal verifier — exit 0 ⇔ goal achieved (the spine)
	cmdConverged = "cu" // exit 0 ⇔ no open/unmerged work remains
	cmdReview    = "vi" // independent post-convergence review — exit 0 ⇔ merged result is clean
)

// conductorSteps is the shared phase body of every conductor variant:
// PLAN files the work; RESEARCH runs only when the goal is complex (a
// branch); FAN-OUT enlists the team; STEER and CONVERGE shepherd them;
// RETRO captures what was learned. RESEARCH and STEER carry latitude=judge
// (bounded judgement); the rest are deterministic command steps.
func conductorSteps() []skills.Step {
	return []skills.Step{
		{Name: "sa", Primitive: PrimRun, Args: ref(cmdPlan)}, // PLAN: decompose + queue
		{Name: "se", Branch: &skills.Branch{ // RESEARCH only when the goal is complex
			Cond: skills.Check{Predicate: PredExit, Args: ref(cmdComplex)},
			Then: []skills.Step{
				{Name: "si", Primitive: PrimRun, Latitude: skills.LatJudge, Args: ref(cmdResearch)},
			},
		}},
		{Name: "so", Primitive: PrimRun, Args: ref(cmdFanout)},                           // FAN-OUT: enlist the team
		{Name: "su", Primitive: PrimRun, Latitude: skills.LatJudge, Args: ref(cmdSteer)}, // STEER (judgement)
		{Name: "ta", Primitive: PrimRun, Args: ref(cmdConverge)},                         // CONVERGE: wait + pull verified
		{Name: "te", Primitive: PrimRun, Args: ref(cmdRetro)},                            // RETRO: report card / learn
	}
}

// conductorSkill builds a conductor with the given goal-met predicate (the
// only axis that varies between the deterministic and judge variants). The
// Contract — the spine — is goal-met ∧ converged ∧ reviewed: a run is valid
// only if the goal holds, all dispatched work has merged, AND an
// independent post-convergence review of the merged result passes (the
// regression gate, so a merged combination that breaks the tree is caught
// before accept). OnFail is PolicyBlockers: an unmet goal surfaces blockers
// and exits gracefully rather than crashing — each run makes progress,
// re-invoke to continue (the "goal-oriented until done" loop).
func conductorSkill(goalPredicate string) skills.Skill {
	return skills.Skill{
		Name: "coniducatoro", // = dhnt.EncodeWord("conductor")
		EffectCap: []skills.Effect{
			skills.EffRead, skills.EffWrite, skills.EffNet, skills.EffSpend, skills.EffTime,
		},
		Steps: conductorSteps(),
		Contract: []skills.Check{
			{Predicate: goalPredicate, Args: ref(cmdGoalMet)}, // the goal holds (the spine)
			{Predicate: PredExit, Args: ref(cmdConverged)},    // ∧ all dispatched work converged
			{Predicate: PredExit, Args: ref(cmdReview)},       // ∧ merged result passes review
		},
		OnFail: skills.PolicyBlockers,
	}
}

// ConductorSkill is the goal-oriented orchestrator with a deterministic,
// exit-coded goal-met gate (a verify command).
func ConductorSkill() skills.Skill { return conductorSkill(PredExit) }

// ConductorJudgeSkill is the conductor variant whose goal-met check is a
// model JUDGE (PredJudge) instead of a deterministic exit code: for goals
// with no clean pass/fail verifier, an agent reads the evidence (the
// converged work) and judges whether the goal is achieved. The convergence
// and review gates stay deterministic (exit-coded), and the phases are
// identical. A different contract means a different content address — this
// is a distinct skill from ConductorSkill, not a reconfiguration of it.
func ConductorJudgeSkill() skills.Skill { return conductorSkill(PredJudge) }

// agentHeadlessArgv returns the one-shot (headless) argv for a catalogued
// agent CLI. It is intentionally a small local table rather than a
// dependency on skills/tui (which would couple the two leaves): the
// conductor's RESEARCH phase only needs the headless prompt form. Unknown
// agents fall back to the claude-style `-p` flag.
func agentHeadlessArgv(agent, prompt string) []string {
	switch agent {
	case "codex":
		return []string{"codex", "exec", prompt}
	case "opencode":
		return []string{"opencode", "run", prompt}
	case "aider":
		return []string{"aider", "--message", prompt}
	case "gemini":
		return []string{"gemini", "-p", prompt}
	default: // claude and anything else line-oriented with -p
		return []string{"claude", "-p", prompt}
	}
}

// ConductorSpec binds the conductor's phases to a concrete `ycode weave`
// orchestration surface for a project at dir, driving the goal to
// completion with the given agent roster. verifyArgv is the goal verifier
// (e.g. ["go","test","./..."]); when empty it defaults to the convergence
// gate (done == all queued work merged). complexThreshold is the queued-
// issue count above which the RESEARCH phase runs (default 3). The goal
// text is injected via the process environment (GOAL=…), never embedded in
// an argv string — the same discipline that keeps free text out of the
// canonical skill.
func ConductorSpec(dir, goal string, roster, verifyArgv, reviewArgv []string, complexThreshold int) Spec {
	if complexThreshold <= 0 {
		complexThreshold = 3
	}
	tool := "claude"
	if len(roster) > 0 && roster[0] != "" {
		tool = roster[0]
	}
	convergedArgv := []string{"sh", "-c",
		`test "$(ycode weave list --json | jq '[.[] | select(.state != "merged")] | length')" -eq 0`}
	if len(verifyArgv) == 0 {
		verifyArgv = convergedArgv
	}
	// The review gate defaults to re-running the goal verifier on the
	// merged result (a post-convergence regression check); a distinct
	// --review (linter / code-review / smoke) makes it an independent gate.
	if len(reviewArgv) == 0 {
		reviewArgv = verifyArgv
	}
	research := agentHeadlessArgv(tool, "Research approaches, prior art, and risks for: "+goal)
	fanout := []string{"sh", "-c",
		`ycode weave list --json | jq -r '.[] | select(.state == "todo") | .number' | ` +
			`while read n; do ycode weave start --issue "$n" -- ` + tool + ` & done; wait`}
	return Spec{
		Dir:     dir,
		Timeout: 60 * time.Minute,
		Commands: map[string]Command{
			cmdPlan: {
				Argv:    []string{"sh", "-c", `ycode weave add "$GOAL" --priority p1 --body "$GOAL"`},
				Env:     []string{"GOAL=" + goal},
				Effects: []skills.Effect{skills.EffWrite, skills.EffNet},
			},
			cmdComplex: {
				Argv:    []string{"sh", "-c", fmt.Sprintf(`test "$(ycode weave list --json | jq 'length')" -gt %d`, complexThreshold)},
				Effects: []skills.Effect{skills.EffRead, skills.EffNet},
			},
			cmdResearch: {
				Argv:    research,
				Effects: []skills.Effect{skills.EffRead, skills.EffNet, skills.EffSpend, skills.EffTime},
			},
			cmdFanout: {
				Argv:    fanout,
				Effects: []skills.Effect{skills.EffWrite, skills.EffNet, skills.EffSpend, skills.EffTime},
			},
			cmdSteer: {
				Argv:    []string{"ycode", "weave", "list"},
				Effects: []skills.Effect{skills.EffRead, skills.EffNet},
			},
			cmdConverge: {
				Argv:    []string{"sh", "-c", `ycode weave wait || true; ycode weave pull`},
				Effects: []skills.Effect{skills.EffWrite, skills.EffRead, skills.EffNet, skills.EffTime},
			},
			cmdRetro: {
				Argv:    []string{"ycode", "weave", "list", "--summary"},
				Effects: []skills.Effect{skills.EffRead, skills.EffNet},
			},
			cmdGoalMet: {
				Argv:    verifyArgv,
				Effects: []skills.Effect{skills.EffRead, skills.EffTime},
			},
			cmdConverged: {
				Argv:    convergedArgv,
				Effects: []skills.Effect{skills.EffRead, skills.EffNet},
			},
			cmdReview: {
				Argv:    reviewArgv,
				Effects: []skills.Effect{skills.EffRead, skills.EffTime},
			},
		},
	}
}

// ConductorJudgeSpec binds the judge variant: the goal-met command becomes
// an EVIDENCE command (what the fleet produced — by default a summary of
// merged work plus recent history) whose output is handed to the judge
// completer, and Spec.Goal/Spec.Judge are set so the `judged` predicate is
// bound. evidenceArgv overrides the default evidence command. Use this with
// ConductorJudgeSkill.
func ConductorJudgeSpec(dir, goal string, roster, evidenceArgv, reviewArgv []string, judge skills.Completer, complexThreshold int) Spec {
	spec := ConductorSpec(dir, goal, roster, nil, reviewArgv, complexThreshold)
	if len(evidenceArgv) == 0 {
		evidenceArgv = []string{"sh", "-c", `ycode weave list --summary 2>/dev/null; echo '--- recent commits ---'; git log --oneline -10 2>/dev/null`}
	}
	gm := spec.Commands[cmdGoalMet]
	gm.Argv = evidenceArgv
	gm.Effects = []skills.Effect{skills.EffRead, skills.EffNet}
	spec.Commands[cmdGoalMet] = gm
	spec.Goal = goal
	spec.Judge = judge
	return spec
}
