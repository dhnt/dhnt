---
name: explore-codebase
description: Get oriented in an unfamiliar repository — what it is, how it's built, where execution starts
phase: discover
user_invocable: true
aliases: [orient, codebase-tour]
origin: new
---

# /explore-codebase — orient yourself in a new repo

You've just landed in a repository you've never seen. Goal: in 10–20
minutes, build enough mental model to answer "what is this, what
language/framework, where does execution start, what's the build
command, where do new contributions land?"

## Steps

1. **Read the README** end-to-end (or skim if it's a novel — note the
   table of contents). Capture: project name, one-line purpose,
   stated audience.
2. **Read `AGENTS.md` / `CLAUDE.md` / `.agents/`** — many projects
   now ship agent-onboarding docs that compress the "what an LLM
   should know" view of the repo. Read these before falling back to
   inference.
3. **Identify the stack** from the manifest files:
   - `go.mod` → Go module + version
   - `package.json` → Node + lockfile (pnpm/yarn/npm/bun)
   - `Cargo.toml` → Rust crate or workspace
   - `pyproject.toml` / `requirements.txt` / `setup.py` → Python
   - `Gemfile` → Ruby; `pom.xml` / `build.gradle` → JVM
   - Multiple manifests → polyglot; note each.
4. **Map the top-level layout** with `ls -la` (or `tree -L 2`).
   Common conventions:
   - `cmd/` → main entry points (Go).
   - `src/` → source root (multi-language).
   - `internal/` → private packages (Go).
   - `pkg/` → public packages (Go) or scripts (general).
   - `lib/` → library code.
   - `test/` `tests/` `__tests__/` → test trees.
   - `docs/` → documentation.
   - `scripts/` → build/deploy helpers.
   - `examples/` `samples/` → runnable demos.
5. **Find the entry point** (use `/find-entry-point` if uncertain).
   Trace from there for ~50 lines: the first function in main.go,
   the export of index.ts, the `if __name__ == "__main__"` block.
6. **Read the build command** from `Makefile` / `package.json`
   scripts / `Cargo.toml` aliases / `pyproject.toml` tooling. The
   build command is often the most accurate "what this project
   actually does" signal.
7. **Skim recent commits** with `git log --oneline -20`. Active
   areas show up here; abandoned corners don't.
8. **Note open questions** to resolve later: anything you can't
   yet explain (a directory whose purpose isn't obvious, a
   build-system dependency you don't recognise, a config flag
   referenced everywhere).

## Output shape

A short brief, optimised for handing to your future self or
another agent:

```
## Project
<name> — <one-line purpose>

## Stack
- Language: <…>
- Framework: <…>
- Package manager: <…>

## Build / test
- `<build command>`
- `<test command>`

## Layout
- <path> — <one-line role>
- <path> — <one-line role>
- (top 5–10 dirs only)

## Entry point
<file:line>

## Recent activity
<2-3 themes from `git log`>

## Open questions
- <thing you couldn't yet explain>
```

## What NOT to do

- Read every file before forming a hypothesis. Skim broadly, deep-
  read narrowly.
- Trust marketing claims in the README without checking them
  against the code.
- Get sucked into refactoring while exploring. Note smells; resist
  fixing them.
- Open more than 30 files. If you need to, you're bisecting badly.
