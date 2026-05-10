---
name: plan-refactor
description: Plan a code reorganisation as a sequence of small, individually-verifiable steps
phase: plan
user_invocable: true
aliases: []
origin: new
---

# /plan-refactor — plan the refactor before doing it

A refactor is risky in proportion to its blast radius. The way to
de-risk one is to plan it as a sequence of small, individually-
verifiable steps where each step compiles and passes tests.

This skill produces the plan. The execution is `/refactor`.

## Steps

1. **State the goal.** What's the smell or constraint driving this?
   - Reduce duplication.
   - Decompose a too-large module.
   - Lower coupling between two layers.
   - Tighten types / introduce stronger types.
   - Improve testability (extract pure functions, inject deps).
   - Remove dead code.
   - Rename for clarity (functions, types, packages).
2. **Confirm test coverage** for the affected code. A refactor
   without tests is a rewrite. If coverage is thin, the first step
   of the plan is "write characterisation tests."
3. **Map blast radius.** For every symbol you'll rename / move /
   change, find every caller. Use the tooling (LSP / `gopls` /
   `find-references`). The blast radius is the union of those
   call-sites.
4. **Sequence the mechanical edits.** Each step should:
   - Compile.
   - Pass the existing test suite.
   - Be small enough to review in one sitting.
   - Have a single intent ("rename X", not "rename X and fix Y").
5. **Plan rollback.** What's the recovery if step N goes red?
   Usually `git reset --soft` to the prior step. If the refactor
   spans multiple commits, the rollback unit is each commit.
6. **Identify the riskiest step** and put it last. By then the
   safer prep work is done; rolling back the risky bit doesn't
   waste the rest.
7. **List what you're NOT doing.** A refactor that bundles in
   feature work or behaviour fixes is hard to review and harder to
   revert. Be explicit about the boundary.

## Output shape

```
## Refactor plan: <one-line goal>

## Why
<one paragraph: smell / constraint driving this>

## Blast radius
- <symbol> — <N call-sites in M files>
- <symbol> — <N call-sites in M files>

## Test coverage status
- <file/module>: <covered / partial / uncovered>
- (if uncovered: first step must be characterisation tests)

## Steps
1. <mechanical edit> — <verify: build + test>
2. <mechanical edit> — <verify: build + test>
3. <mechanical edit> — <verify: build + test>
4. <riskiest step last> — <verify: build + test + manual smoke if needed>

## Rollback
- Per-step: `git reset --soft <prev>`
- Whole refactor: `git revert <range>` or branch reset.

## Out of scope
- <thing the user might expect to be folded in but isn't>

## Estimated time
<hours/days; honest>
```

## What NOT to do

- Plan a refactor as one giant step ("rewrite module X"). One step
  ≠ one PR. Big-bang refactors are hard to review and impossible to
  partially revert.
- Skip the test-coverage check. Find out coverage is thin *before*
  committing to the plan, not when step 3 silently broke
  production.
- Bundle behaviour changes ("while I'm in there, fix that bug").
  The bug fix is a separate diff.
- Make every callsite change in step 1. Move the abstraction first,
  migrate callers second, delete the old abstraction third.
- Ignore the API contract. If the refactor changes a public
  interface, that's not a refactor — it's a breaking change.

## When the right answer is "don't refactor"

- The code is being rewritten / replaced soon.
- The "improvement" is purely stylistic and the project doesn't
  share the preference.
- The blast radius is enormous and the benefit is incremental.
  Defer; revisit when the cost-benefit tilts.
- You can't articulate the goal in one sentence.
