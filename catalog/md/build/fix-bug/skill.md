---
name: fix-bug
description: Diagnose a reported bug, find the root cause, fix it, and add a test that would have caught it
phase: build
user_invocable: true
aliases: [bugfix]
origin: new
---

# /fix-bug — root-cause analysis and fix

A bug was reported. Don't guess at the fix. Reproduce, isolate,
fix the root cause, and add a regression test.

## Steps

1. **Reproduce.** Get the bug in front of you in a minimal form.
   - If the report has a repro: run it.
   - If not: ask for the smallest input that triggers it, or
     construct one from the description.
   - If you can't reproduce, stop and say so. Don't ship a
     speculative fix.
2. **Isolate the failing component.** Bisect by:
   - Reading the stack trace / error log line by line.
   - Adding temporary log statements *only* inside test runs (not
     in shipped code).
   - Checking recent git history for changes near the reported
     site.
3. **Find the root cause, not the symptom.** Ask "why does this
   line fail?" and keep asking until the answer is "because of
   how the code was written" rather than "because of the input".
   Symptoms are easy to silence; root causes prevent the next
   five bugs of the same shape.
4. **Write the regression test first.** A test that fails on the
   current code and would catch this bug if it ever returns. Run
   it; confirm it fails.
5. **Apply the fix.** Smallest change that makes the test pass.
   Don't refactor while you fix — that's a separate skill.
6. **Re-run the full test suite + linter.**
7. **Audit nearby code** for the same shape of bug — root causes
   often have siblings. Note them; either fix in this commit if
   trivial, or file follow-ups.

## Output shape

```
## Root cause
<one paragraph — what was wrong, why the symptom appeared>

## Fix
<one paragraph — what changed, why this is the right level>

## Regression test
<file:line of the new test>

## Sibling concerns (N)
- <file:line — same shape of bug? note or fix>
```

## What NOT to do

- Silence the symptom (e.g. wrap in a try/catch and continue).
- Add a "defensive" check that hides the bug rather than fixing
  it.
- Skip the regression test ("it's obvious how the fix works").
- Mark the bug fixed without running the test you wrote.
- Stretch the diff with unrelated cleanup. Cleanup is a separate
  commit.

## When to escalate

- The fix is in code you don't have permission to change
  (third-party, generated, vendored).
- The root cause is a design flaw and the right fix is a
  refactor at scale — call that out instead of patching.
- The reproduction itself reveals a more serious issue than the
  reported one.
