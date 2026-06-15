<!-- Copyright 2026 The dhnt Authors. Licensed under Apache-2.0. -->

# dhnt paper (arXiv draft)

`dhnt.tex` is a self-contained LaTeX draft of the dhnt skills paper, targeting
**arXiv (primary cs.AI; cross-list cs.SE, cs.PL)**.

## Build

```sh
pdflatex dhnt && pdflatex dhnt   # run twice to resolve references
```
Only standard packages are used (geometry, amsmath, booktabs, enumitem, listings,
hyperref). No external `.bib`; the bibliography is inline.

## Status

This is a **design + reference-implementation draft**. Sections 1–9 and 11–12 are
complete prose; the **Evaluation (§10)** reports the results currently
demonstrable from the implementation and marks the controlled cross-model study
(E1–E5) as `[tbd]`. To move from preprint to a peer-reviewed venue, fill E1–E5
with real numbers (see `../docs/positioning-and-standardization.md` §4).

## Before submitting

- **Endorsement:** since Jan 2026 arXiv requires a personal endorsement from an
  established author in the domain (cs.AI) for first-time submitters — line one
  up early.
- **Verify citations:** the 2025–2026 arXiv ids were collected from a web scan;
  confirm exact identifiers and add DOIs/venues before submission.
- **Authors/affiliation:** fill in the author block.
- **Numbers:** replace every `[tbd]` with measured results.
