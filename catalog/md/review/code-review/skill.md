---
name: code-review
description: Review pending changes on the current branch with the rigour of a senior reviewer
phase: review
user_invocable: true
aliases: [review]
origin: priorart/continue
---

# /code-review — review the pending changes

Read the diff between the current branch and the base branch, and
produce review feedback as if a senior engineer were looking at it
for the first time. Optimise for catching real defects, not for
nitpicking.

## Steps

1. **Find the base.** Default to the repo's main branch (`main` or
   `master`). If the user named a branch or PR, target that.
2. **Read the full diff.** Not just the latest commit — the
   complete diff `git diff <base>...HEAD`. Note the magnitude
   (files changed, ±lines).
3. **Read the surrounding code** for any non-trivial change. A
   function modified in isolation often only makes sense once you
   see its callers and tests.
4. **Run the type-checker / linter / tests** if the project
   supports it cheaply (a few seconds). Don't skip this step
   silently.
5. **Group findings by severity.** Do NOT dump a flat list.

## Severity bands

- **Blocker** — the change is incorrect, will break a contract,
  introduces a security/data-loss risk, or violates a documented
  invariant. Always cite the specific line.
- **Concern** — the change is plausibly correct but the design
  choice is questionable, the test coverage is thin, or the
  interaction with other code is unclear. Explain the worry.
- **Suggestion** — small improvements: clearer naming, removing
  dead code, simplifying control flow. Group these together; one
  sentence each.
- **Praise** — when the change actually nailed something hard,
  say so briefly. One line per item, max two items.

## What to look for

- **Correctness**: does the code do what the diff/PR description
  claims? Off-by-one, error swallowing, race conditions, nil
  dereferences, resource leaks.
- **Security**: untrusted input boundaries, command injection,
  path traversal, secrets in logs, unsafe deserialisation.
- **Tests**: do they cover the change, including failure modes?
- **API + contract changes**: backwards-compat breaks, type
  changes, new error returns. Are downstream callers updated?
- **Hidden costs**: O(n²) where O(n) was fine, an extra DB round
  trip per loop iteration, an unnecessary allocation in a hot
  path.
- **Dead code, leftover debug prints, commented-out blocks.**

## What NOT to do

- Don't restate the diff. Reviewers can read it.
- Don't pile on stylistic nits unless the change introduces them.
- Don't suggest refactors that are out of scope for the diff.
- Don't claim a finding without citing the file and line.
- Don't approve work that has *Blocker*-band findings.

## Output shape

```
## Summary
<1-2 sentences: what the diff does, your overall verdict>

## Blockers (N)
- <file:line> — <issue> — <why it matters>

## Concerns (N)
- <file:line> — <issue> — <what to verify>

## Suggestions (N)
- <file:line> — <one-line suggestion>

## Praise (≤2)
- <one-line>
```

Verdict at the top, evidence below. Anyone reading should be able
to triage in 30 seconds.
