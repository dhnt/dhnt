---
name: update-test
description: Update an existing test to reflect an intentional change in the code's contract
phase: test
user_invocable: true
aliases: []
origin: priorart/openhands
---

# /update-test — update a test for a deliberate contract change

The code's contract changed (new field, renamed function, changed
return shape, broader/narrower error type). Existing tests assert
the old contract and now fail. Update them to assert the new one.

This is a different skill from `/fix-test`. `fix-test` triages
whether the test or the code is wrong. `update-test` is invoked
when you already know: the contract changed deliberately, and the
tests need to follow.

## Steps

1. **Confirm the new contract** is what the user expects. A bad
   shortcut here is "delete the assertions, the test is annoying
   to update" — that loses the test's value.
2. **Locate every affected test**, not just the one that
   surfaced. Searching for the old function/field/error name
   typically finds the rest.
3. **For each test**, decide whether the test is still
   meaningful under the new contract:
   - Still meaningful → update the assertions/fixtures.
   - No longer meaningful (the test was asserting the old
     contract specifically) → delete it cleanly. Don't leave a
     hollow shell.
   - Now redundant with another test → consolidate.
4. **Update fixtures and golden files** that reference the old
   shape. Keep these terse — fixtures grow forever if no one
   prunes them.
5. **Add at least one new test** for any new edge case the
   contract change introduced (a new field needs validation
   tests; a new error type needs a test that exercises it).
6. **Run the full test suite + linter.**

## What NOT to do

- Update tests by mass-replace without reading each one. The
  replace might compile but make the test meaningless.
- Delete tests because they're "no longer relevant" without
  checking whether the behaviour they were asserting still
  matters.
- Loosen assertions ("anything is fine, just assert no error")
  to make tests pass — that defeats the purpose.
- Leave commented-out assertions as a "TODO".

## Output shape

```
## Contract change
<one paragraph: what changed>

## Tests updated (N)
- <file:test> — <what shifted in the assertion>

## Tests deleted (N)
- <file:test> — <why it's no longer meaningful>

## Tests added (N)
- <file:test> — <what new edge case it covers>
```
