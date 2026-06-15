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
