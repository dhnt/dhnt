# GEMINI.md - dhnt Project Context

## Project Overview
`dhnt` is a constructed interlingua and its Go reference implementation, designed to unify programming, constructed, and natural languages into a single normalized form. It serves as a foundation for a deterministic "Skill-CNL" (Controlled Natural Language) for authoring agentic procedures (SOPs, playbooks) that are both human-readable and machine-verifiable.

### Key Components
1.  **dhnt Language Primitives**: Core encoding/decoding logic for the CV-syllabic language, including numeral encoding (decimal `ju`, binary `bu`, hexadecimal `pu`).
2.  **Skill Catalog (`catalog/`)**: A curated, embedded catalog of 42+ SDLC skills organized by phase (discover, plan, build, test, review, commit, integrate, release, deploy, operate, maintain, document, onboard).
3.  **Skill-CNL (`skills/`)**: A four-layer pipeline for deterministic skill specifications:
    - **L1**: Human-readable projection (English, Chinese, etc.).
    - **L1.5**: Machine-facing canonical form (a-z + spaces only).
    - **L2**: Typed AST keyed by dhnt identifiers.
4.  **Reference CLI (`cmd/dhnt`)**: Tools for exporting, running, normalising, verifying, and orchestrating skills.

### Vocabulary and syntax boundary

dhnt-lang defines canonical vocabulary and word rules; Bash# defines executable
grammar/syntax. Conventional forms follow the current dhnt rules: `ls` expands
to `liso` without a registry of pairs. Canonical forms remain strict; multiple
alternative spellings can share an expansion. Do not require one preferred
alias or describe the distinction as written/spoken forms.

Retain original spelling when exact source reproduction matters. Rule-derived
spelling equivalence is separate from meaning and execution binding. Preserve
encoder behavior, canonical Skill-CNL AST identifiers, serialization and
hashes. Its closed glossary is application-specific, not an exhaustive dhnt
vocabulary. Do not rewrite source strings, flags or paths.
Homonyms sharing a canonical form are permitted; context determines the sense.
Existing runtime name resolution and glossary duplicate checks remain intact.

## Building and Running
The project requires **Go 1.22+**.

### Key Commands
- **Build**: `go build ./...`
- **Test**: `go test -race ./...`
- **CLI Help**: `go run ./cmd/dhnt help`
- **Run Examples**:
    - `go run ./examples/encoder_basic`
    - `go run ./examples/release_pipeline`
    - `go run ./examples/tui_drive`
- **Verify Module**: `go run ./cmd/dhnt verify` (runs gofmt, vet, and tests).

## Development Conventions
- **Pure Go**: No CGO, no build tags. Minimal dependencies (primarily `yaml.v3` and `pty`).
- **dhnt Alphabet**: 26 lowercase ASCII letters `a–z`. Syllables are CV (consonant-vowel) or a single vowel.
- **Deterministic Identity**: Skills are content-addressed; their identity is a hash of their canonical L1.5 form.
- **Outcome-First**: Procedures are defined by **contracts** (post-conditions) rather than just steps.
- **Bounded Effects**: Skills declare an **effect cap** (read, write, net, spend, destroy, time) that is enforced by the runtime.
- **Adaptation**: The runtime supports self-healing via a repair loop (`--adapt`) that learns and caches verified variants for specific host environments.

## Repository Structure
- `catalog/md/`: The source markdown files for the embedded skill catalog.
- `cmd/dhnt/`: Main CLI implementation.
- `examples/`: Usage demonstrations for various layers of the project.
- `skills/`: Core implementation of the Skill-CNL AST, parser, and runtime.
- `dhnt.md`: The authoritative language specification.
- `SPEC.md`: The Skill-CNL specification (within `skills/`).

## Tooling Usage
- **`dhnt export`**: Converts a `.dhnt` skill into a standard `SKILL.md` bundle.
- **`dhnt run`**: Executes a skill against a tool/agent and emits a verifiable attestation.
- **`dhnt verify`**: A specialized driver for Go module health checks.
- **`dhnt conductor`**: A goal-oriented orchestrator that enlists a team of agents to reach a verified goal.
