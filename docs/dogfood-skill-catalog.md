<!-- Copyright 2026 The dhnt Authors. Licensed under Apache-2.0. -->

# Dogfooding dhnt: a catalog of high-value skills for real development

Before finalizing the spec/paper/RFC, dhnt should be *used* — authored as real
skills for the day-to-day development of the umbrella projects (cloudbox,
outpost, ycode, bashy, kg, coreutils, sh, bonsai), then refined as those skills
hit real friction. This document catalogs the recurring development procedures
worth turning into dhnt skills, derived from a survey of the umbrella scripts,
per-submodule build/test conventions, and CI/release workflows.

## Selection criteria

A procedure is a good dogfood skill when it is (1) **frequent**, (2) **broadly
applicable** across projects, and (3) **a place where the dhnt contract / effect
cap / self-healing earns its keep over prose** — i.e. the step is easy to get
wrong, varies by environment, or has a silent failure mode a contract would
catch. The last criterion is what makes a skill worth expressing in dhnt rather
than a shell alias.

## Catalog (prioritized)

### Tier A — flagship, build first (universal + high payoff)

| Skill | What it does | Contract (success) | Effects | Applies to | Why dhnt (vs prose) |
|---|---|---|---|---|---|
| **go-verify** | tidy + format + vet + test gate | `fmt-clean ∧ vet-clean ∧ tests-green` | read, write, time | **every Go project** | The *test command differs per repo* (see below) — those differences become env-keyed self-healing branches the skill learns once; the contract catches "forgot to run tests/fmt". |
| **submodule-pin-bump** | cd into submodule → commit → push → return → `git add <sub>` → commit `sync: bump <sub> pin` | submodule HEAD pushed ∧ umbrella pin staged ∧ submodule tree clean | write, net | umbrella | Encodes the documented *#1 footgun* (editing in a submodule then committing from the umbrella root silently loses the edits); the contract detects exactly that. |
| **safe-commit** | format + vet + test, then stage **by name** (never `-A`) + conventional message | tree builds ∧ tests green ∧ commit exists ∧ no `-A`/`.` staging | write | all | Enforces the "stage by name, never `git add -A`" rule (avoids sweeping in secrets/unrelated files) and the format-before-commit gate. |

### Tier B — broadly applicable, frequent

| Skill | What it does | Contract | Effects | Applies to | Why dhnt |
|---|---|---|---|---|---|
| **sync-submodules** | `script/sync.sh` → review → commit `sync: bump submodule pins` | `.sibling-pins` match umbrella SHAs ∧ submodules at origin HEAD | read, write, net | umbrella | Multi-step, easy to leave half-done; contract verifies sibling-pin coherence (the "one identical toolset on every platform" invariant). |
| **bootstrap** | `script/bootstrap.sh` / `bootstrap-siblings.sh` (materialize `../sh`, `../coreutils`, `../nadir`, `../bonsai`) | required siblings present at pinned SHA ∧ project builds | read, write, net | all consumers | The standalone-vs-umbrella sibling materialization is fiddly; the contract is "it builds", reached by an env-keyed branch (in-umbrella vs standalone). |
| **tag-release** | `git tag vX && git push origin vX` → release workflow | tag pushed ∧ release workflow green ∧ expected assets present | write, net | outpost, ycode, bashy | Same shape across three repos with per-repo asset matrices → one skill, per-repo branches. |

### Tier C — narrower, but the strongest "why-dhnt" demonstrations

| Skill | What it does | Contract | Effects | Applies to | Why dhnt |
|---|---|---|---|---|---|
| **cloudbox-ui-rebuild** | rebuild `ui.zip` (`pnpm build && pnpm package:zip`) when `ui/src/**/*.tsx` changed | `ui.zip` newer than the newest changed `.tsx` | read, write | cloudbox | A **documented repeat error** (editing the SPA without refreshing the embedded zip); the contract catches the forgotten step *every* time. |
| **replace-signed-binary** | overwrite a live signed Mach-O safely: `rm` → `cp` → `codesign --force` | target executes (not kernel-killed) | write, destroy | macOS / dragon | A platform-specific quirk (`cp` over a live signed binary kernel-kills it); the env-keyed branch applies only on darwin/arm64, and the contract is "it runs". |
| **clean-path-test** | run tests with a sanitized PATH so the `ycode` shell wrapper doesn't shadow `sh` | tests green | read, time | sh, bashy | The failure is *silent and confusing* (signal-trap tests fail only when the wrapper is on PATH); an env-keyed branch fixes it and the contract confirms. |
| **trace-verify** | `script/verify-tracing.sh` after a telemetry/reverse-proxy change | one `trace_id` spans ≥2 (strict: 3) services | read, net, time | cloudbox, outpost, ycode | Cross-process, hard to eyeball; the contract is a precise, machine-checkable assertion. |

## Why `go-verify` is the flagship dogfood skill

`go-verify` is the same *intent* everywhere — "leave the tree formatted, vetted,
and green" — but the *implementation* differs by repo, which is exactly the
self-healing story:

| Repo | test command (the per-context "fix") |
|---|---|
| coreutils, sh | `go test ./...` (hermetic / full) |
| outpost | `go test ./...` (no `-short`; needs sibling bootstrap) |
| ycode, kg, bonsai, nadir, gfy | `go test -short ./...` (kg adds `-p 4`) |
| cloudbox | `make test` (go-in-docker; bare `go test` breaks on darwin) |
| bashy | `go test ./...` + clean-PATH for `make test-bash` |

Authored once as a contract (`fmt-clean ∧ vet-clean ∧ tests-green`) with the
*generic* steps (`gofmt -w`, `go vet`, `go test`), the skill **learns the right
test invocation per repo as an environment-guarded branch** on first failure
(e.g. on darwin in cloudbox, the bare `go test` fails → repair to `make test` →
folded in, keyed on that context). After dogfooding across the eight projects,
one `go-verify` skill carries the correct invocation for each — a concrete,
in-house demonstration of every dhnt pillar at once.

## What's needed to author these (the gap)

The current seed glossary covers a toy vocabulary (print/log/format, green/
signed). Dogfooding needs a small **dev-ops glossary + a shell-backed executor
leaf**:

- **Primitives** (Env bindings to real commands): `run` (a glossary-named shell
  command), `gofmt`, `govet`, `gomodtidy`, `gotest`, `gitadd`, `gitcommit`,
  `gitpush`, `pnpmbuild`, `codesign`, … — each declaring its effects.
- **Predicates** (contract checks): `tests-green`, `fmt-clean`, `vet-clean`,
  `tree-clean`, `pushed`, `pin-coherent`, `zip-fresh`, `binary-runs`,
  `trace-stitched`.
- These are the same shape as the existing `skills/tui` leaf, but bound to
  process execution / filesystem checks instead of a PTY.

Free-text command strings stay out of the canonical form (a-z only): a step
names a glossary command id; the actual argv lives in the executor's binding
table (the `tui.Spec` pattern), so per-repo specifics are runtime config and the
skill stays portable.

## Dogfooding plan

1. **Add the dev-ops glossary + a `skills/dev` shell-backed executor leaf**
   (primitives + predicates above), mirroring `skills/tui`.
2. **Author `go-verify` first**, run it via `dhnt run --adapt` across all eight
   Go projects, and let it learn each repo's test invocation as a folded branch.
   This alone exercises contract, effects, self-healing, and accretion on real
   work.
3. **Add `submodule-pin-bump` and `safe-commit`** — the highest-frequency,
   highest-footgun git procedures.
4. **Layer in Tier B/C** as friction is hit; the Tier-C skills are the best
   demos of the contract catching real, historically-recurring mistakes (the
   `ui.zip` rebuild, the macOS signed-binary kernel-kill).
5. **Measure** (and feed the paper's E3): for each skill, repairs vs.
   invocations across repos, and whether the contract caught real mistakes that
   prose skills had let through.

## Dogfood findings (running log)

Real signal from running the skills on the umbrella; each finding is a
per-context refinement (configuration or a future learned branch).

- **2026-06-15, `go-verify`/`go-check`:**
  - Building `go-verify` surfaced a real bug in the **core verifier**:
    contract results were keyed by predicate id, collapsing two checks that
    share a predicate but differ in args (vet vs test). Fixed (checkLabel =
    predicate + args). *Dogfooding found a correctness bug synthetic tests
    missed.*
  - `dhnt verify --check` on `gfy` and `nadir` → valid. On `coreutils` →
    invalid, but only because `gofmt -l .` flagged **vendored `priorart/…/
    vendor`** files (the tree coreutils excludes with `grep -v
    '/priorart/'`); vet and tests passed. **Refinement:** the fmt check must
    be *scoped* per repo (exclude `vendor`/`priorart`/`external`); this is
    per-repo Spec config today and a good candidate for a learned/config
    branch. Net: `go-verify`'s substance is sound; its default fmt scope is
    too broad for repos with vendored trees. **Resolved:** the default fmt
    commands now exclude `vendor`/`priorart`/`external` (a `find`-based file
    list); `coreutils` then verifies `valid=true`. A dogfood finding became a
    broadly-correct default, verified back on the same repo.
  - Per-repo **test command** confirmed as Spec config (`--test "make
    test"` for cloudbox/ycode); folding it into the skill needs the
    command-ref-branch design (test argv lives in the Spec binding, not the
    skill AST).

## Placement and boundary

Authored skills live host-locally (the adaptation overlay) during dogfooding;
broadly-useful ones can later be promoted (human-reviewed) to a shared catalog.
Keep skill bodies and commits free of proprietary cloudbox source — reference
only public build/test commands (this catalog does the same).
