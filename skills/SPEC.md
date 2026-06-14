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

> **Naming.** We call these **dhnt skills**: the word "skill" every
> developer and agentic tool already understands, qualified by the
> language they are written in. A *dhnt skill* is a machine-checkable
> skill (a typed `Skill` AST with a contract); a regular *skill* is prose
> a model reads. Same word, made precise by the qualifier — no new jargon.

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

## 5a. Determinism — what dhnt makes deterministic (and what it does not)

A regular (prose) skill is instructions a model *interprets*: what it does
varies by provider/model, varies run-to-run (sampling), its success is
model-judged (a wrong-but-plausible result passes silently), and it has no
stable identity. dhnt does **not** make an LLM deterministic. It makes
three *other* things deterministic, and pushes model non-determinism into
explicit, gated, optional places. Determinism lives at three layers:

1. **The artifact — fully deterministic.** Canonical form (L1.5), AST,
   `Identity` (content hash), language projections, and the round-trip are
   pure functions: the same skill yields the same bytes and the same
   identity on any machine, every time (invariants §11). A prose skill has
   no canonical form and no machine equality; a dhnt skill is a stable,
   comparable object.

2. **Verification — fully deterministic.** The contract (predicates ⇒
   `bua`/`bub`), the effect-cap check (`EffectsWithin`), and the
   attestation are deterministic functions of a run's results. *Whether a
   run satisfied the spec* is decided identically for every executor,
   every run, every provider. Success becomes machine-decided, not
   model-judged.

3. **Execution — deterministic iff bound to code.** A step bound to a
   `LatExact` leaf (write a file, run tests, stat a path) is deterministic
   like a shell line. A step bound to a **model** (or marked `LatJudge`)
   is as non-deterministic as that model. dhnt does not change that — it
   **contains and gates** it.

**The precise claim.** dhnt converts *"hope the model interpreted the
prose the same way"* into *"let executors vary, but judge every result
against one fixed, deterministic contract."* Variance in the *how* may
remain when a step is a model call, but it can no longer pass silently:
two models — or two runs — are accepted only if they pass the same gate,
stay within the same effect bound, and emit a re-checkable attestation.
Hence a weak model + dhnt contract ≈ a strong model + dhnt contract *for
the success criterion*: outputs may differ, but "did it meet spec"
converges. The **determinism dial** (P4, `Latitude`) is literally how much
of a skill you move from model leaves to code leaves — the author chooses,
per step, how deterministic the skill is.

| | regular skill (prose) | dhnt skill |
|---|---|---|
| representation / identity | none (text) | deterministic (canonical form + content hash) |
| success criterion | model-judged, implicit | deterministically verified (contract) |
| blast radius (effects) | unbounded, unchecked | deterministically bounded + checked |
| "did it succeed" across models/runs | varies, silent | converges (same gate) |
| the work itself | always model-interpreted | author's choice: code leaf = deterministic, model leaf = gated |

**Two honest caveats.** (a) "Deterministic" means *as a function of (skill,
world-state, bindings)* — a `LatExact` leaf running tests is deterministic
for the same code, but the filesystem/network/tool output is still the
real world, as for any program. (b) The prose→skill normaliser (§7, L0→L2)
is itself a model call and so is non-deterministic *at authoring time* —
but its output is validated (must transpile) and then **frozen as a
deterministic artifact with an identity**. Like an LLM writing code: the
writing varies; once you have the skill, running and verifying it does not.
The non-determinism is captured once, not re-incurred on every use.

## 5b. Runtime self-improvement (adaptation)

A dhnt skill can learn from failure and reuse a verified fix — soundly,
because the contract is a deterministic oracle (`adapt.go`). A prose skill
cannot: it has no stable identity to key a cache on and no oracle to
certify a cached fix, so it re-interprets (and often re-fixes) every run.

`RunAdaptive(skill, env, tier, probes, store, repair)`:

1. **Lookup** — `ContextKey(probes)` fingerprints the environment (OS,
   arch, tool versions). If a variant is cached for `(Identity(skill),
   ContextKey)`, run it and **re-verify** (this also handles drift: a
   stale variant simply fails and is re-learned).
2. **Baseline** — otherwise run the skill as written; if it passes, done.
3. **Repair** — otherwise a model (`Completer`, e.g. an agent CLI) proposes
   a corrected implementation. It is accepted **only if it satisfies the
   ORIGINAL contract within the ORIGINAL effect cap** — the model may
   change the *how*, never the *what* (the verification skill grafts the
   original `Contract`+`EffectCap` onto the candidate's steps, so a model
   cannot "fix" a run by weakening the spec or escalating blast radius).
   A passing variant is cached with its attestation.

Next call in the same context — *even by a different agent sharing the
store* — reuses the verified variant; no re-error, no re-repair. Stores:
`MemStore` and `FileStore` (`<dir>/<skill_id>/<context_key>.json`).

Guardrails (all enforced or tested): promote only contract-verified
variants; effect containment (over-cap repairs are rejected and not
cached); re-verify on read; context-keyed so drift re-learns; bounded
repair attempts. Phase 2 (not yet built): emit the repair as a learned
`when/else` branch folded back into the skill, so it self-heals into a
more general playbook and the fix ships *in the skill* rather than a cache.

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

## 12. Implementation status

All eight pillars are now implemented in this package:

| Pillar | Where |
|---|---|
| P0 identity | `attest.go` `Identity` |
| P1 contract | `ast.go` `Check`, `parse.go`/`linearise.go` `enisure` |
| P2 gradual ladder | §7 (doc rule) + the executor `Env` tiers in `exec.go` |
| P3 effects | `effect.go`, `Skill.EffectCap`, `efefecato` block |
| P4 latitude dial | `ast.go` `Latitude`, reserved `latitude` step arg |
| P5 attestation | `attest.go` `Attestation` / `Attest` / `Consistent` |
| P6 composition | `library.go` `Library` (deps, closure, effect audit) |
| P7 bidirectional | `parse_lang.go` `ParseLang` |
| flow control | `ast.go` `Branch`, `when`/`else` in parse/linearise/exec |

**Flow control (playbook branches).** A `Step` may be a `Branch` —
`when <predicate> … [else …]` — instead of a leaf call. Branches nest,
round-trip through L1.5 and every language projection (in Layer 1 a
branch is closed by `fini`, the one delimited construct there), and the
executor runs the Then or Else arm by evaluating the condition predicate.
This turns a runbook (linear steps) into a playbook (steps + decisions)
while the contract stays the spine. The TUI driver uses it for
expect-with-alternatives: `seen <pattern>` is a non-blocking predicate,
so one skill reacts to whatever a tool shows — an error, an approval
prompt, a clarification request.

**Layer 3 reference executor (`exec.go`).** `Env` binds dhnt
primitive/predicate ids to real Go implementations; `Run(skill, env,
tier)` runs the steps, evaluates the contract against real world-state,
and seals the result into an `Attestation`. `Valid` is computed from the
contract (P1) and effect cap (P3), never asserted by the executor. See
`examples/contract_run`: one skill, three executor tiers (diligent /
lazy / rogue), one verdict each — convergence enforced, not hoped for.

**TUI driver leaf (`skills/tui`).** A real applied executor: it binds the
abstract dhnt primitives `spawn`/`send`/`expect`/`quit` and the
`clean-exit` predicate to a pseudo-terminal (the expect(1) idiom) via
`github.com/creack/pty` (MIT). One tool-agnostic dhnt skill drives any
terminal app — `cat`, `sh`, or an agentic CLI (claude / codex / gemini /
aider / opencode) — by swapping the `Spec` (argv + regex patterns +
inputs + timeout); the skill encodes the *protocol*, the Spec supplies
the *tool*. See `examples/tui_drive` and `skills/tui/tui_test.go`
(drives `cat`/`sh` over a real PTY, contract-verified).

## 12a. Interop with Anthropic Skills (zero-config consumption)

`ExportSkillMD(skill, glossary, meta)` renders a dhnt skill as a standard
`SKILL.md` (YAML frontmatter + markdown body). Dropped into a skills
directory (e.g. `~/.claude/skills/<name>/SKILL.md`), it is usable by ANY
Skills-capable tool with no extra config:

- a generic tool reads `name`/`description`, then follows the body — a
  self-contained rendering of the constraints (allowed effects), the
  steps (branches rendered as if/otherwise), and the success criteria
  (the contract);
- a dhnt-aware runtime sees `executor: cnl` and executes the embedded
  canonical form, emitting an attestation. A tool that doesn't know dhnt
  ignores the unknown executor and just follows the prose.

So the same artifact degrades gracefully across tiers. The reverse
direction — prose `SKILL.md` → dhnt skill — is `Normalise` (§ L0→L2).

## 13. Non-goals / still out of scope

- **L0 → L2 normaliser** (the constrained-decoded slot-filler that turns
  free prose into the AST). This spec fixes its output contract (the AST);
  building it is a separate effort that needs a model in the loop.
- **A Wasm leaf adapter.** `exec.go` proves the leaf *interface* with
  native Go bindings; a sandboxed Wasm backend that *enforces* (not just
  detects) the effect cap before a side effect lands is future work.
- **A wire/storage format** for the content-addressed library (P6 is an
  in-memory graph today).
```
