---
name: tag-release
description: Tag a release with an annotated git tag and push it to the canonical remote
phase: release
user_invocable: true
aliases: [release, cut-release]
origin: new
---

# /tag-release — tag and push the release

Cut the release tag once the version is bumped, the changelog is
written, and CI is green on the commit you intend to release.

## Pre-flight checks

Before tagging, verify:

1. **Current branch is the release branch.** Usually `main` /
   `master`, but some projects use a `release/X.Y` branch.
2. **CI is green** on the commit you're about to tag. Tag the
   commit, not "HEAD whatever's on it now."
3. **Version-bump commit landed** — `git log -1` should show the
   `chore(release): bump to X.Y.Z` commit (or your project's
   equivalent).
4. **Changelog updated** — `CHANGELOG.md` (or equivalent) has the
   new version's entry.
5. **No local changes** — `git status` clean.
6. **Remote is the canonical one** — `git remote -v` shows the
   right `origin`.

If any of those fail, fix that first; don't tag from a dirty or
ambiguous state.

## Steps

1. **Choose the tag format** the project uses. Common conventions:
   - `vX.Y.Z` (most projects, including Go modules).
   - `X.Y.Z` (some npm packages).
   - `<package>-X.Y.Z` (monorepos with multiple publishables).
   Use the existing convention; check `git tag` for prior tags.
2. **Annotate the tag** with the release notes (or a one-line
   reference to `CHANGELOG.md`):
   ```
   git tag -a vX.Y.Z -m "vX.Y.Z — <one-line summary>
   <body of release notes>"
   ```
   Lightweight tags (`git tag vX.Y.Z` without `-a`) are
   discouraged — they don't carry metadata.
3. **Sign the tag** if the project requires it: `git tag -s` or
   ensure your gpg/ssh key is configured. Some package registries
   (npm provenance, Go module checksums) verify signatures.
4. **Push the tag** explicitly: `git push origin vX.Y.Z`. A bare
   `git push` does NOT push tags.
5. **Verify the tag is visible** on the remote: refresh the
   GitHub/GitLab releases page, or `git ls-remote --tags origin`.
6. **Trigger downstream steps** — many projects auto-publish on
   tag push (npm, PyPI, container registry). If yours doesn't, do
   the publish step explicitly and confirm.

## What NOT to do

- Tag from a feature branch or with uncommitted changes.
- Move a published tag (`git tag -f`) — that breaks every
  consumer caching by tag SHA. If you tagged the wrong commit,
  cut a new patch version.
- Tag without an annotation message — historians will hate you.
- Push a tag and leave the changelog stale. Consumers will read
  it from the tag.
- Skip the dry-run. Most release pipelines have a "draft" or
  "dry run" mode; use it for the first release in a series.

## Output shape

```
Tagged: <tag>
Commit: <sha>
Pushed to: <remote>
Visible at: <release page url>
Auto-publish status: <triggered / manual step needed>
```
