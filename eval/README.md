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
each arm on five columns: (a) surface (the exact words the arm asked for:
canonical for arm A, conventional for arm B), (b) equivalence via
`EncodeWord` (`ls` ~ `liso`), (c) parser acceptance via `skills.ParseDhnt`,
(d) semantic correctness via AST `Identity`, (e) exact preservation of flags,
paths, and quoted text. The suite covers shared expansions (`bash`/`bsh`), an unseen
nonce word (`wug`), a homonym case (`liso` as predicate vs step), `aaa`
(canonical as-is), the `ls`/`cd` basics, and a rewriting task with literal
preservation.

Column (c) is measured, not assumed: with the S2 parser (which accepts
rule-derived conventional spellings on input) arm B reads `parses: yes`,
and column (b) is likewise expected to pass for both arms. Both columns
stay measured — the harness keeps scoring every column separately.
Column (a) is what tells the arms apart when the encoder accepts both: arm A
answering `ls` fails surface but still passes equivalence.

```sh
go test ./eval/                                   # hermetic (stub transcripts only)
```

### Live smoke run (opt-in, skipped by default)

The manager invokes this once on a subscription model. The prompt travels on
the command's stdin; the completion is read from its stdout. This worker did
not run it and used no metered keys.

```sh
go test ./eval/ -run TestSpellingLiveSmoke \
  -spelling-live-cmd "my-model-cli complete" -v
```

#### Smoke run 2 (2026-10-19)

The manager ran all 8 tasks × both arms (16 prompts) once against a small
subscription Claude model, harness at 399d76b. Summary lines:

canonical: 8/8 surface, 8/8 equivalent, 7/8 parse, 6/8 correct, 8/8 preserved
conventional: 8/8 surface, 8/8 equivalent, 6/8 parse, 6/8 correct, 8/8 preserved

The failures were identical across arms — prompt defects, not spelling:

- homonym-liso: the prompt never showed the contract form, so the model
  wrote "sokire liso" (canonical arm) or put "ensure ls" after the step
  (conventional arm). Fixed by adding the ensure form (`enisure
  <predicate> fini` plus a LineariseDhnt-built contract example from a
  real AST, like the worked example) to the shared grammar note for that
  task, identical in both arms.
- preserve-rewrite: the draft line ran into the instruction ("... fini
  fini Put the rewritten skill ..."), and the conventional arm answered
  it saw no conventional draft. Fixed by putting the draft on its own
  line between clear delimiters, identical in both arms.

The earlier run 1 was invalid, not a result: its prompts lacked the
Layer 1.5 grammar teaching, so only 1/16 outputs parsed.

Interpretation: n=8 tasks, one model, one run; both arms reached the
requested surface 8/8 with equal correctness, so this smoke shows no
spelling penalty for conventional input with the current parser and no
measured benefit. It is not evidence for the hypothesis either way.

#### Smoke run 3 (placeholder)

Reserved for the manager's rerun with the fixed prompts. Record the two
summary lines, any remaining failures, and the interpretation here.

