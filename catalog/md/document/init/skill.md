---
name: init
description: Generate or refresh AGENTS.md (a.k.a. CLAUDE.md) — the canonical agent-onboarding doc for this repo
phase: document
executor: builtin
user_invocable: true
aliases: [agents-md, init-agents]
origin: new
---

# /init — generate AGENTS.md

Produce a concise, accurate `AGENTS.md` at the repo root that
brings any agentic coding assistant (Claude Code, Cursor, Codex,
Aider, ycode itself, …) up to speed on the project in one read.

The host harness implements this skill as a deterministic Go
executor (single-shot LLM call with code-graph context). The body
below documents the contract that executor honours and the
behaviour an LLM-driven fallback must reproduce.

## What AGENTS.md should contain

1. **One-paragraph project summary** — what this repo is, in plain
   prose. No marketing.
2. **Quick orientation** — entry points (where execution starts),
   the conversation/runtime loop if there is one, the public API.
   Cite file paths.
3. **First-time setup** — the exact commands a new contributor (or
   agent) runs to get the project building locally.
4. **Build commands** — the canonical `make` / `npm` / `cargo`
   targets, what each one does in one line.
5. **Architecture** — the major components, where they live, how
   they connect. One short paragraph per component, with file
   pointers.
6. **Conventions** — naming, layout, dependency rules, formatting.
   What the project is opinionated about.
7. **Directory boundaries** — read-only zones (e.g. `priorart/`,
   `external/`), peer modules, vendored copies.
8. **References** — pointers to longer docs in `docs/` rather than
   inlining their content.

## What AGENTS.md should NOT contain

- Generated content from `README.md` (link to it instead).
- A full task-list, roadmap, or release plan — those age too fast.
- Subjective claims ("fast", "modern", "best") — agents don't
  benefit and humans skim past them.
- Secrets, hostnames, or per-developer paths.

## Process

1. Read the existing `AGENTS.md` (or `CLAUDE.md`, which is often
   a symlink) if present. Preserve sections the user has marked
   load-bearing; refresh the rest.
2. Use the host's code-graph / repo-map to identify the top
   ~30 files and modules ranked by inbound references.
3. Read `README.md`, the top-level `Makefile` / build config, and
   any existing `docs/architecture.md` or equivalent.
4. Generate the doc in roughly the order above.
5. Keep it short. A working AGENTS.md is dense, not exhaustive.
   Aim for 200–400 lines.

## Tone

Write for a competent reader who has 10 minutes. Be concrete —
"`internal/runtime/conversation/runtime.go`: assemble request →
send to provider → dispatch tool calls → loop" beats "the
conversation runtime orchestrates the agentic flow."
