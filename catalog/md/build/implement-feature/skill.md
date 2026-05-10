---
name: implement-feature
description: Implement a new feature end-to-end — code, tests, docs, no scope creep
phase: build
user_invocable: true
aliases: [feature, build-feature]
origin: new
---

# /implement-feature — ship a new feature cleanly

A user-facing or API-facing capability needs to land. Drive it from
"described" to "merged" with discipline.

## Steps

1. **Restate the requirement** in your own words; surface gaps.
   - What's the user outcome?
   - What's the API contract (function signature / endpoint /
     CLI flag)?
   - What's explicitly out of scope?
   - What's the success criterion (test, behaviour, metric)?
2. **Skim relevant existing code** to find the right insertion
   point and match local conventions. Don't introduce a new
   architectural pattern when an existing one will do.
3. **Sketch the smallest viable implementation.** Identify each
   file that needs changes. Order them so the codebase stays
   compilable at every step.
4. **Write tests first** for the externally-observable behaviour.
   Run them; confirm they fail.
5. **Implement** the smallest change that turns those tests
   green. Resist refactoring adjacent code now — separate diff.
6. **Run the full test suite + linter + type checker.**
7. **Update docs** that are now stale: README sections, public
   API docs, examples. Don't update internal design docs unless
   the feature changes the design.
8. **Self-review the diff** against the original requirement
   before declaring done.

## Constraints

- **No drive-by changes.** Files unrelated to the feature stay
  untouched.
- **Minimal new dependencies.** If you need one, justify it.
- **Backwards compatibility** unless the user okayed a break.
- **Match existing conventions** — naming, error handling,
  logging, test layout. Don't introduce a new style without
  buy-in.
- **Feature flags / build tags** if the project's strategy doc
  requires new features behind a flag (most projects do for
  alpha-tier work; check the strategy doc).

## What NOT to do

- Build a generic abstraction for one consumer ("future-proofing").
  Three concrete uses is the bar for an abstraction; one isn't.
- Add error-handling for cases that can't happen (overly
  defensive code at trusted-internal boundaries).
- Ship without tests because "this is a small change". Coverage
  catches regressions you don't anticipate.
- Skip the doc updates because "we'll do them later".

## Output shape

```
## Summary
<1-2 sentences: what shipped>

## Files changed
- <path> — <one-line role>

## Tests
- <test name> — <what it asserts>

## Docs updated
- <file>

## Out of scope (followups)
- <thing the user might expect that's intentionally not here>
```
