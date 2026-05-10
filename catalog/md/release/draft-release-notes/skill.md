---
name: draft-release-notes
description: Write the release notes / changelog entry from the commit history since the last tag
phase: release
user_invocable: true
aliases: [changelog, release-notes]
origin: new
---

# /draft-release-notes — write the release notes

Turn `git log <last-tag>..HEAD` into a changelog entry users can
actually read.

## Steps

1. **Get the commit list** since the last release tag:
   `git log <last-tag>..HEAD --oneline --no-merges`. Read it
   end-to-end.
2. **Group commits by user-visible category**, not by type prefix:
   - **Highlights** — 1–3 things users will notice most.
   - **Features added** — what's new and useful.
   - **Bug fixes** — what stopped happening.
   - **Performance** — measurable improvements (cite numbers if
     you have them).
   - **Breaking changes** — explicit list, with migration
     guidance.
   - **Deprecations** — things slated for removal in a future
     version.
   - **Internal / chore** — usually omit unless it affects
     contributors.
3. **Write each entry as the user impact**, not the implementation:
   - Bad: "Refactor cache layer."
   - Good: "Cache hits are ~3× faster; cold-start unchanged."
   - Bad: "Add `--retry` flag."
   - Good: "`fetch` now retries transient failures up to 3 times
     by default; pass `--retry=0` to disable."
4. **Cite the commits or PRs** for traceability:
   `(#123, #145)` or `(<short-sha>)`.
5. **For breaking changes**, write the migration note: what the
   user sees, what they should change, when (if ever) the old
   behaviour stops working.
6. **Add to `CHANGELOG.md`** at the top under a new version
   header. Keep prior entries; release notes are append-mostly.
   Many projects use Keep-a-Changelog or Conventional Changelog
   formats — match the existing style.

## Tone

Write like you'd want a release note written for a tool you
depend on. Plain language, concrete examples, honest about
breaking changes. No marketing.

## What NOT to do

- Dump the raw commit list. Half of it is "fix typo" and "wip";
  filter ruthlessly.
- Hide breaking changes inside a "Misc" section. They get their
  own heading.
- Cite issues or PRs that are private — most consumers can't
  access them.
- Promise things that aren't actually shipping yet. The release
  notes describe *this* version; future plans go in a roadmap
  doc.
- Forget to mention the previous-version → this-version diff
  link if the project's convention includes one.

## Output shape

A markdown block ready to paste into `CHANGELOG.md`:

```
## [X.Y.Z] - YYYY-MM-DD

### Highlights
- <user-visible thing>

### Added
- <feature> (#PR)

### Changed
- <change> (#PR)

### Fixed
- <bug> (#PR)

### Breaking
- <change> — Migration: <one paragraph>

### Deprecated
- <thing> — will be removed in <version>
```
