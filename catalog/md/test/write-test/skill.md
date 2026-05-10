---
name: write-test
description: Write a new test for code that lacks coverage
phase: test
user_invocable: true
aliases: [add-test]
origin: new
---

# /write-test — author a new test

A function, behaviour, or contract is uncovered. Write the test
that asserts it.

## Steps

1. **Identify what to assert.** A test exists to catch a future
   bug. Ask: "what would silently break this code?"
   - Happy path with realistic input.
   - Boundary inputs (empty, max-size, off-by-one).
   - Failure modes (invalid input, downstream failure, timeout).
   - Concurrency / ordering, if the code is concurrent.
   - Documented invariants.
2. **Read the existing tests** in the same package/module to
   match conventions: file naming, function naming, fixture
   layout, assertion style, mock vs real-dependency patterns.
3. **Pick the smallest dependency you can get away with.** Real
   over mocked when feasible (real DB in integration tests, real
   filesystem via `t.TempDir()` rather than mocks).
4. **Author the test.** One assertion per test (table-driven
   tests count as one test per row). Clear name that reads as a
   sentence: `TestParse_RejectsEmptyInput`, not `TestParse2`.
5. **Run it both ways**: confirm it passes against the current
   code (or fails meaningfully if you're driving with a bug).
   Then mutate the code under test minimally — does the test
   catch the mutation? If not, the test is too weak.
6. **Run the full test suite** to confirm you haven't depended
   on global state that breaks other tests.

## Conventions

- **Test isolation.** Use `t.TempDir()` or equivalent — never
  write to the working directory.
- **No global mutable state** in the test. If the code under test
  has it, that's a refactor opportunity; flag it.
- **No sleeping** to wait for things. Use signals,
  `WaitForEventually`, or polling with a deadline.
- **Failure messages explain expected vs actual**, not just "got
  unexpected value".

## What NOT to do

- Test the implementation, not the contract. If the code's
  internals change but the contract holds, the test should
  still pass.
- Mock things that should be real. Mocked tests passed for many
  projects right up until the production migration broke.
- Skip the "mutate-the-code-and-verify-it-fails" step. A test
  that always passes is no test.
- Add tests for trivial wrappers / passthrough code unless the
  project explicitly requires coverage everywhere.

## Output shape

```
## Test added
<file:test-function-name>

## What it asserts
- <one sentence>

## How it would catch a regression
- <example mutation that would now fail>
```
