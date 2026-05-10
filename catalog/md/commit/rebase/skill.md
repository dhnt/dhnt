---
name: rebase
description: Rebase the current branch onto an updated base — clean linear history, conflicts resolved
phase: commit
user_invocable: true
aliases: [git-rebase]
origin: new
---

# /rebase — rebase onto the updated base

Rebase the current branch onto its base (typically `main`) to
keep history linear and incorporate upstream changes.

## When to rebase

- The base branch advanced while your work was in flight; you
  want a linear history and to test your changes against the
  current base.
- Before opening a PR — many projects prefer a clean linear
  history.
- Squashing a sequence of WIP commits into a clean set before
  review.

## When NOT to rebase

- The branch is shared with other contributors. Rebasing
  rewrites history; their clones will be confused.
- The branch is already in review and reviewers have started
  comments — they may lose context. Merge instead, or coordinate
  the rebase explicitly.
- You don't have a clean working tree. Stash or commit first.

## Steps

1. **Update the base** locally:
   ```
   git fetch origin
   ```
2. **Verify your branch is unshared**:
   ```
   git branch -r --contains HEAD
   ```
   If it's only `origin/<your-branch>`, you're safe.
3. **Rebase**:
   ```
   git rebase origin/<base-branch>
   ```
4. **Resolve conflicts as they arise**:
   - Read both sides; understand the intent.
   - Keep what your commit needs; accept the base's intent
     elsewhere.
   - `git add <file>` to mark resolved; `git rebase --continue`.
   - Don't blindly take "theirs" or "ours"; that's how silent
     bugs get reintroduced.
5. **Run the build + tests** after the rebase. Conflict
   resolutions can compile but be wrong.
6. **Force-push** with safety:
   ```
   git push --force-with-lease origin <your-branch>
   ```
   Never bare `--force`. Never push to protected branches.

## Interactive rebase (squashing / reordering)

`git rebase -i origin/<base>` opens the rebase plan. Common
operations:

- `pick` — keep the commit as-is.
- `squash` — combine into the previous commit, edit message.
- `fixup` — combine into the previous, discard message.
- `reword` — keep the commit, edit just the message.
- `drop` — remove the commit entirely.

Use squash/fixup to clean up WIP commits before review. Don't
overuse — small atomic commits are useful for `git bisect` later.

## What NOT to do

- Rebase a published / shared branch without coordination.
- Resolve a conflict by deleting one side wholesale ("their
  whole change") without reading what the other side intended.
- Skip the build/test after rebase.
- Use `--force` (no `-with-lease`). It overwrites silently if
  someone else has pushed.
- Rebase across a merge commit — usually wrong; use
  `--rebase-merges` if you really need it.

## Output shape

```
Rebased: <branch> onto <base>
Commits replayed: <N>
Conflicts: <resolved | none>
Build/tests: <ok>
Pushed: <force-with-lease ok | not yet>
```
