---
name: deep-planning
description: Engage extended reasoning for an ambitious, multi-step task — surface assumptions, decompose, sequence
phase: plan
user_invocable: true
aliases: [think, reason]
origin: priorart/cline
---

# /deep-planning — extended reasoning for ambitious work

Use this skill when the task ahead is meaningfully harder than a
single feature or a single bug fix: a substantial refactor, a
new subsystem, a migration, an architectural change. The point is
to slow down before acting.

## When to invoke

Symptoms that warrant deep planning:

- The task touches more than ~5 files or ~3 components.
- The "obvious" implementation has non-obvious edge cases.
- Multiple plausible designs exist and the right choice isn't
  clear.
- The work spans more than a single working session.
- A wrong move would be expensive to undo (data migration,
  public-API change, infrastructure swap).
- The user explicitly said "think harder", "plan first", "be
  thorough."

If none of the above apply, this skill is overkill — use
`/design-feature` or `/plan-refactor` instead.

## Steps

1. **Restate the task** in your own words. State what you think
   the user wants in one paragraph. Surface every assumption you
   noticed yourself making.
2. **Question those assumptions.** For each one, ask: what if it's
   wrong? What does the right design look like in that case? At
   least one assumption is usually wrong; this is how you find it
   before committing.
3. **Decompose into milestones.** A milestone is a checkpoint with
   verifiable behaviour — not just a chunk of code. "Migration
   tool reads old format" is a milestone; "wrote 200 lines of
   migration code" isn't.
4. **For each milestone, identify**:
   - The smallest output that proves it's done.
   - The dependencies on prior milestones.
   - The verification gate (test, manual check, demo).
   - The risk that it might not work and what we'd do then.
5. **Find the biggest unknown** and design a spike for it. A spike
   is a throwaway exploration to learn the answer to the question:
   "is this approach actually viable?" Doing it before committing
   to the plan is cheap insurance.
6. **Pick the order** that surfaces risk earliest. Tackle the
   uncertain milestones before the certain ones — failure on a
   late, certain milestone is a nuisance; failure on an early,
   uncertain one is a course correction.
7. **Write down what success looks like.** What does the system do
   (or fail to do) that proves the work is done? "All tests
   pass" is not enough. "User can do X without seeing Y" is
   better.

## Output shape

```
## Task
<one paragraph: what the user wants, in your words>

## Assumptions I'm making (and what changes if they're wrong)
1. <assumption> — if wrong: <consequence>
2. <assumption> — if wrong: <consequence>

## Milestones
1. <milestone name>
   - Done when: <verifiable check>
   - Depends on: <prior milestones>
   - Risk: <what might not work>

2. <milestone name>
   …

## Biggest unknown
<question>
- Plan to resolve: <spike approach, time budget>

## Order
<which milestone first and why>

## Definition of done
<what the system does or doesn't do that proves it's complete>

## What I'm NOT doing
- <out-of-scope thing that might be expected>
```

## What NOT to do

- Plan in service of looking thorough. The plan is for finding
  problems, not for show. If three sentences cover it, three
  sentences is the plan.
- Avoid commitment by piling on alternatives. Pick a direction;
  document why; move on.
- Skip the assumption-questioning step. Most expensive bugs were
  unquestioned assumptions.
- Plan once, freeze forever. The plan is a hypothesis; revise it
  when the spike or first milestone reveals new information.
- Confuse "deep" with "long." The best deep plans are short and
  surgical because they identify the *one* hard thing and address
  it directly.

## After the plan

Hand it to the user for sign-off before executing. Substantial
work is best aligned-on once than course-corrected mid-flight.
