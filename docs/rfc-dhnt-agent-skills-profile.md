<!-- Copyright 2026 The dhnt Authors. Licensed under Apache-2.0. -->

# RFC: The dhnt Profile for Verifiable Agent Skills

- **Status:** Draft / Request for Comments (v0.1)
- **Intended track:** an optional, back-compatible **profile of the Agent Skills
  (`SKILL.md`) standard**, proposed for discussion under the Linux Foundation
  Agentic AI Foundation. *Not* an IETF wire protocol.
- **Reference implementation:** https://github.com/dhnt/dhnt
- **Authors:** Qiang Li and the dhnt Authors

## Abstract

This document defines the **dhnt profile**: an additive, back-compatible
extension to the Agent Skills (`SKILL.md`) format that lets a skill carry a
machine-checkable success **contract**, a typed **effect** bound, a content
**identity**, and portable run **attestations**. A consumer that does not
implement this profile ignores it and follows the existing prose body; a
conforming consumer additionally verifies and may execute the skill
deterministically. The profile defines the artifact, four conformance levels, a
processing model, a canonical media type, and security considerations.

## 1. Introduction

Agent Skills (`SKILL.md`) standardize a skill as prose a model interprets. Prose
skills are non-deterministic across models and runs, their success is
model-judged, and they carry no stable identity, which prevents sound reuse of a
learned fix. This profile adds the missing *contract* without changing the base
format or breaking existing tools.

### 1.1 Goals

- Be **additive and back-compatible**: a base-format consumer is unaffected.
- Make skill **success machine-verifiable**, identically across executors.
- Make skill **side effects declared, bounded, and auditable**.
- Give skills a **content identity** enabling caching, dedup, and lineage.
- Define **portable attestations** of skill runs.

### 1.2 Non-goals

- Replacing `SKILL.md`, MCP, A2A, or any transport/discovery protocol.
- Mandating a particular execution engine, sandbox, or model.

## 2. Conventions and Terminology

The key words MUST, MUST NOT, REQUIRED, SHALL, SHALL NOT, SHOULD, SHOULD NOT,
RECOMMENDED, MAY, and OPTIONAL are to be interpreted as described in BCP 14
[RFC2119] [RFC8174] when, and only when, they appear in all capitals.

- **Skill:** a reusable procedure expressed as an Agent Skill.
- **Canonical form:** the deterministic `[a-z]`+space serialization of a skill
  (Layer 1.5), defined in Appendix A. The canonical form is the unit of identity.
- **Identity:** `"h"` followed by the lowercase hex SHA-256 of the canonical
  form's UTF-8 bytes.
- **Contract:** a set of typed predicate checks; a run satisfies the contract iff
  every check evaluates true.
- **Effect:** a member of the closed lattice
  `{read, write, net, spend, destroy, time}`.
- **Effect cap:** the upper bound on effects a skill's run MAY cause.
- **Attestation:** a re-checkable record of a run (Section 7).
- **Conforming consumer:** software that implements at least Level 1 (Section 5).

## 3. Relationship to the Agent Skills standard

This profile is signaled by the frontmatter field `executor: cnl`. All other base
`SKILL.md` fields and semantics are unchanged. The profile data lives in (a)
frontmatter keys defined here and (b) a fenced **canonical-form block** in the
body. A base-format consumer that does not recognize `executor: cnl` MUST still
be able to use the skill via its prose body; therefore an authoring tool that
emits this profile MUST also emit a self-contained prose body (Section 6.1).

## 4. The dhnt Skill Artifact

### 4.1 Frontmatter

A dhnt skill `SKILL.md` MUST include:

- `name` (string) — as in the base format.
- `description` (string) — as in the base format; used for discovery.
- `executor: cnl` — marks the profile.

It MAY include base-format fields (`phase`, `aliases`, `user_invocable`, …) and:

- `dhnt-identity` (string) — the skill identity (Section 2). If present it MUST
  equal the SHA-256 of the canonical block.

### 4.2 Canonical form (body block)

The body MUST contain exactly one fenced code block whose first line is the
canonical form (Appendix A). The canonical form MUST be valid per Appendix A
(i.e., it MUST re-parse; validity is transpilability). The media type of this
content is `application/dhnt` (Section 9).

### 4.3 Contract

The canonical form MAY contain zero or more `ensure` clauses. A run is **valid**
only if every `ensure` predicate evaluates to the dhnt boolean true. Predicates
are identifiers resolved by the executor's binding registry (Section 6).

### 4.4 Effects

The canonical form MAY contain one `effects` block declaring the effect cap. A
conforming executor MUST treat the observed effect set of a run as part of the
validity verdict: a run is valid only if observed effects are a subset of the
cap. An executor SHOULD prevent, and MUST at minimum detect and record, an
attempt to exceed the cap.

### 4.5 Steps, branches, composition

Steps are leaf actions; a step MAY carry a `latitude` (`exacato` exact, `judage`
judgement). A `when`/`else` branch expresses conditional flow. A step's primitive
MAY name another skill (composition); a conforming executor that supports
composition MUST enforce that a caller's effect cap contains each callee's.

## 5. Conformance levels

A consumer MUST declare the highest level it implements.

- **Level 0 — Prose consumer.** Ignores the profile; follows the base prose body.
  (Any base-format tool is Level 0 for free.)
- **Level 1 — Verifier.** Parses the canonical form; MUST reject a skill whose
  canonical form is invalid; MAY recompute and check `dhnt-identity`; MAY
  re-check a supplied attestation (Section 7).
- **Level 2 — Executor.** Runs the skill against a binding registry; MUST compute
  the validity verdict from the contract and effect cap (Section 4.3, 4.4) and
  MUST NOT report success unless the verdict is valid; SHOULD emit an attestation.
- **Level 3 — Adaptive executor.** MAY repair a failing run and cache the result,
  but MUST accept a repaired variant only if it satisfies the **original**
  contract within the **original** effect cap, and MUST re-verify a cached
  variant before reuse (Section 8).

## 6. Processing model

A Level 2 consumer processing a dhnt skill:

1. Extract the canonical-form block; parse it (reject if invalid).
2. (OPTIONAL) verify `dhnt-identity` equals SHA-256 of the canonical form.
3. Bind each primitive/predicate identifier to an implementation. Unbound
   identifiers MUST cause the run to fail, not be silently skipped.
4. Execute steps/branches.
5. Evaluate every contract predicate and collect observed effects.
6. Compute the verdict `valid = (all checks true) AND (effects ⊆ cap)`.
7. Report success iff valid; SHOULD emit an attestation (Section 7).

### 6.1 Graceful degradation (REQUIRED of authors)

An authoring tool emitting this profile MUST also render a self-contained prose
body (constraints from the effect cap, steps including branches, success
criteria from the contract) so a Level 0 consumer behaves correctly.

## 7. Attestation

A run attestation is a JSON object with at least:

```
{ "skill": "<identity>", "tier": "<executor id>",
  "passed": ["<predicate>", ...], "failed": ["<predicate>", ...],
  "effects": ["read","write", ...], "valid": true }
```

`valid` MUST be computed as in Section 6, not asserted by the executor. A Level 1
consumer holding the skill and an attestation MUST be able to re-derive whether
the attestation is internally consistent (the recorded evidence implies the
recorded `valid`). Attestations MAY be signed; this profile does not mandate a
signature scheme.

## 8. Adaptation (Level 3)

A Level 3 consumer MAY learn from failure:

- A repaired variant MUST be verified against the original contract and original
  effect cap (a model MUST NOT be permitted to weaken the contract or widen the
  cap).
- A cached variant MUST be re-verified before reuse.
- A learned variant SHOULD be stored host-locally and keyed by
  `(identity, context)`. Promotion of a learned variant to a shared catalog
  SHOULD require human review.

## 9. Media type and registration considerations

- This document requests the media type **`application/dhnt`** for the canonical
  form (`[a-z]` and space; UTF-8).
- The frontmatter `executor` value **`cnl`** denotes this profile.
- The effect lattice (Section 2) is a closed set; new effects require a profile
  revision (Section 11). A registry of standard predicate/primitive identifiers
  MAY be maintained alongside this profile.

## 10. Security considerations

- **Effect containment.** The effect cap bounds a run's blast radius; consumers
  SHOULD enforce it before side effects land (a sandboxed executor), and MUST at
  least detect and record violations. A repaired variant MUST NOT widen the cap.
- **Untrusted skills.** A skill from an untrusted source MUST be treated as
  untrusted code; the effect cap and contract bound but do not eliminate risk.
  Consumers SHOULD run untrusted skills in a sandbox matching the declared cap.
- **Cache / repair poisoning.** A shared adaptation cache is an attack surface;
  consumers MUST re-verify cached variants and SHOULD scope caches by trust
  domain and prefer signed attestations.
- **Attestation trust.** An attestation asserts what an executor observed; a
  verifier re-checks internal consistency but cannot, without a signature and a
  trusted execution base, prove the observations themselves. Treat unsigned
  attestations as advisory.
- **Determinism boundary.** Steps bound to a model are non-deterministic; the
  contract gates them but does not make them deterministic (see the paper, §5a).

## 11. Versioning and extensibility

The profile is versioned. Additive changes (new optional frontmatter, new
predicates) do not require a version bump for existing consumers. Changes to the
canonical-form grammar, the effect lattice, or the identity function are
breaking and MUST bump the profile version. Authoring tools SHOULD record the
profile version they targeted.

## 12. References

### Normative

- [RFC2119] Bradner, *Key words for use in RFCs*, 1997.
- [RFC8174] Leiba, *Ambiguity of Uppercase vs Lowercase in RFC 2119 Key Words*,
  2017.
- Agent Skills (`SKILL.md`) standard, Anthropic, 2025.

### Informative

- Q. Li et al., *dhnt: Verifiable, Self-Healing Agent Skills* (this project's
  paper, `paper/dhnt.tex`).
- Model Context Protocol; Agent-to-Agent; AGENTS.md (Linux Foundation AAIF).
- in-toto / SLSA (attestation analogues).

## Appendix A. Canonical-form grammar (ABNF)

All tokens are space-separated; the whole string is `%x61-7A` (a–z) and SP.
Structural keywords are themselves canonical dhnt words.

```abnf
skill     = "sokilili" SP name *( SP block ) SP "fini"
block     = needs / effects / ensure / step / branch
needs     = "needaso" 1*( SP id ) SP "fini"
effects   = "efefecato" 1*( SP effect ) SP "fini"
ensure    = "enisure" SP id *( SP id SP value ) SP "fini"
step      = "sotepo" SP name SP id
            [ SP "latitude" SP latval ]
            *( SP id SP value ) SP "fini"
branch    = "wuheni" SP id *( SP id SP value )
            *( SP ( step / branch ) )
            [ SP "elise" *( SP ( step / branch ) ) ] SP "fini"
value     = numeral / boolean / id
boolean   = "bua" / "bub"          ; true / false
latval    = "exacato" / "judage"   ; exact / judgement
numeral   = "ju" 1*ALPHALOW        ; decimal, dhnt-encoded
name      = id
id        = 1*ALPHALOW             ; canonical dhnt (no consonant clusters)
ALPHALOW  = %x61-7A
```

## Appendix B. Example

````markdown
---
name: release-pipeline
description: Cut a release; on green tests log success, else report.
executor: cnl
---
# release-pipeline
... self-contained prose body (constraints, steps, success criteria) ...

```
sokilili rilease efefecato reada wurite fini enisure gereeni fini
enisure sigeneda fini sotepo alile porinito value texuto fini
wuheni gereeni sotepo se loge fini elise sotepo so porinito fini fini fini
```
identity: `h643991ef...`
````
