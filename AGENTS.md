# Repository Guidelines

## Authoritative Specs

`dhnt.md` defines the language; `skills/SPEC.md` defines the Skill-CNL. Both win over implementation. If spec and code disagree, surface the mismatch — do not paper over it.

The module is `github.com/dhnt/dhnt`, pure Go, deps: `gopkg.in/yaml.v3`, `github.com/creack/pty`. No CGO, no build tags, Go 1.22+.

## Files to Ignore

- `probe` — checked-in compiled binary, not source.
- `.aider*` — gitignored AI assistant artifacts.

## Build, Test, and Development Commands

```sh
go build ./...                       # build all packages
go vet ./...                         # CI runs this
go test -race -count=1 ./...         # full suite (CI uses -v too)
go test -race ./skills/...           # one package tree
go test -race -run TestEncodeWord .  # single test (root package)
go run ./examples/encoder_basic      # CI smoke-runs this
go run ./examples/release_pipeline   # CI smoke-runs this; multilingual roundtrip
go run ./cmd/dhnt verify             # dogfood: runs gofmt + vet + tests for a Go project
```

CI (`.github/workflows/ci.yml`) builds + vets + tests with `-race -count=1 -v` on Go 1.22, 1.23, and stable, then smoke-runs the two examples above. Keep examples runnable — they are the integration-test surface.

## Architecture: Four-Layer Pipeline (skills/)

Understand this before editing anything in `skills/`:

```
[Layer 0]  free prose, any natural language          (author-facing; LLM normaliser, out of scope)
  ↓
[Layer 1]  glossary-locked CNL in source language     (LineariseLang)
  ↓  dhnt encode
[Layer 1.5] dhnt canonical form — a-z + spaces ONLY   (LineariseDhnt ⇄ ParseDhnt must roundtrip)
  ↓
[Layer 2]  typed AST keyed by dhnt identifiers        (Skill / Step / Arg / Expr / Contract)
  ↓
[Layer 3]  executor leaves bind dhnt ids to real work (skills/dev/, skills/tui/)
```

**Validity = transpilability.** A skill is valid iff it linearises cleanly into Layer 1.5. `roundtrip_test.go` guards the invariant.

## Critical Design Idioms

### Spec / Skill split

A dhnt `Skill` encodes the *protocol* using abstract CV-atom command ids (e.g. `cmdPlan = "pa"`). The concrete argv, regexes, timeouts, and verifier live in a runtime `Spec` (e.g. `dev.Spec`, `tui.Spec`, `ConductorSpec`). Never bake repo-specific or tool-specific strings into a `Skill` — put them in the `Spec`.

### The seven design pillars (P0–P6)

Cite these in commit messages and design decisions:

- **P0** One identity, many faces — the canonical L1.5 form *is* the content address (`Identity`).
- **P1** Outcome-first — a procedure is a **Contract** (postconditions); Steps are optional implementations.
- **P2** Gradual formalization — prose-only is still legal; each increment of formality buys a stronger guarantee.
- **P3** Typed, bounded effects — an executor cannot exceed the declared `EffectCap` ({read, write, net, spend, time, destroy}).
- **P4** Determinism dial — per-step `Latitude` from exact to judge.
- **P5** Verifiable attestation — every run emits a re-checkable `Attestation` receipt.
- **P6** Composable graph — a primitive may resolve to another skill (content-addressed sub-call via `Library`).

### Executor never decides validity

`Run` computes `Valid` purely from the contract (P1) and effect cap (P3). A missing primitive/predicate binding is a hard error, not a silent skip.

## Package Map

- **Root** (`encode.go`, `numeral.go`) — `EncodeWord`/`EncodePhrase`, `IsCanonical`, `EncodeDecimal`/`DecodeDecimal`. Everything downstream depends on `IsCanonical`.
- **`skills/`** — Layer 1–2. `ast.go` (types), `glossary.go` (bidirectional YAML lexicon), `linearise.go`/`parse.go`/`parse_lang.go` (transpilers), `exec.go` (Env/Run), `policy.go`/`rounds.go` (drivers), `library.go` (P6 composition), `attest.go` (identity/attestation).
- **`skills/dev/`** — Layer 3: binds `run`, `exit-zero`, `empty-output`, `judged` to real process/filesystem. `conductor.go` is the goal-oriented orchestrator.
- **`skills/tui/`** — Layer 3: PTY driver binds `spawn`/`send`/`expect`/`quit`.
- **`catalog/`** — 42 embedded SDLC skills from `catalog/md/<phase>/<name>/skill.md`. YAML frontmatter + markdown body. `executor:` field: `markdown` | `builtin` | `cnl`.
- **`cmd/dhnt/`** — CLI subcommands: `export`, `run`, `normalise`, `promote`, `verify`, `commit`, `bump`, `conductor`.
- **`examples/`** — Each subdir is a runnable `main.go`. Prefer adding an example over prose.

## Coding Conventions

- All Go files must start with the Apache 2.0 header: `// Copyright 2026 The dhnt Authors` followed by the license block (copy from `doc.go`).
- `gofmt` formatting; goimports-compatible imports.
- Command ids inside ASTs are canonical dhnt CV atoms (e.g. `"pa"`), not English labels. Human-readable names come from the glossary.
- Prefer table-driven tests. Tests go in `*_test.go` files beside the code.
- Key invariants: `LineariseDhnt` ⇄ `ParseDhnt` roundtrip, effect cap enforcement, deterministic catalog lookups.

## Commits

Conventional Commit subjects: `feat(scope): ...`, `fix(scope): ...`, `chore: ...`, `dev: ...`. Mention pillars (P1, P6) when they explain the change. PRs note spec/API impact and tests run. CLI output/screenshots only when changing user-visible behavior.
