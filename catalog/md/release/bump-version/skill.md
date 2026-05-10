---
name: bump-version
description: Bump the project version (semver) — pick the right bump based on the changes since the last release
phase: release
user_invocable: true
aliases: [version-bump]
origin: new
---

# /bump-version — bump the project version

Increment the version in the right place(s) based on what changed
since the last release.

## Steps

1. **Find the current version** — typically in one of:
   - `package.json` `version` field
   - `Cargo.toml` `[package] version`
   - `pyproject.toml` `[project] version` or `setup.py`
   - `VERSION` file at the repo root
   - Git tags (`git describe --tags`)
   - Multiple of the above (keep them in sync).
2. **Read the diff since the last release** —
   `git log <last-tag>..HEAD --oneline`. Categorise:
   - **Breaking changes** (removed/renamed public API, changed
     defaults, changed data formats) → MAJOR bump.
   - **New features** (additive, backwards-compatible) → MINOR
     bump.
   - **Bug fixes / docs / internal** only → PATCH bump.
   - For 0.x.y projects, the rules are softer — minor bumps can
     break, patch bumps shouldn't. Match project convention.
3. **Pick the bump**: `MAJOR.MINOR.PATCH` per semver.org. If
   you're under v1.0, the project is signaling instability;
   minor bumps may include breaking changes by convention.
4. **Update every place the version lives.** Search for the
   current version string with `git grep`; expect 1–4 hits.
   Lockfiles update automatically when you bump and run the
   package manager.
5. **Run the build / tests** after the bump — some projects
   embed the version into binaries / generated docs and break
   if the embedded version goes out of sync.
6. **Commit** with a single message: `chore(release): bump
   version to X.Y.Z` or whatever the project's convention is.
   Don't bundle unrelated work.

## Pre-release tags

For alpha / beta / rc cycles, use semver pre-release tags:
`X.Y.Z-alpha.N`, `X.Y.Z-beta.N`, `X.Y.Z-rc.N`. These are valid
semver and most package managers / module proxies handle them
correctly.

## What NOT to do

- Bump the version and ship in the same commit as feature work.
  The version bump is its own boundary.
- Bump MAJOR for a one-line breaking-change to a public API
  unless the project's stability promises require it. Some
  projects defer breaking changes to scheduled major releases.
- Forget the lockfile / docs / changelog. The next skill in the
  release flow (`/draft-release-notes`) catches the changelog;
  the lockfile is your job here.
- Skip tagging — that's `/tag-release`, but the version bump is
  the prerequisite.

## Output shape

```
Bumped: <old> → <new>
Files changed:
- <path>
- <path>
Commit: <sha>
```
