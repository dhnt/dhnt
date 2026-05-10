---
name: run-tests
description: Run the project's test suite and report results clearly
phase: test
user_invocable: true
aliases: [test]
origin: priorart/aider
---

# /run-tests — run the test suite and report

Find the project's test command, run it, summarise the result. If
tests fail, report the first few failures clearly enough that the
user can act on them.

## Discovery

1. Look for explicit test commands in this order:
   - `make test`, `make test-all`, `make ci` (cited in
     `Makefile`).
   - `npm test` / `yarn test` / `pnpm test` (from
     `package.json`).
   - `go test ./...` / `cargo test` / `pytest` / `mvn test`
     (language defaults).
   - A project-specific script in `scripts/`, `bin/`, or
     `.github/workflows/`.
2. Prefer the project's documented command from `AGENTS.md` /
   `README.md` over guessing.
3. If multiple test tiers exist (unit, integration, e2e), default
   to the unit tier unless the user specified otherwise. The
   slower tiers are explicit invocations.

## Execution

- Run with output streaming so the user sees progress.
- Use the project's preferred verbosity (most prefer concise; CI
  wants verbose).
- Don't reduce coverage by adding `-skip` or `-short` flags
  unless the user asked.
- Honour the project's race-detector default (e.g. `-race` in Go
  if that's the standard locally).

## Reporting

### All green

```
## Tests passed
- Suite: <command>
- Total: <N tests in M packages/modules>
- Time: <duration>
```

### One or more failures

```
## Tests FAILED
- Suite: <command>
- Failed: <count of failed tests>

## First failures
1. <package/module> · <test name>
   <one-line failure summary>
   <relevant assertion line, max 5 lines of stack>

2. <…>

## How to reproduce
<the exact command the user can re-run for one failing test>
```

Cap the failure list at 5 — if more failed, mention the count and
suggest narrowing.

## What NOT to do

- Run only the test you think might pass. Run the project's
  default test command unless the user specified narrower scope.
- Hide failures behind "all tests passed" because the runner
  exited 0 due to a wrapper script error.
- Add new tests as a side effect of running tests.
- Modify the test command to make it pass.
- Quote the entire failure log when 5 lines tell the story.
