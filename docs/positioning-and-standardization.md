<!--
Copyright 2026 The dhnt Authors
Licensed under the Apache License, Version 2.0.
-->

# dhnt skills: positioning, research, and standardization plan

> Goal: make the dhnt skill *one of its kind* — establish it as frontier work,
> drive wide adoption, and pursue two byproducts: (1) a research paper, (2) an
> open specification on a track toward a recognized standard. This document is
> the plan; it is grounded in a June 2026 SOTA scan (sources at the end).

## 0. One-sentence positioning

**dhnt is a semantic, verifiable layer on top of the Agent Skills (`SKILL.md`)
standard**: it keeps the universal artifact every tool already reads, and adds
a machine-checkable *contract*, typed *effects*, content *identity*,
deterministic *verification/attestation*, and runtime *self-healing*. It is not
a competitor to `SKILL.md`, MCP, or A2A — it is the *upgrade path* for skills,
and it degrades gracefully (a generic tool follows the prose; a dhnt-aware
runtime executes and verifies).

This framing matters because the ground shifted in late 2025: Agent Skills is
*already* the open standard (32+ tools in <90 days). The winning move is not a
rival format — it is to be the **rigor layer inside the format that won**.

## 1. SOTA landscape (June 2026)

- **Agent Skills standard.** Anthropic's `SKILL.md` (YAML frontmatter + Markdown
  + bundled scripts, progressive disclosure) is the de-facto skill format,
  adopted across vendors. Strength: ubiquity, simplicity. Gap: skills are
  *prose* — non-deterministic, unverified, no identity, no effect bound. This
  gap is dhnt's entire thesis (see SPEC §5a).
- **Tool/agent protocols (Linux Foundation AAIF).** MCP (tool invocation,
  JSON-RPC), A2A (agent-to-agent), AGENTS.md, goose — all now under the
  **Agentic AI Foundation**. These standardize *transport and discovery*, not
  *procedure semantics*. dhnt is complementary: it specifies what a *skill* is,
  not how agents talk.
- **Discovery/identity (IETF, W3C).** IETF has competing drafts for agent
  discovery (APIX, DNS-AID); W3C has an AI-Agent-Protocol Community Group.
  Again: discovery/transport, not skill semantics.
- **Skill acquisition (research).** ReAct, Toolformer, Voyager (LLM builds a
  reusable skill library). Gap: acquired skills are unverified and
  re-discovered; dhnt's contribution is *verified, content-addressed,
  contract-gated* skills with sound reuse.
- **Neurosymbolic / Design-by-Contract for agents (hot, 2025-26).** A DbC-
  inspired neurosymbolic layer mediating LLM calls (arXiv 2508.03665);
  neuro-symbolic verification of instruction following (2601.17789);
  neuro-symbolic specification synthesis / autoformalization NL→SMT
  (2504.21061); verified agentic policy generation (VERAFI, 2512.14744). dhnt
  is in this family but goes further: a full *language + interlingua + runtime*,
  not a per-call wrapper.
- **Controlled natural language.** Attempto Controlled English (ACE → DRS →
  FOL/Prolog), Grammatical Framework, Gherkin/BDD. dhnt's CNL is differentiated
  by a content-addressed canonical form, multilingual projection, effect
  typing, and attestation.
- **Verifiable computation / supply chain.** in-toto, SLSA (signed, re-checkable
  attestations of *how* an artifact was produced) — the closest analogue to
  dhnt's run attestation, applied to procedures rather than builds.

## 2. What makes dhnt frontier (the defensible novelty claims)

1. **Skills as contracts, not prose** — Design-by-Contract for agent skills, but
   as a first-class typed AST with a deterministic verifier, not a per-call
   guard. (Differentiates from 2508.03665.)
2. **Content-addressed procedure identity** — "git for skills": a stable hash of
   the canonical form gives cross-model equality, dedup, caching, and lineage.
   Novel for agent skills.
3. **Typed, bounded, *enforced* effects** — an effect lattice on the skill type,
   checked in the attestation; the blast radius is part of the contract.
   (Relates to permission-manifests / pre-action authorization, 2603.20953.)
4. **Determinism, made precise** — the three-layer determinism distinction
   (artifact / verification / execution) and the "converge on the success
   criterion" claim (SPEC §5a). A clean conceptual contribution.
5. **Verifiable attestation for procedures** — portable, re-checkable run
   receipts (in-toto/SLSA, but for skill execution).
6. **Sound runtime self-improvement** — context-keyed, *contract-verified*
   repair caching + branch accretion (SPEC §5b). The contract is what makes
   cached self-healing *sound* — a property prose skills cannot have. This is
   the most novel single piece.
7. **Multilingual interlingua CNL with a canonical form** — one identity, many
   human-language faces, bidirectional.

The throughline (and the paper's thesis): **a deterministic contract is the
missing oracle that turns agent skills from "interpreted prose" into
verifiable, cacheable, self-healing, portable artifacts** — and it can be added
*inside the standard that already won*, so adoption cost is ~zero.

## 3. Adoption strategy (ride the standard, add the rigor)

- **Zero-integration first.** Ship dhnt skills as ordinary `SKILL.md` (via
  `ExportSkillMD`) → instant use by 32+ tools, no tool change. The extra value
  (contract, effects, identity, attestation) is carried in the bundle and
  consumed by a thin external CLI. See `docs/zero-integration-features.md` —
  the concrete list of wins that require little/no change from tools.
- **External-tool ergonomics.** Distribute `dhnt` as a CLI tools shell out to
  (like `eslint`/`prettier`/`git`): `dhnt verify`, `dhnt run`, `dhnt lint`,
  `dhnt attest`. Adoption looks like adding a linter, not an SDK.
- **Killer demos.** (a) cross-model convergence (same skill, weak+strong model,
  same verified outcome); (b) self-healing skill (fix once, reuse across
  agents); (c) drive-any-CLI; (d) effect-cap catches an unsafe "fix".
- **Reference implementation** (this repo) + catalog + glossary growth.
- **Standardization as an Agent Skills *profile*** (see §5), not a rival.

## 4. Research paper plan

- **Working title:** *"dhnt: Verifiable, Self-Healing Agent Skills via a
  Contract-Carrying Controlled Language."* (alt: *"Skills as Contracts: …"*)
- **Venue:** arXiv first — primary **cs.AI**, cross-list **cs.SE** and
  **cs.PL**/**cs.MA**. Note: since Jan 2026 arXiv tightened endorsement — a
  first-time submitter needs a personal endorsement from an established author
  in the domain (or institutional email + a prior accepted paper). *Action:
  line up an endorser early.* Then target a peer-reviewed venue: an agents
  workshop at NeurIPS/ICLR, or a SE/PL venue (ICSE/FSE/OOPSLA) for the
  verification angle.
- **Abstract (draft):** Agent "skills" are today natural-language instructions a
  model interprets; their behavior varies across models and runs, their success
  is model-judged, and they have no stable identity — so a fix discovered on one
  run is re-discovered on the next. We present **dhnt**, a controlled language
  and runtime in which a skill is a *contract-carrying typed artifact*: a
  canonical, content-addressed form with a machine-checkable success contract, a
  typed effect bound, and portable run attestations. We show that the contract
  acts as a deterministic oracle that (i) makes success verification independent
  of the executing model, (ii) bounds and audits side effects, and (iii) enables
  *sound* runtime self-improvement — a verified repair is cached and folded back
  into the skill as an environment-guarded branch, reused across agents without
  re-incurring the failure. dhnt round-trips to the `SKILL.md` standard, so it is
  adoptable with no change to existing tools. [+ headline eval numbers]
- **Structure:** intro & thesis · background/SOTA · the language (layers L0–L3,
  canonical form, identity) · contract + effects + attestation · determinism
  (§5a) · self-healing (§5b) · interop with Agent Skills · implementation
  (Go reference) · **evaluation** · related work · limitations · conclusion.
- **Evaluation needed to be credible (the honest gap).** A design paper won't
  land; we need experiments:
  - **E1 cross-model convergence:** one skill, N models (incl. weak), measure
    success-rate and output variance *with vs without* the contract gate.
  - **E2 determinism/repeatability:** variance across repeated runs, code-leaf
    vs model-leaf.
  - **E3 self-healing:** repair count vs invocations across environments; show
    "learn once, reuse" (we already have the synthetic version — needs a real
    task suite, e.g. cross-platform shell skills).
  - **E4 safety:** effect-cap containment catches unsafe repairs / escapes
    (true/false positive rates).
  - **E5 interop:** generated `SKILL.md` runs across ≥3 standard tools.
  - Baselines: plain prose skills; DbC-wrapper (2508.03665). Metrics: success
    rate, variance, repair amortization, containment recall.

## 5. Standardization plan (and a realistic venue call)

**Recommendation: pursue an *Agent Skills extension/profile* under the Linux
Foundation Agentic AI Foundation — not an IETF RFC.** Rationale:

- The agent-standards center of gravity is **AAIF** (MCP, A2A, AGENTS.md, goose
  all live there). A skill-semantics spec belongs next to Agent Skills, not in
  IETF (whose agent work is wire-level discovery: APIX, DNS-AID).
- dhnt is a *format/language*, not a network protocol — the natural artifact is
  an **open specification + reference implementation + conformance tests**,
  exactly the AAIF/Agent-Skills shape, not an IETF I-D.
- Concretely: propose **`executor: cnl`** (already reserved) plus the canonical
  block as an **optional profile of the Agent Skills standard** — "a SKILL.md
  MAY carry a dhnt canonical form and contract; conforming runtimes MAY execute
  and attest it." This is additive and back-compatible by construction.

**Path:**
1. Freeze `skills/SPEC.md` as a versioned spec (done) + add conformance tests +
   a media type for the canonical form (`application/dhnt`).
2. Publish reference impl + CLI + the zero-integration wins (this repo).
3. Write the Agent Skills *extension proposal* (the "RFC" in the colloquial
   sense) and circulate to the Agent Skills maintainers (Anthropic) and AAIF.
4. *If* a narrow wire-level piece emerges worth IETF (e.g. the attestation
   envelope as a media type / well-known), file a focused I-D then — but that's
   optional and secondary.

**Honest read:** "becoming an official IETF RFC" is the wrong target for a
skill *language*; "becoming the verifiable profile of the Agent Skills
standard, under AAIF" is the achievable, higher-impact goal. arXiv + a workshop
paper builds the credibility that makes the standards proposal land.

## 6. Risks / what's still missing

- **Empirical evidence** (E1–E5) — the biggest gap between "strong prototype" and
  "frontier, cited, adopted."
- **Glossary scale** — the closed glossary is a bottleneck; the loan-word rule
  helps, but a curated + extensible vocabulary process is needed.
- **A real killer app** — kg / weave / cloudbox-style internal use as the first
  production proof points.
- **Ecosystem buy-in** — needs an endorser (arXiv) and a champion inside the
  Agent Skills / AAIF community.

## Sources (June 2026 scan)

- Anthropic Agent Skills open standard & adoption — thenewstack.io,
  agensi.io/learn/agent-skills-open-standard, paperclipped.de.
- MCP under Linux Foundation; **Agentic AI Foundation (AAIF)** forming around
  MCP/goose/AGENTS.md; A2A under AAIF — linuxfoundation.org press release;
  modelcontextprotocol.io 2026 roadmap.
- IETF agent drafts (APIX `draft-rehfeld-apix-core`, DNS-AID); W3C AI Agent
  Protocol Community Group — datatracker.ietf.org; techtimes.com.
- DbC neurosymbolic layer arXiv:2508.03665; neuro-symbolic instruction
  verification arXiv:2601.17789; spec synthesis arXiv:2504.21061; VERAFI
  arXiv:2512.14744; permission/pre-action auth arXiv:2603.20953.
- Attempto Controlled English (ACE) — attempto.ifi.uzh.ch; Grammatical
  Framework; Gherkin/BDD.
- arXiv endorsement policy update (Jan 2026) — blog.arxiv.org.
- Verifiable build attestation analogues — in-toto, SLSA.
