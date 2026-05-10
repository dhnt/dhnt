---
name: commit
description: Stage and commit local changes with a Conventional Commits message generated from the diff
phase: commit
executor: builtin
user_invocable: true
aliases: [git-commit]
origin: priorart/aider
---

# /commit — generate and apply a Conventional Commits message

Inspect the staged and unstaged changes in the working tree, decide
which files belong in this commit, and write a Conventional Commits
message that explains *why* — not just what.

The host harness implements this skill as a deterministic Go
executor (single-shot LLM call against a bounded diff). The body
below documents the contract that executor honours and the
behaviour an LLM-driven fallback must reproduce.

## Steps

1. **Survey the working tree.** Run `git status` (don't use `-uall`
   on large repos). Read both staged and unstaged diffs. If nothing
   is staged, decide which files to add.
2. **Stage by name, not `-A` or `.`.** Avoid sweeping in unrelated
   work, secrets (`.env`, `credentials.json`), or large binaries.
   If a file looks like it might contain secrets, refuse to stage
   it and surface the concern.
3. **Pick the message scope.** Identify the smallest scope that
   honestly describes the change. Match the prefix style from
   `git log` (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`,
   `chore:`, `build:`, `ci:`, `perf:`, `revert:`).
4. **Write the message.**
   - Subject line ≤ 72 chars, imperative mood, no trailing period.
   - Optional body: 1–3 short paragraphs explaining the *why*. Skip
     for trivial changes (typos, formatting).
   - Don't list the files changed — `git show` does that.
5. **Commit.** Use a heredoc to avoid quoting issues. Don't pass
   `--no-verify`, `--amend`, or skip hooks unless the user asked.
6. **Verify** with `git status` after the commit.

## Tone

Subject line states the change as if completing the sentence "if
applied, this commit will…". Body explains motivation, not
mechanics. Honest about scope: a bug fix isn't a feature; a
refactor isn't a rewrite.

## What this skill must NOT do

- Add the `Co-Authored-By:` trailer unless the user explicitly
  asked for it.
- Force-push, amend a published commit, or skip hooks.
- Stage pre-existing modifications that weren't part of the
  conversation's work.
- Push to a remote — that is a separate explicit instruction.

## When to call out instead of committing

- Staged changes look unrelated; ask which subset to commit.
- A pre-commit hook fails — fix the underlying issue, then make a
  *new* commit (do not `--amend` after a hook failure; the previous
  commit didn't happen and `--amend` would target the wrong one).
- The diff exceeds the executor's input budget; ask the user to
  split the commit.
