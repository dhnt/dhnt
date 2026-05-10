# dhnt

> A constructed interlingua designed to unify all languages —
> programming, constructed, natural — into a single normalised form.

This repository holds two things:

1. **The dhnt language specification** — see [`dhnt.md`](./dhnt.md).
2. **The Go reference implementation** — modules under
   `github.com/dhnt/dhnt` (this directory).

## What is dhnt?

dhnt is built from the 26 lowercase ASCII letters `a–z`, organised
into five vowel-group rows that drive a CV-only phonotactic rule. The
language is *tam-less* (no tense, aspect, mood) and inflection-free.
Loan words from any source language transform deterministically into
dhnt-conformant strings, which means programming languages,
constructed languages, and natural languages can all share one
normalised vocabulary.

A few worked examples:

| Source         | dhnt          | abbreviated |
|----------------|---------------|-------------|
| `bash`         | `basohe`      | `bsh`       |
| `cd`           | `cada`        | `cd`        |
| `hello`        | `helilo`      | `hllo`      |
| `thanks`       | `tohanikiso`  | `thanks`    |
| `2018` (year)  | `jubajiahe`   | `bjah`      |

See [`dhnt.md`](./dhnt.md) for the full specification including the
alphabet table, syllable rules, contraction rules, numeral encoding
(decimal `ju`, binary `bu`, hexadecimal `pu`), and the loan-word
transformation rules for English, Chinese (via Pinyin), Esperanto,
and Latin/Cyrillic scripts.

## Why the Go reference implementation?

A spec without a deterministic implementation drifts. This repository
ships both — the spec is the authority, the Go module is the
machine-checkable instance.

The implementation is deliberately minimal: pure-Go, single direct
dependency (`gopkg.in/yaml.v3` for the glossary loader), no CGO, no
build tags. It compiles on Go 1.22+ and runs anywhere Go runs.

## Packages

### `github.com/dhnt/dhnt` — language primitives

```go
import "github.com/dhnt/dhnt"

// Apply the vowel-insertion rule.
out, _ := dhnt.EncodeWord("hello")          // "helilo"
out, _  = dhnt.EncodePhrase("how are you")  // "howu are you"

// Validate a candidate dhnt word.
ok := dhnt.IsCanonical("basohe")            // true

// Round-trip non-negative integers.
enc := dhnt.EncodeDecimal(2018)             // "jubajiahe"
n, _ := dhnt.DecodeDecimal(enc)             // 2018
```

### `github.com/dhnt/dhnt/catalog` — curated SDLC skill catalog

A community-shared, lifecycle-organised catalog of LLM-instruction
skills any agent harness can adopt. Embedded in the binary via
`//go:embed`; consumers pay zero filesystem cost.

```go
import "github.com/dhnt/dhnt/catalog"

for _, s := range catalog.ByPhase("review") {
    fmt.Println(s.Name, "—", s.Description)
}

if s, ok := catalog.Lookup("commit"); ok {
    // s.Body is the markdown instruction the LLM should follow.
    // s.Executor is one of: markdown, builtin, cnl.
    runSkill(s)
}
```

Skills are organised by SDLC phase: discover, plan, build, test,
review, commit, integrate, release, deploy, operate, maintain,
document, onboard. Each entry has YAML frontmatter (name,
description, phase, executor, aliases) plus a markdown body the
LLM follows.

The `executor:` field signals dispatch:

- `markdown` (default): hand the body to the LLM as instructions.
- `builtin`: consumer dispatches to a Go executor it registered
  under the same name (e.g. ycode's `internal/runtime/builtin/`).
- `cnl`: dispatch via the typed-AST machinery in
  `github.com/dhnt/dhnt/skills`. Reserved for the future
  programmatic-skill layer.

`v0.2.0-alpha.3` ships **42 skills across all 13 lifecycle
phases** — discover (3), plan (4), build (5), test (4),
review (3), commit (3), integrate (4), release (3),
deploy (2), operate (4), maintain (2), document (3),
onboard (2). The catalog covers the full software-development
lifecycle from greenfield exploration through to production
operation and maintenance. Future alphas refine and extend
this set rather than fill new phases.

### `github.com/dhnt/dhnt/skills` — multilingual skill CNL

A four-layer pipeline for authoring deterministic skill specifications
that can be written and read in any natural language:

```
[Layer 0]  free prose in any natural language    (author-facing)
   ↓ LLM normaliser  (separable; not in this drop)
[Layer 1]  glossary-locked CNL in source language
   ↓ deterministic glossary lookup + dhnt encode
[Layer 1.5]  dhnt canonical form  (a-z + spaces only)
   ↓ regular parse
[Layer 2]  typed AST keyed by dhnt identifiers
   ↓ interpret + dispatch  (separable)
[Layer 3]  Wasm leaves + AST orchestrator
```

Validity is defined by transpilability: a skill is valid iff it
transpiles cleanly into Layer 1.5. The dhnt encoder *is* the
validator. Layer 1.5 is purely machine-facing — humans never have to
read it; if they want to, they can run the transpiler to check.

```go
import "github.com/dhnt/dhnt/skills"

// The package ships an embedded seed glossary covering the structural
// keywords and a handful of generic primitives. Layer your own
// domain glossary on top with Glossary.Merge.
g, _ := skills.SeedGlossary()

s := skills.Skill{
    Name: "salutoyu",
    Caps: []string{"core"},
    Steps: []skills.Step{{
        Name:      "feritisitu",
        Primitive: "porinito",
        Args:      []skills.Arg{{Name: "value", Value: skills.NewRef("texuto")}},
    }},
}

dh,  _ := skills.LineariseDhnt(s)              // a-z + spaces only
ast, _ := skills.ParseDhnt(dh)                  // back to AST
en,  _ := skills.LineariseLang(s, g, "en")     // human-readable English
zh,  _ := skills.LineariseLang(s, g, "zh")     // human-readable Chinese
```

`examples/release_pipeline` shows the full multi-language roundtrip.

## Quick start

```sh
go get github.com/dhnt/dhnt@latest

go run github.com/dhnt/dhnt/examples/encoder_basic
go run github.com/dhnt/dhnt/examples/release_pipeline
```

Or clone and run the test suite:

```sh
git clone https://github.com/dhnt/dhnt
cd dhnt
go test -race ./...
```

## Status

**Alpha (`v0.1.x`).** The API may change at minor versions. `v1.0`
will be reserved for a stability commitment after at least one
downstream consumer dogfoods the release in production.

The deterministic core ships in this drop:

- dhnt encoder/decoder (CV-syllabic vowel-insertion + ju-prefixed
  numerals)
- closed multilingual `Glossary` with bidirectional lookup
- Layer 2 typed AST
- Layer 1.5 ↔ AST roundtrip
- Layer 1 linearisation per language

Out of scope for v0.1.x:

- LLM constrained-decoded slot-filler (Layer 0 → Layer 2)
- Wasm Component Model leaves (Layer 3)
- Pinyin tonal disambiguation, ISO-9 Cyrillic transliteration,
  Esperanto diacritic mapping (English + toneless Pinyin only for now)
- Agent / tool specification packages built on this machinery

## License

Apache License 2.0. See [`LICENSE`](./LICENSE). The patent grant
matters for spec-track projects — without it, a future contributor
with an undisclosed patent could later assert it against users.

## Contributing

Issues and PRs welcome. The spec is authoritative; if implementation
and spec disagree, the spec wins (file an issue and we'll fix the
implementation). Keep changes focused and add tests for any new
behaviour.

For the rationale behind the architecture (why dhnt as a canonical
machine form, why a closed glossary, why multilingual from day 1),
see the design notes upstream:
[ycode/docs/skill-cnl-rationale.md](https://github.com/qiangli/ycode/blob/main/docs/skill-cnl-rationale.md).
