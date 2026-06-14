<!--
Copyright 2026 The dhnt Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# The dhnt Skill-CNL Specification

> Status: **draft / RFC.** This document specifies the procedure language
> layered on the dhnt interlingua (see `../dhnt.md`). It extends — never
> replaces — the machinery already shipping in this package
> (`ast.go`, `parse.go`, `linearise.go`, `glossary.go`). Every construct
> here preserves the existing transpilability invariant and the roundtrip
> tests in `roundtrip_test.go`.

## 0. What this language is for

Author a procedure **once** so that it is, at the same time:

- **free-style for humans** — writable and readable in any natural language;
- **unambiguous to agents** — one canonical machine identity;
- **close to deterministic, like code** — lowerable to an executable leaf.

Then hand it to *any* executor — a frontier model, a weak model, or a dumb
runtime — and get results that are **consistently close**. This is the
SOP / runbook / playbook spectrum unified into one artifact that is
simultaneously human-readable, machine-executable, and **verifiable**.

The load-bearing idea: **the contract, not the steps, is the spine.** Prose
can be read a hundred ways; a contract gate is passed exactly one way. That
single inversion is what lets a weak model and a strong model converge.

## 1. Relationship to the existing layers

This package already defines (do not rebuild):

```
L0   free prose, any natural language            (author-facing)
  ↓  LLM normaliser (out of scope for this pkg)
L1   glossary-locked CNL in a source language     LineariseLang / [L1 parser]
  ↓  glossary lookup + dhnt encode
L1.5 dhnt canonical form (a-z + spaces only)      LineariseDhnt / ParseDhnt
  ↓  parse
L2   typed AST keyed by dhnt identifiers          Skill / Step / Arg / Expr
  ↓  interpret + dispatch (out of scope for this pkg)
L3   Wasm leaves + AST orchestrator
```

**Validity is transpilability** (unchanged): an expression is valid iff it
round-trips cleanly through L1.5 — `ParseDhnt(LineariseDhnt(x)) == x`, output
strictly `[a-z]` + spaces. Every new construct below obeys this rule. Free
prose, which is *not* a-z-clean, never enters L1.5 (see §7, Intent).

All identifiers remain dhnt-canonical glossary keys; all new structural
keywords are themselves well-formed dhnt words (the package's "no out-of-band
syntax" rule). The dhnt form of every new keyword below was produced by
`dhnt.EncodeWord` and is kept in sync by the existing
`TestGlossary_DhntFormsMatchEncoder` property test.

## 2. Design pillars (normative)

| # | Pillar | One-line guarantee |
|---|---|---|
| **P0** | One identity, many faces | canonical L1.5 form **is** the content address |
| **P1** | Outcome-first | a procedure is a **contract**; steps/code are optional implementations judged by it |
| **P2** | Gradual formalization | every increment of formality buys a stronger guarantee; prose-only is still legal |
| **P3** | Typed, bounded effects | an executor cannot exceed the declared blast radius |
| **P4** | Determinism dial | per-step latitude: exact ↔ judgment |
| **P5** | Verifiable attestation | every run emits a re-checkable receipt |
| **P6** | Composable graph | a primitive may resolve to another procedure (content-addressed) |
| **P7** | Bidirectional | edit any language projection; it re-normalizes to the one identity |

## 3. The AST (L2) — additive extensions

All additions are additive fields on the existing types in `ast.go`. Existing
skills (no contract, no effects, default latitude) linearise byte-identically
to today, so `roundtrip_test.go` keeps passing.

```go
type Skill struct {
    Name      string
    Caps      []string   // existing: needaso block
    EffectCap []Effect   // P3: upper bound on effects this skill may cause
    Contract  []Check    // P1: postconditions; ALL must evaluate to bua
    Steps     []Step     // existing; now formally optional (already roundtrips)
    Flow      []Branch   // P6/playbook: conditional steps
    OnFail    string     // P1: policy glossary id (e.g. abort/retry/blockers)
    IntentRef string     // P7/§7: content hash of the L0 prose (NOT prose itself)
}

type Check struct {            // P1
    Predicate string           // glossary id, EntryKind "predicate"
    Args      []Arg            // typed args (same Arg as steps)
}

type Branch struct {           // playbook
    Cond Check                 // a predicate ⇒ bu
    Then []Step
    Else []Step
}

type Step struct {
    Name      string
    Primitive string            // may resolve to a primitive OR a sub-Skill (P6)
    Latitude  Latitude          // P4; default LatExact ⇒ omitted from L1.5
    Args      []Arg
}

type Effect int                 // P3 lattice
const ( EffRead Effect = iota; EffWrite; EffNet; EffSpend; EffDestroy; EffTime )

type Latitude int               // P4
const ( LatExact Latitude = iota; LatJudge )
```

`Expr` gains one variant so contracts/args can carry truth values directly:

```go
const ( ExprInvalid ExprKind = iota; ExprRef; ExprNumber; ExprBool ) // + ExprBool
func NewBool(b bool) Expr // linearises to "bua"/"bub" (reuses dhnt booleans)
```

## 4. Grammar (L1.5) — extension of the current grammar

Reserved structural keywords (existing + new). All are canonical dhnt; the
parser/lineariser recognise them by literal value.

| Concept | dhnt keyword | status |
|---|---|---|
| skill | `sokilili` | existing |
| needs | `needaso` | existing |
| step | `sotepo` | existing |
| end-of-block | `fini` | existing |
| **ensure (contract)** | `enisure` | new |
| **effect cap** | `efefecato` | new |
| **when (branch)** | `wuheni` | new |
| **else** | `elise` | new |
| **on-fail** | `onifaili` | new |
| **latitude (reserved arg name)** | `latitude` | new |

Effect atoms (closed set, `EntryKind: effect`):
`reada` read · `wurite` write · `neto` net · `sopenida` spend ·
`desotoroyu` destroy · `time` time.

Grammar (EBNF; `{}` = zero-or-more, `[]` = optional). Extends `parse.go`:

```
skill     = "sokilili" name
            { needs | effectcap | ensure | step | branch | onfail }
            "fini"
needs     = "needaso" { cap } "fini"                       (existing)
effectcap = "efefecato" { effect-atom } "fini"
ensure    = "enisure" predicate { argname value } "fini"
step      = "sotepo" name primitive [ "latitude" latval ]
            { argname value } "fini"                       (existing + latitude)
branch    = "wuheni" predicate { argname value }
            { step | branch }
            [ "elise" { step | branch } ] "fini"
onfail    = "onifaili" policy "fini"
value     = numeral | boolean | ref
boolean   = "bua" | "bub"                                  (from dhnt.md)
latval    = "exacato" | "judage"
```

Notes that keep parsing unambiguous and roundtrip-safe:

- An argument-pair loop stops as soon as it sees any reserved keyword
  (`fini`, `sotepo`, `wuheni`, `elise`) — same discipline as today's
  `parseStep`, which stops args at `fini`.
- `latitude exacato` (the default) is **omitted** by `LineariseDhnt`, so
  existing skills are byte-identical. Only `latitude judage` is emitted.
- A `Check`/`Branch` predicate is one canonical token (a glossary
  `predicate` entry); multi-word natural-language names map to it via the
  glossary's per-language labels (P7), exactly like primitives today.

## 5. Contract semantics (P1) — the convergence guarantee

A skill **run** is the act of an executor (any tier) attempting to satisfy the
skill. A run is **valid** iff, *after* execution:

1. every `Check` in `Contract` evaluates to `bua` (true); and
2. the observed effect set ⊆ `EffectCap` (§6).

Steps and `Flow` are *implementations* — hints for how to reach the
contracted end-state. Different executors MAY take different paths:

- a frontier model may ignore the steps and synthesise better ones;
- a weak model follows the steps literally;
- a code/Wasm leaf is the deterministic limit.

All three are judged **only** by (1) and (2). This is the Terraform/k8s
reconcile model applied to agent work, and it is why results converge:
*the spec pins the destination, not the journey.*

`OnFail` policy fires when a `Check` is false at terminal time. Defined
policies (glossary `policy` entries): `aboroto` (abort, default),
`retoroyu` (retry the implementation up to a budget), `balocakeroso`
(write a blockers artifact and exit non-fatally — generalises weave's
`BLOCKERS.md` escape).

> Lineage: this generalises weave's `verify_command` from an opaque
> `bash -c` string into a **typed predicate over the closed glossary**, so
> the check is portable across executor tiers and languages. The shell/Wasm
> that *implements* a predicate lives in L3, per executor.

## 6. Effects (P3) — typed and bounded

Every `primitive` and `predicate` glossary entry declares the effects it may
cause (`Entry.Effects []Effect` in `glossary.go`). A skill declares an upper
bound via `efefecato … fini`. The runtime (L3) enforces:

```
effects(any step actually run)  ⊆  Skill.EffectCap
```

A run that attempts an effect outside the cap is aborted and recorded as
invalid in the attestation (§8) — *before* the side effect lands. This makes
it safe to hand a procedure to an arbitrary, even untrusted, model: the blast
radius is declared, checkable, and enforced, not hoped for. (This is the
weave sandbox contract elevated into the type system.)

Numeric effect bounds (e.g. "spend < 5") are expressed as a contract
`Check` using a comparison predicate (§9), keeping the effect lattice itself
purely qualitative.

## 7. Intent & gradual formalization (P2, P7)

**Intent (L0 prose)** is free natural-language text and therefore cannot be
a-z-clean. It never enters L1.5. Instead the L0 source is stored beside the
catalog entry (markdown body / frontmatter) and the AST carries only
`IntentRef` — the content hash (§8) of that prose. The transpilability and
charset invariants are thus untouched.

**The ladder.** Each rung is independently legal; each buys a stronger
guarantee:

| Rung | What's present | Guarantee |
|---|---|---|
| 0 | prose only (`IntentRef`) | a strong model can interpret; weak guarantee |
| 1 | + `Contract` | results converge (any tier judged by the gate) |
| 2 | + `EffectCap` | results converge **and** stay within blast radius |
| 3 | + `Steps` / `Flow` | weak models get a path to follow |
| 4 | + code/Wasm leaf at `LatExact` | deterministic execution |

Authors are never forced up the ladder; agents simply gain determinism as
they climb. **P7**: the L1 inverse parser (the mirror of `LineariseLang`,
specified as future work) lets a human edit any language projection; the edit
re-normalizes to the one canonical identity.

## 8. Identity & attestation (P0, P5)

**Identity (P0).** The canonical L1.5 string from `LineariseDhnt` is the
procedure's content address. Define a helper (root pkg, over the existing
canonical output):

```
Identity(s Skill) = "h" + dhnt-hex( sha256( LineariseDhnt(s) ) )
```

Equal canonical form ⇒ equal identity ⇒ free dedup, caching, diffing, and a
global, language-neutral namespace of procedures.

**Attestation (P5).** Every run emits a portable receipt (`attest.go`, pure
data, no new deps):

```go
type Attestation struct {
    Skill     string    // Identity(skill)
    Tier      string    // executor tier: model id, "wasm", "builtin", …
    Passed    []string  // predicate ids that evaluated to bua
    Failed    []string  // predicate ids that evaluated to bub
    Effects   []Effect  // observed effect set (must ⊆ EffectCap)
    WorldDiff string    // opaque, executor-supplied summary of changes
    Valid     bool      // (all Checks bua) && (Effects ⊆ EffectCap)
}
```

Anyone holding the skill + the attestation can re-check `Valid`. "Consistently
close" thereby upgrades to "provably within contract."

## 9. Starter glossary additions

Seed into `testdata/glossary.yaml` (alongside the existing entries). dhnt forms
shown are `EncodeWord` output; per-language labels are illustrative.

```yaml
# structural keywords  (en[0] must be a clean a-z word: dhnt == EncodeWord(en[0]))
- { dhnt: enisure,    kind: keyword, labels: { en: [ensure, require], zh: ["确保", quebao] } }
- { dhnt: efefecato,  kind: keyword, labels: { en: [effect, effects],zh: ["副作用", fuzuoyong] } }
- { dhnt: wuheni,     kind: keyword, labels: { en: [when, if],       zh: ["当", dang] } }
- { dhnt: elise,      kind: keyword, labels: { en: [else, otherwise],zh: ["否则", fouze] } }
- { dhnt: onifaili,   kind: keyword, labels: { en: [onfail],         zh: ["失败时", shibaishi] } }

# effect atoms (closed lattice)
- { dhnt: reada,      kind: effect, labels: { all: [read] } }
- { dhnt: wurite,     kind: effect, labels: { all: [write] } }
- { dhnt: neto,       kind: effect, labels: { all: [net] } }
- { dhnt: sopenida,   kind: effect, labels: { all: [spend] } }
- { dhnt: desotoroyu, kind: effect, labels: { all: [destroy] } }
- { dhnt: time,       kind: effect, labels: { all: [time] } }

# starter predicates (EntryKind: predicate; declare their own effects)
# en[0] is a clean a-z word; the hyphenated human name is a secondary synonym.
- { dhnt: gereeni,    kind: predicate, effects: [read], labels: { en: [green, tests-green], zh: ["测试通过", ceshitongguo] } }
- { dhnt: builida,    kind: predicate, effects: [read], labels: { en: [build, build-ok],    zh: ["构建通过", goujiantongguo] } }
- { dhnt: exito,      kind: predicate, effects: [read], labels: { en: [exit, exit-zero],    zh: ["退出码零", tuichumaling] } }
- { dhnt: sigeneda,   kind: predicate, effects: [read], labels: { en: [signed, tag-signed], zh: ["标签已签名", biaoqianyiqianming] } }
- { dhnt: lesoso,     kind: predicate, effects: [read], labels: { en: [less, less-than],    zh: ["小于", xiaoyu] } }  # args: value N, bound N

# policies
- { dhnt: aboroto,      kind: policy, labels: { en: [abort],    zh: ["中止", zhongzhi] } }
- { dhnt: retoroyu,     kind: policy, labels: { en: [retry],    zh: ["重试", chongshi] } }
- { dhnt: balocakeroso, kind: policy, labels: { en: [blockers], zh: ["阻塞记录", zusejilu] } }

# boolean type (predicate results)
- { dhnt: booleani,   kind: type, labels: { en: [boolean, bool], zh: ["布尔", buer] } }
```

`EntryKind` gains `predicate` and `effect` and `policy` alongside the existing
`keyword | capability | type | primitive`.

## 10. Worked example — "cut a release"

The same identity, four faces. (Step bodies elided where obvious.)

**L1 — English projection** (`LineariseLang(s, g, "en")`):
```
skill release  effects read write  ensure tests-green  ensure tag-signed
needs core
when dirty  step stop abort
step tag-it sign-tag
onfail abort
```

**L1 — Chinese projection** (`LineariseLang(s, g, "zh")`) — *same structure,
different surface*:
```
技能 release  副作用 read write  确保 测试通过  确保 标签已签名
需要 core
当 dirty  步骤 stop abort
步骤 tag-it sign-tag
否则时 abort
```

**L1.5 — canonical machine form** (`LineariseDhnt(s)`, a-z + spaces):
```
sokilili release efefecato reada wurite fini enisure gereeni fini
enisure sigeneda fini needaso core fini wuheni dirotoyu sotepo ... fini
onifaili aboroto fini fini
```

This re-parses (`ParseDhnt`) to the identical AST (roundtrip), hashes to one
`Identity`, and projects back to both languages. The contract
(`tests-green ∧ tag-signed`) plus the effect cap (`{read,write}` — note: no
`spend`, no `destroy`, no `net`) gate **every** executor tier identically.

## 11. Invariants this spec MUST preserve

1. **Roundtrip** — `ParseDhnt(LineariseDhnt(x)) == x` for every shape,
   including all new constructs (extend `roundtrip_test.go`).
2. **Charset** — `LineariseDhnt` output stays `[a-z]` + spaces
   (`validateLayer15Charset`); prose lives outside L1.5 as `IntentRef`.
3. **Idempotence** — `LineariseDhnt` is stable across a parse cycle.
4. **Default-omission** — a skill with no contract / no effect cap / default
   latitude linearises byte-identically to today (back-compat).
5. **Glossary sync** — every new keyword's `dhnt:` field equals
   `EncodeWord(primary-en-label)` (`TestGlossary_DhntFormsMatchEncoder`).

## 12. Non-goals / staging

Specified here; built incrementally afterwards (this RFC is spec-only):

- **L0 → L2 normaliser** (constrained-decoded slot-filler) — out of scope, as
  today. This spec only fixes its output contract (the AST above).
- **L3 leaf adapter** (Wasm / builtin predicate & primitive bindings) — out of
  scope; this spec fixes the predicate/effect interface it must honour.
- Suggested build order: **P1 contract** (`ast.go` + `parse.go` + a roundtrip
  test) → **P3 effects** → **P5 attestation** → **P7 L1 inverse parser** →
  **P6 composition resolver**. P0/P2/P4 are doc rules plus tiny helpers.
```
