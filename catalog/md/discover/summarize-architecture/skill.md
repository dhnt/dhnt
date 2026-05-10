---
name: summarize-architecture
description: Produce a concise architecture brief for an unfamiliar codebase — components, data flow, key boundaries
phase: discover
user_invocable: true
aliases: [architecture, arch-summary]
origin: new
---

# /summarize-architecture — write the architecture brief

After `/explore-codebase` and `/find-entry-point` you have the lay
of the land. Now produce the one-page architecture brief: how this
system is shaped, where state lives, what the major boundaries are.

## What an architecture brief contains

1. **Diagram-in-prose** — the major components and how requests/
   data flow between them. ASCII art is fine and often clearer than
   "the X service calls the Y service."
2. **State** — where data lives (DB, cache, in-memory, on disk),
   and what's authoritative vs derived.
3. **Boundaries** — process boundaries, network boundaries, trust
   boundaries. These are where bugs hide.
4. **External dependencies** — what other services / APIs / queues
   this system depends on.
5. **Lifecycle** — startup sequence, shutdown sequence, hot-reload
   if any.
6. **Conventions** — opinions the codebase has that aren't obvious
   from any one file (error handling, logging, dependency
   injection, configuration).

## Process

1. **Read the entry point** thoroughly (~200 lines into the
   process startup).
2. **Identify the major components** by listing top-level
   directories under `internal/` / `src/` / `lib/` / `pkg/` and
   skimming the package docs (Go) / module docstrings (Python) /
   crate docs (Rust). One sentence per component.
3. **Trace one realistic request** end-to-end. For an HTTP server:
   route → handler → service → store → response. For a CLI: arg
   parse → command dispatch → tool call → output. Pick one
   commonly-exercised path and follow every hop.
4. **Identify state stores** by searching for: DB connection
   strings, ORM imports, cache clients, file I/O, in-memory caches
   (sync.Map, threading.Lock, RwLock).
5. **Look at external integrations** — HTTP clients, gRPC stubs,
   message queue subscribers/publishers, observability sinks.
6. **Note the boundaries** explicitly — anywhere a request crosses
   a process / network / trust line.

## Output shape

```
## <project> — architecture brief

## Shape
<2-3 paragraphs of prose, or an ASCII diagram, or both>

## Components
- **<name>**: <one-line role>; lives at `<path>`
- **<name>**: <one-line role>; lives at `<path>`
- (5–10 entries; not exhaustive)

## State
- <store>: <what lives there, authoritative or derived>

## External dependencies
- <service / API>: <what we use it for>

## Lifecycle
- Startup: <one paragraph>
- Shutdown: <one paragraph if non-trivial>

## Conventions worth knowing
- <opinion> — <where it's enforced>
```

## What NOT to do

- Confuse the *intended* architecture (from docs) with the *actual*
  one (from code). Read the code; cite the code.
- Document every package — pick the load-bearing 5–10.
- Speculate about why a design choice was made unless you found a
  comment / commit message / ADR explaining it.
- Editorialise ("this should be refactored", "this is overcomplicated").
  Stick to descriptions.
- Produce a 50-page document. The brief should fit on one screen
  per section.

## When you can't tell

If a chunk of the system is genuinely opaque — generated code,
vendored bundle, a binary blob — say so explicitly:

```
## Opaque regions
- <path> — <why we can't see in>
```

That's more honest than guessing. The next agent or human reader
will appreciate the warning.
