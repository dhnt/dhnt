---
name: fix-test
description: Diagnose and fix a failing test — find out whether the test is wrong or the code is wrong
phase: test
user_invocable: true
aliases: [repair-test]
origin: priorart/openhands
---

# /fix-test — repair a failing test

A test is red. Before doing anything, decide: is the **test**
wrong, or is the **code** wrong?

## Triage

1. **Reproduce locally** — run the failing test in isolation. If
   it doesn't fail locally, you probably have an environment
   delta (different OS, different Go/Node version, race
   condition, flaky external service). Don't fix what you can't
   reproduce.
2. **Read the failure carefully.** The assertion message often
   tells you which side of the contract is wrong. Look at:
   - Expected vs actual values.
   - Stack trace — where did execution diverge?
   - Recent changes in the file under test and the test file
     itself (`git log -n 5`).
3. **Decide who's wrong**:
   - **Test is wrong** if: the contract changed deliberately, the
     test was written against a stale implementation, the
     fixture is out of date, the mock no longer matches the real
     thing.
   - **Code is wrong** if: the contract is the same but the code
     diverged, an edge case was missed, a race condition was
     introduced.
   - **Both wrong** if: the contract is fuzzy and neither the
     test nor the code commits to it. Surface this and ask for
     direction.

## Process per case

### Code is wrong

Apply `/fix-bug` to the underlying issue. The test stays as the
regression that catches it.

### Test is wrong

1. **Confirm intent**: ask the user (or the PR author, via
   blame/log) what behaviour was intended. Don't silently rewrite
   a test to match buggy code.
2. **Update the test** — assertion, fixture, or mock — to reflect
   the corrected contract.
3. **Verify the test now fails for the right reason** by
   temporarily breaking the code and re-running. The test should
   catch the break.

### Flaky test

If the test passes sometimes and fails sometimes:

1. Run it 100 times in a tight loop with different seeds /
   orderings.
2. Common causes: time-dependent assertions, ordering of map
   iterations, external service latency, shared global state,
   real network calls.
3. **Don't mark flaky tests as "pending" or skipped**. Either fix
   them or delete them with a clear note.

## What NOT to do

- Comment out the assertion to make CI green.
- Add a `sleep()` to "fix" timing flakiness.
- Update the assertion to match buggy output ("expected = actual"
  via copy-paste from the failure).
- Skip the test ("it's an integration test, it's noisy").
- Land a fix without running the test you fixed at least 5 times
  consecutively (or 100 for flake-prone areas).

## Output shape

```
## Diagnosis
- Test or code? <one>
- Root cause: <one paragraph>

## Fix
<file:line — what changed>

## Verification
- Test now passes: <run count>
- Reverting the fix makes the test fail again: <yes/no>
```
