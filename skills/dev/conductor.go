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
	cmdParallel  = "ni" // safe to fan out? exit 0 ⇔ >1 issue AND scopes DISJOINT ⇒ FLEET, else SEQUENTIAL
	cmdSolo      = "lo" // ycode weave start — ONE worker, grind+resume (single issue or shared implementation)
	cmdStuck     = "tu" // worker stuck/blocked OR stopped early w/ goal unmet (judge vs goal, not state)? (exit 0 ⇒ ESCALATE)
	cmdEscalate  = "ke" // nudge / RE-DRIVE: weave say + resume w/ a harder iterative prompt until the goal is met (partial submit ≠ done)
)

// conductorSteps is the shared phase body of every conductor variant:
// PLAN files the work; RESEARCH runs only when the goal is complex (a
// branch); FAN-OUT enlists the team; STEER and CONVERGE shepherd them;
// RETRO captures what was learned. RESEARCH and STEER carry latitude=judge
// (bounded judgement); the rest are deterministic command steps.
func conductorSteps() []skills.Step {
	return []skills.Step{
		{Name: "sa", Primitive: PrimRun, Args: ref(cmdPlan)}, // PLAN: decompose + queue
		{Branch: &skills.Branch{ // RESEARCH only when the goal is complex
			Cond: skills.Check{Predicate: PredExit, Args: ref(cmdComplex)},
			Then: []skills.Step{
				{Name: "si", Primitive: PrimRun, Latitude: skills.LatJudge, Args: ref(cmdResearch)},
			},
		}},
		// ROUTE by PARALLEL-SAFETY, not scale (RETRO lesson, dogfood 2026-06-16):
		// fan out ONLY when work is many AND DISJOINT. Tasks that share an
		// implementation (one feature, overlapping source) must run
		// SEQUENTIALLY — one worker grinding + resuming — because parallel
		// agents on shared code produce competing rewrites that COLLIDE
		// irreconcilably at merge, so the parallel attempt + conflict-
		// resolution costs more time and tokens than doing it sequentially.
		{Branch: &skills.Branch{
			Cond: skills.Check{Predicate: PredExit, Args: ref(cmdParallel)},
			Then: []skills.Step{
				{Name: "fo", Primitive: PrimRun, Args: ref(cmdFanout)}, // FLEET: many DISJOINT issues, in parallel
			},
			Else: []skills.Step{
				{Name: "so", Primitive: PrimRun, Args: ref(cmdSolo)}, // SEQUENTIAL: one worker (single issue OR shared implementation)
			},
		}},
		{Name: "su", Primitive: PrimRun, Latitude: skills.LatJudge, Args: ref(cmdSteer)}, // STEER (judgement)
		{Branch: &skills.Branch{ // ESCALATE when a worker is stuck/blocked OR stopped early with the goal unmet
			Cond: skills.Check{Predicate: PredExit, Args: ref(cmdStuck)},
			Then: []skills.Step{
				{Name: "ne", Primitive: PrimRun, Latitude: skills.LatJudge, Args: ref(cmdEscalate)}, // nudge / re-drive partial
			},
		}},
		{Name: "ta", Primitive: PrimRun, Args: ref(cmdConverge)}, // CONVERGE: wait + pull verified
		{Name: "te", Primitive: PrimRun, Args: ref(cmdRetro)},    // RETRO: report card / learn
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
	solo := []string{"sh", "-c", `ycode weave start -- ` + tool}
	escalate := []string{"sh", "-c",
		`for n in $(ycode weave list --json | jq -r '.[] | select(.state == "blocked") | .number'); do ` +
			`ycode weave say "$n" "you appear blocked — continue, or write a BLOCKERS note and exit"; done`}
	return Spec{
		Dir:     dir,
		Timeout: 60 * time.Minute,
		Commands: map[string]Command{
			cmdPlan: {
				// Idempotent: file the goal only if it isn't already queued,
				// so re-running rounds (--max-rounds) doesn't duplicate work.
				// If the queue can't be read the guard falls through to add
				// (worst case = the prior always-add behaviour).
				Argv: []string{"sh", "-c",
					`if ycode weave list --json 2>/dev/null | jq -e --arg g "$GOAL" 'any(.[]; .title == $g)' >/dev/null 2>&1; then :; ` +
						`else ycode weave add "$GOAL" --priority p1 --body "$GOAL"; fi`},
				Env:     []string{"GOAL=" + goal},
				Effects: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffNet},
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
			cmdParallel: {
				// Parallel-safe iff >1 open issue AND their scopes are pairwise
				// disjoint (no two share a top-level scope dir). Unknown or
				// overlapping scopes ⇒ non-zero ⇒ SEQUENTIAL — the safe default
				// that encodes "shared implementations collide in parallel".
				Argv: []string{"sh", "-c",
					`j=$(ycode weave list --json); ` +
						`n=$(printf '%s' "$j" | jq '[.[]|select(.state=="todo")]|length'); ` +
						`[ "${n:-0}" -gt 1 ] || exit 1; ` +
						`dups=$(printf '%s' "$j" | jq '[.[]|select(.state=="todo")|(.scope // "")|split("/")[0]]|group_by(.)|map(select(length>1))|length'); ` +
						`[ "${dups:-1}" -eq 0 ]`},
				Effects: []skills.Effect{skills.EffRead, skills.EffNet},
			},
			cmdSolo: {
				Argv:    solo,
				Effects: []skills.Effect{skills.EffWrite, skills.EffNet, skills.EffSpend, skills.EffTime},
			},
			cmdStuck: {
				Argv:    []string{"sh", "-c", `test "$(ycode weave list --json | jq '[.[] | select(.state == "blocked")] | length')" -gt 0`},
				Effects: []skills.Effect{skills.EffRead, skills.EffNet},
			},
			cmdEscalate: {
				Argv:    escalate,
				Effects: []skills.Effect{skills.EffWrite, skills.EffNet, skills.EffTime},
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

// --- per-task contracts (P6 composition) ------------------------------

// Command ids for the per-task contract skill.
const (
	cmdTaskGreen = "va" // the task's own tests pass (its expected output)
	cmdCommitted = "mi" // the task's work is committed
)

// ConductorTaskSkill is the per-task contract template — the dhnt analogue
// of CrewAI's per-task expected_output: what ONE enlisted task must satisfy
// (its scoped tests pass AND the work is committed), within a task-sized
// effect cap {read, write, time}. Conductor composes it (pillar P6): the
// fleet dispatches one task per issue, each gated by this contract. Because
// a task's cap is a subset of the conductor's, Library.EffectViolations can
// prove statically that composing tasks never widens the orchestrator's
// blast radius.
func ConductorTaskSkill() skills.Skill {
	return skills.Skill{
		Name:      "tasoki", // = dhnt.EncodeWord("task")
		EffectCap: []skills.Effect{skills.EffRead, skills.EffWrite, skills.EffTime},
		Contract: []skills.Check{
			{Predicate: PredExit, Args: ref(cmdTaskGreen)}, // expected output: tests green
			{Predicate: PredExit, Args: ref(cmdCommitted)}, // work committed
		},
	}
}

// ConductorComposedSkill is the conductor with its FLEET leaf rewritten into
// a sub-call to the per-task skill (P6): it expresses that the fleet
// dispatches per-task sub-skills, each carrying its own contract. It is for
// static composition analysis (Library.Closure / EffectViolations), not
// execution — the runnable ConductorSkill binds leaf primitives.
func ConductorComposedSkill() skills.Skill {
	s := ConductorSkill()
	steps := conductorSteps()
	subcallFleet(steps, ConductorTaskSkill().Name)
	s.Steps = steps
	return s
}

// subcallFleet rewrites the fleet leaf (the step calling cmdFanout) into a
// sub-call to task, recursing into branch arms.
func subcallFleet(steps []skills.Step, task string) {
	for i := range steps {
		st := &steps[i]
		if st.Branch != nil {
			subcallFleet(st.Branch.Then, task)
			subcallFleet(st.Branch.Else, task)
			continue
		}
		if len(st.Args) == 1 && st.Args[0].Value.Kind == skills.ExprRef && st.Args[0].Value.Ref == cmdFanout {
			st.Primitive = task
			st.Args = nil
		}
	}
}

// ConductorLibrary returns a library holding the per-task skill and the
// composed conductor that calls it, plus the composed skill, so a caller can
// audit the composition: Closure(composed) ⊇ {task} and
// EffectViolations(composed) is empty (the task's cap is within the
// conductor's).
func ConductorLibrary() (*skills.Library, skills.Skill) {
	lib := skills.NewLibrary()
	_ = lib.Add(ConductorTaskSkill())
	composed := ConductorComposedSkill()
	_ = lib.Add(composed)
	return lib, composed
}
