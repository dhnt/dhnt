<!--
Copyright 2026 The dhnt Authors
Licensed under the Apache License, Version 2.0.
-->

# dhnt features that enhance agentic tools with zero / little integration

The adoption thesis: dhnt distributes value two ways that cost tools almost
nothing — (a) a dhnt skill **is** a standard `SKILL.md`, so any of the 32+
Agent-Skills tools use it unchanged; and (b) `dhnt` ships as an **external CLI /
sidecar** a tool shells out to (like `eslint`, `prettier`, `git`) — not an SDK,
not a protocol. The deeper wins need only a few lines.

Tiers by what the tool must do:

- **Tier 0 — zero change** (value rides the `SKILL.md` artifact).
- **Tier 1 — shell out** to the `dhnt` CLI (one command, like adding a linter).
- **Tier 2 — filesystem convention** (point at a shared dir; opt-in).
- **Tier 3 — tiny adapter** (a few lines) for the deepest value.

## Tier 0 — zero change (the artifact carries it)

| Feature | What the tool gets | Tool does |
|---|---|---|
| **Explicit success criteria + constraints in the body** | Even an unaware model behaves better and fails less silently — the skill states the contract and the allowed effects as prose. | nothing (`ExportSkillMD` already wrote it) |
| **Effect / permission manifest** | The skill declares its blast radius ("may only read, write; must not net/spend/destroy"); a tool's existing permission/sandbox UI can surface or enforce it. | nothing (read prose) / optional surface |
| **Multilingual skill body** | The skill renders in the user's language from one source. | pick the projection |
| **Content identity (hash)** | Dedup, cache, "is this the same skill", lineage — for any registry or tool. | read one field |
| **Embedded canonical form (`executor: cnl`)** | Unaware tools ignore it and follow prose; aware ones execute it. Graceful degradation from one file. | nothing |

## Tier 1 — shell out to `dhnt` (linter-level)

| Feature | What the tool gets | Tool does |
|---|---|---|
| **`dhnt lint` / validate** | CI / marketplace catches a `SKILL.md` whose embedded form doesn't transpile — broken skills never ship. | one CI step |
| **`dhnt verify` / `run`** | A *deterministic, model-independent* pass/fail for "did this skill succeed", with an attestation — instead of trusting the model's self-report. | shell out |
| **`dhnt attest`** | A portable run receipt (skill identity, passed checks, effects, tier) for audit / compliance / CI gating. | shell out + store a file |
| **Dependency + effect audit** | A skill registry checks a skill's sub-skill calls and effect containment — supply-chain safety for a skill marketplace. | one check |
| **`dhnt export`** | Author once in dhnt; publish drop-in `SKILL.md` bundles every tool already reads. | build step |

## Tier 2 — filesystem convention (opt-in, cross-tool)

| Feature | What the tool gets | Tool does |
|---|---|---|
| **Shared self-healing overlay** (`~/.dhnt/versions`) | A fix learned by tool A in an environment is reused by tool B — cross-tool, cross-agent learning through a shared directory, no protocol. | point at the dir (or use the default) |
| **Shared attestation / adaptation store** | A common audit trail and verified-variant cache across tools and agents. | shared dir |

## Tier 3 — tiny adapter (a few lines) for the deepest value

| Feature | What the tool gets | Tool does |
|---|---|---|
| **Contract-gated execution** (`Runtime` / `RunAdaptive`) | Cross-model convergence + sound self-healing *inside* the tool: weak and strong models converge on the verified outcome; failures are repaired once and reused. | route `cnl` skills through the runner instead of raw exec |
| **Bring-your-own model** (`Completer`) | dhnt's normalise/repair uses the tool's *existing* model — no new model dependency or key. | implement one function |
| **prose → skill** (`Normalise`) | Turn the user's plain-English request into a verifiable skill on the fly, using the tool's model. | call one function |
| **Drive other CLIs** (`skills/tui`) | A tool that orchestrates other terminal tools (weave-style) gets the expect-with-alternatives PTY driver + the agent-CLI catalog for free. | use the package |

## The one-line pitch per audience

- **A SKILL.md tool:** "Your users' skills now carry success criteria, an effect
  bound, and a content hash — for free; install `dhnt` to also verify them."
- **A skill registry / marketplace:** "Run `dhnt lint`/audit in CI — reject
  skills that don't transpile or that exceed their declared effects."
- **An agent framework:** "Route `cnl` skills through `dhnt run` and your weak
  and strong models converge on the same verified outcome, and fixes are learned
  once and shared across hosts."

Every row above maps to code already in this repo (`ExportSkillMD`, `dhnt`
CLI, `Runtime`/`RunAdaptive`, `FileVersionStore`, `Normalise`, `skills/tui`).
The integration cost ranges from *nothing* (Tier 0) to *a few lines* (Tier 3) —
deliberately, so a large number of tools can adopt with little or no change.
