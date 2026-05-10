---
name: update-deps
description: Update project dependencies to current versions — selectively, with verification, not all at once
phase: maintain
user_invocable: true
aliases: [bump-deps, upgrade-deps]
origin: new
---

# /update-deps — update dependencies

Bring the project's dependencies up to date. Carefully — the
"update everything" sweep is how you ship surprises.

## Steps

1. **List outdated dependencies** using the package manager:
   - `npm outdated`, `yarn outdated`, `pnpm outdated`
   - `go list -m -u all`
   - `cargo outdated`, `cargo upgrade --dry-run`
   - `pip list --outdated`, `uv pip list --outdated`
2. **Categorise** each outdated dep:
   - **Patch** (`X.Y.Z` → `X.Y.Z'`) — safe-by-spec; ship in
     batches.
   - **Minor** (`X.Y.0` → `X.(Y+1).0`) — usually safe; read the
     changelog before bumping.
   - **Major** (`X.0.0` → `(X+1).0.0`) — breaking by spec;
     dedicated PR per major bump.
   - **Pre-release / transitive** — skip unless specifically
     needed.
3. **Read the changelog** for each non-patch bump. If a dep has
   no changelog, that's a yellow flag; check release notes / git
   tags instead.
4. **Bump in batches** by category, not all at once:
   - PR 1: all patch bumps (one PR, easy review).
   - PR 2..N: minor bumps grouped by area.
   - PR per major bump (substantial review needed).
5. **Update the lockfile** (`package-lock.json`, `pnpm-lock.yaml`,
   `Cargo.lock`, `go.sum`, `poetry.lock`). Don't hand-edit; let
   the package manager regenerate.
6. **Run the full test suite + linter** after each batch.
7. **Run the build** — some bumps surface only at compile or
   bundle time.
8. **Smoke-test** anything the deps touch heavily — auth,
   serialisation, networking, crypto.

## What NOT to do

- Bump every dep in one PR. The diff becomes unreviewable; if
  one dep is broken, the bisect is painful.
- Skip the lockfile commit. The next CI run will regenerate it
  and your changes look broken.
- Bump major versions without reading the migration guide. Most
  surprises are documented.
- Override a transitive-dep version without understanding why
  it's pinned. Pins exist for reasons.
- Bump security-advisory deps to a version that's also outdated.
  Bump to the latest fixed.

## Security advisories

If `npm audit`, `cargo audit`, `govulncheck`, etc. report a
known CVE, that's a higher-priority bump:

1. Verify the advisory affects this project (severity, attack
   path, exploitability).
2. Bump to the lowest version that includes the fix.
3. If no fix is available upstream, look for a patch / fork or
   add a runtime mitigation; don't ignore.

## Output shape

```
Updated: <N> dependencies (<patch>P / <minor>m / <major>M)

## Patch bumps
- <name>: <old> → <new>

## Minor bumps
- <name>: <old> → <new> — changelog: <one-line summary>

## Major bumps (separate PRs)
- <name>: <old> → <new> — migration: <link>

## Tests
- Suite: <pass / N failed>

## Security advisories addressed
- <CVE-id>: <name> bumped to <version>
```
