---
name: project-onboarding
description: Run a full first-week onboarding flow on a project — explore, set up, build, find a starter task
phase: onboard
user_invocable: true
aliases: [onboarding]
origin: priorart/openhands
---

# /project-onboarding — first-week onboarding

You're new to this project (as a contributor or as an agent
working in it for the first time). Run a structured onboarding
that gets you from "cloned" to "made a small contribution" in a
day or two of effective work.

## The flow

1. **Explore the codebase** (`/explore-codebase`) to build the
   mental model: what is this, how is it organised, where does
   execution start.
2. **Find the entry point** (`/find-entry-point`) and trace the
   first ~200 lines of startup.
3. **Summarise the architecture** (`/summarize-architecture`) for
   future-you and any agent that follows. Save the output —
   you'll re-read it.
4. **Set up the dev environment** (`/setup-dev-env`). Build,
   test, editor, linter.
5. **Read the contribution guide** (`CONTRIBUTING.md` /
   `DEVELOPING.md`). Capture: branch naming, PR review process,
   test expectations, signing/CLA requirements.
6. **Read the last 50 commits** with `git log --oneline -50`.
   The project's recent direction shapes which contributions are
   welcome and which are out of scope.
7. **Find a starter task**:
   - Issues tagged `good first issue`, `help wanted`,
     `easy starter`.
   - TODO / FIXME comments in code that align with your interest.
   - Doc gaps you noticed during exploration.
   - Test coverage holes (`grep -r "TODO" tests/`).
   Pick one with a clear scope and an owner you can ask
   questions of.
8. **Make the contribution.** Use the daily-loop skills:
   `/implement-feature` or `/fix-bug`, then `/run-tests`,
   `/code-review` your own diff, `/commit`, `/open-pr`.
9. **Capture what you learned** — gotchas, conventions that
   surprised you, places where the docs lied. Update
   `AGENTS.md` if you saw a documentation gap; doc updates are
   the most welcoming kind of first contribution.

## What NOT to do

- Skip the exploration phase and dive straight into the issue
  tracker. Without the mental model, you'll pick a "starter"
  task that's actually hard.
- File issues for things you found confusing without first
  checking whether the project knows about them.
- Pick a starter task that touches the build system or core
  abstractions — those are deceptively hard.
- Refactor "while you're in there." First contributions earn
  trust; refactors earn pushback.
- Submit a PR without reading at least one prior PR end-to-end.
  PR norms vary widely between projects.

## Output shape

```
## Project
<name> — <one-line>

## Mental model
- Architecture: <summary or doc link>
- Entry point: <file:line>
- Key conventions: <2-3 things to know>

## Dev env
- Setup: <ok / pending issues>

## Contribution norms
- Branch: <pattern>
- Review: <process>
- Tests required: <unit / integration / e2e>

## Starter task picked
- <title> — <issue link>
- Scope: <small / medium>
- Owner: <name / team>

## Progress
- [ ] Explored
- [ ] Set up
- [ ] Read 50 commits + CONTRIBUTING
- [ ] Picked task
- [ ] Implemented
- [ ] Tested
- [ ] PR opened
```

## Tone

Onboarding well = humility plus thoroughness. Read more than you
write in the first week. The codebase will teach you which of
your assumptions are wrong; let it.
