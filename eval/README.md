<!-- Copyright 2026 The dhnt Authors. Licensed under Apache-2.0. -->

# eval — reproducible harnesses for the dhnt paper

## E1 — cross-model convergence (`convergence.go`)

One skill, several model tiers, one contract. Reports, per tier, the raw model
output and the contract verdict; and the with/without-contract acceptance
summary.

```sh
go run ./examples/eval_convergence                 # hermetic fake tiers
GEMINI_CLI_TRUST_WORKSPACE=true \
  go run ./examples/eval_convergence --real gemini,aider   # live (spends tokens)
go test ./eval/                                     # hermetic assertions
```

**Claim demonstrated.** Raw outputs vary across tiers; a prose skill (no verdict)
would accept all; the contract accepts only the correct tiers, and gives the
*same* verdict for correct-but-differently-phrased answers — i.e. the success
verdict converges on correctness, independent of provider/phrasing.

To produce the paper's E1 numbers, run `--real` across ≥2 providers (incl. a
small/weak model) on a task suite and tabulate verdict variance vs. output
variance. The remaining paper experiments (E2–E5) are described in
`../docs/positioning-and-standardization.md` §4.

## Spelling eval — canonical vs conventional spelling (`spelling.go`)

Tests the hypothesis in `../README.md` ("Familiar spelling and LLMs"): every
English word is already an accepted dhnt form as-is (lowercased), so
conventional spellings (`ls`, `bash`, `cd`) are the familiar authoring
surface while canonical spellings (`liso`, `basohe`, `cada`) are the machine
form. `RunSpelling` runs paired tasks — the same skill-authoring/rewriting
task and the same checks, arm A prompted for canonical spellings, arm B for
conventional ones — against one model function (`prompt -> text`), and scores
each arm on four columns: (a) spelling compliance via `EncodeWord`
equivalence, (b) parser acceptance via `skills.ParseDhnt`, (c) semantic
correctness via AST `Identity`, (d) exact preservation of flags, paths, and
quoted text. The suite covers shared expansions (`bash`/`bsh`), an unseen
nonce word (`wug`), a homonym case (`liso` as predicate vs step), `aaa`
(canonical as-is), the `ls`/`cd` basics, and a rewriting task with literal
preservation.

Column (b) is measured, not assumed: against today's parser arm B reads
`parses: no`, and the same harness shows the gap closing once the parser
accepts conventional spellings (parallel story S2) — no harness change.

```sh
go test ./eval/                                   # hermetic (stub transcripts only)
```

### Live smoke run (opt-in, skipped by default)

The manager invokes this once on a subscription model. The prompt travels on
the command's stdin; the completion is read from its stdout. This worker did
not run it and used no metered keys.

```sh
go test ./eval/ -run TestSpellingLiveSmoke \
  -spelling-live-cmd "my-subscription-cli complete" -v
```

Record the smoke results here (one row per task × arm):

| task | arm | compliant | parses | correct | preserved |
|---|---|---|---|---|---|
| | | | | | |

