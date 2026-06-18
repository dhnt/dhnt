# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

Two things ship together:

1. **The dhnt language spec** (`dhnt.md`) — a constructed interlingua over the 26 lowercase ASCII letters `a–z`, organised into five vowel-group rows that drive a CV-only (consonant-vowel) phonotactic rule. It is TAM-less and inflection-free; words from any source language transform deterministically into dhnt-conformant strings.
2. **The Go reference implementation** (`module github.com/dhnt/dhnt`) — pure Go, two deps (`gopkg.in/yaml.v3`, `github.com/creack/pty`), no CGO, no build tags, Go 1.22+.

The spec is the authority; the Go module is the machine-checkable instance. When the encoder and the spec disagree, that is a bug to surface, not to paper over.

## Commands

```sh
go build ./...                       # build everything (probe binary is a checked-in artifact, ignore it)
go vet ./...                         # CI runs this
go test -race -count=1 ./...         # full suite, race detector on (matches CI)
go test -race ./skills/...           # one package tree
go test -race -run TestEncodeWord .  # a single test (root package)
go run ./examples/encoder_basic      # CI smoke-runs this
go run ./examples/release_pipeline   # CI smoke-runs this; full multi-language roundtrip
```

CI (`.github/workflows/ci.yml`) runs build + vet + `go test -race -count=1 -v ./...` against Go 1.22, 1.23, and stable, then smoke-runs the two examples above. Keep both examples runnable.

The `examples/` directory is the integration-test surface — each subdir is a runnable `main.go` demonstrating one capability (conductor, tui_drive, aider_drive, go_verify, eval_*, contract_run, normalise, …). Prefer adding an example over prose when demonstrating a new feature.

The `dhnt` CLI (`cmd/dhnt/main.go`) has subcommands: `export`, `run`, `normalise`, `promote`, `verify`, `commit`, `bump`, `conductor`.

## Architecture: the four-layer pipeline

The whole system is one pipeline from natural-language prose down to a content-addressed canonical form and back up to a typed, executable AST. Understand this before editing `skills/`:

```
[Layer 0]  free prose, any natural language        (author-facing; LLM normaliser, mostly out of scope)
   ↓
[Layer 1]  glossary-locked CNL in a source language (deterministic glossary lookup)
   ↓
[Layer 1.5] dhnt canonical form — a-z + spaces ONLY (the validator IS the dhnt encoder)
   ↓
[Layer 2]  typed AST keyed by dhnt identifiers       (skills.Skill / Step / Arg / Expr)
   ↓
[Layer 3]  executor leaves bind dhnt ids to the real world (skills/dev, skills/tui)
```

**Validity is defined by transpilability**: a skill is valid iff it linearises cleanly into Layer 1.5. `LineariseDhnt` ↔ `ParseDhnt` is the roundtrip that must always hold; `roundtrip_test.go` guards it.

### Package map

- **`github.com/dhnt/dhnt`** (root: `encode.go`, `numeral.go`) — language primitives. `EncodeWord`/`EncodePhrase` apply vowel-insertion; `IsCanonical` validates; `EncodeDecimal`/`DecodeDecimal` round-trip integers via the `ju`-prefixed numeral encoding. Everything downstream depends on `IsCanonical`.
- **`catalog/`** — a curated, SDLC-phase-organised catalog of 42 LLM-instruction skills across 13 phases, embedded via `//go:embed` from `catalog/md/<phase>/<name>/skill.md`. Each skill has YAML frontmatter (name, description, phase, executor, aliases) + a markdown body. The `executor:` field is the dispatch signal: `markdown` (hand body to the LLM), `builtin` (consumer's Go executor), `cnl` (the `skills/` AST machinery). Pure data + lookup (`ByPhase`, `Lookup`); no execution here.
- **`skills/`** — the CNL itself (Layer 1–2). Core types in `ast.go` (`Skill`, `Step`, `Arg`, `Check`, `Branch`, `Effect`, `Latitude`, `Policy`). `glossary.go` = closed bidirectional label↔dhnt lexicon (YAML); `linearise.go`/`parse.go`/`parse_lang.go` = the transpilers; `attest.go` = identity (content hash) + attestation receipts; `exec.go` = `Env`/`Run` (bind dhnt ids to `PrimitiveFn`/`PredicateFn`, execute, attest); `policy.go`/`rounds.go` = drivers (`RunPolicy` honours `OnFail`; `RunRounds` = bounded "goal-oriented until done"); `library.go` = content-addressed skill composition (P6). `seed.go` ships the embedded seed glossary.
- **`skills/dev/`** — a **Layer 3 executor leaf** binding the `run` primitive + `exit-zero`/`empty-output`/`judged` predicates to real process/filesystem execution. `conductor.go` is a goal-oriented orchestrator that drives agent CLIs (via `ycode weave …`); `goverify.go`, `gitskills.go`, `judge.go` are domain leaves.
- **`skills/tui/`** — a Layer 3 leaf driving PTY/terminal apps (the `expect(1)` idiom): binds `spawn`/`send`/`expect`/`quit` to a real pseudo-terminal. Drives agentic CLIs (claude/codex/gemini/aider/opencode) by swapping the `Spec`.
- **`eval/`** — convergence experiments (does the same skill reach the same contracted end-state across executor tiers?).

### The seven design pillars (P0–P6)

These are normative — every type and design decision in `skills/` maps to one. See `skills/SPEC.md` §2. Cite them in commit messages and code comments as the codebase already does:

- **P0** One identity, many faces — the canonical L1.5 form *is* the content address (`Identity`).
- **P1** Outcome-first — a procedure is a **Contract** (postconditions, all must hold); Steps/code are optional implementations judged only by the contract. This is the spine.
- **P2** Gradual formalization — prose-only is still legal; each increment of formality buys a stronger guarantee.
- **P3** Typed, bounded effects — an executor cannot exceed the declared `EffectCap` (`{read, write, net, spend, time, destroy}` lattice).
- **P4** Determinism dial — per-step `Latitude` from exact (deterministic command) to judge (bounded judgement call).
- **P5** Verifiable attestation — every run emits a re-checkable `Attestation` receipt.
- **P6** Composable graph — a primitive may resolve to another skill (content-addressed sub-call via `Library`).

## Conventions

- **The Spec / Skill split is the core idiom.** A dhnt skill encodes the language-neutral *protocol* using abstract dhnt CV-atom command ids; the concrete argv, regexes, timeouts, and verifier live in a runtime `Spec` (e.g. `dev.Spec`, `tui.Spec`, `ConductorSpec`). This is what keeps the content-addressed skill portable. Never bake repo-specific or tool-specific strings into a `Skill` — put them in the `Spec`.
- **Command ids inside skills are canonical dhnt CV atoms** (e.g. `cmdPlan = "pa"`), not English. The English-readable name comes from the glossary, not the AST.
- **The executor never decides validity.** `Run` computes `Valid` purely from the contract (P1) and the effect cap (P3); a missing primitive/predicate binding is a hard error, not a silent skip.
- Apache-2.0 license header on every Go file ("Copyright 2026 The dhnt Authors").
- Spec changes that touch the AST must preserve the L1.5 roundtrip and the invariants in `skills/SPEC.md` §11.

## Files to ignore

`probe` is a checked-in compiled binary; `.aider*` artifacts are gitignored. Neither is source.
