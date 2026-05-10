---
name: refactor
description: Improve the structure of existing code without changing behaviour
phase: build
user_invocable: true
aliases: []
origin: new
---

# /refactor — restructure without changing behaviour

The user has identified code they want improved. Your job is to
change *how* it's written, not *what* it does. Behaviour stays
identical; tests stay green.

## Steps

1. **Confirm the goal.** What specifically should improve? Common
   targets:
   - Reduce duplication (DRY where it actually helps).
   - Clarify naming (variables, functions, types).
   - Decompose a too-long function.
   - Tighten types / introduce stronger types.
   - Remove dead code.
   - Lower coupling between modules.
   - Improve testability (extract pure functions, inject
     dependencies).
2. **Read the relevant code thoroughly** — not just the target
   site, but the callers and tests. Refactors fail when you miss
   a caller.
3. **Read the existing tests.** They define behaviour. If the
   tests are sparse, surface that — refactoring without test
   coverage is risky and you should ask before proceeding.
4. **Plan the smallest sequence of mechanical edits** that
   achieves the goal. Each step should compile and pass tests.
5. **Apply the steps.** After each, run the tests. If anything
   goes red, stop and back out. Do not "fix forward" by changing
   behaviour to match the new code.
6. **Run the full test suite + linter + type checker** before
   declaring done.

## Constraints

- **No behaviour change.** Same inputs produce same outputs. If
  the user wants behaviour to change, that's a different skill
  (`implement-feature` or `fix-bug`).
- **No drive-by features.** Don't expand scope ("while I was in
  there I added X"). Refactor first, separate commit, then add
  features in another diff.
- **No new dependencies** unless the refactor specifically
  swaps an existing dep.
- **Public API stability** unless the user okayed a breaking
  change. Internal APIs are fair game.

## When NOT to refactor

- Tests are red before you start. Fix tests first.
- The target code is being rewritten / replaced soon.
- The "improvement" is purely stylistic and the project doesn't
  share that style preference (check existing patterns).
- The change touches many unrelated files for a small benefit
  (the refactor is too big — split it).

## Output shape

The diff itself. Plus a short summary:

```
## Refactor summary
- Goal: <what improved>
- Approach: <one sentence>
- Files: <count> changed, <±lines>
- Tests: <green | added new|>

## Behaviour invariants verified
- <test or property>
- <test or property>
```
