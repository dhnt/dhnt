---
name: amend
description: Amend the most recent commit — message fix, file added, mistake corrected — only if the commit isn't published
phase: commit
user_invocable: true
aliases: [git-amend]
origin: new
---

# /amend — amend the previous commit

Amend the most recent commit when something small needs fixing
and the commit hasn't been published yet.

## When amending is appropriate

- Typo in the commit message.
- Forgot to stage one file.
- Pre-commit hook caught an issue and the fix is trivial.
- The commit is private (not yet pushed, not yet shared).

## When amending is NOT appropriate

- The commit is already pushed to a shared branch. Force-push
  rewrites history that other clones already have; downstream
  pulls will get confused. Cut a new commit instead.
- A pre-commit hook *failed* — the commit didn't actually happen,
  so `--amend` would target the *previous* commit (the wrong
  one). Stage the fix and create a new commit.
- The change is non-trivial. Amending hides the chronology of
  fixes. Bigger fixes deserve their own commit.
- The original commit was made by someone else. Don't rewrite
  others' commits; cut a new commit on top.

## Steps

1. **Confirm the commit isn't published.**
   `git log origin/<branch>..HEAD` — if it shows the commit, you
   own it and can amend.
2. **Stage the additional fix** (or none, if you're only fixing
   the message).
3. **Amend**:
   - Message only: `git commit --amend -m "new message"`.
   - Add files: `git commit --amend --no-edit` (keeps the
     original message).
   - Both: `git commit --amend` (opens editor).
4. **Verify**: `git log -1 --stat`.
5. **Push** if needed. If you've already pushed the previous
   version, this requires `--force-with-lease` (never `--force`).
   If the branch is shared, double-check no one else has pulled.

## Force-push safety

- **Never** `--force` on `main` / `master` / `develop` /
  protected branches.
- Use `--force-with-lease` instead of `--force` — it refuses to
  overwrite if someone else pushed in the meantime.
- Communicate before force-pushing a shared feature branch.

## What NOT to do

- Amend after a pre-commit hook failure. The hook *prevented* the
  commit; amending would target the previous one. Re-stage and
  commit anew.
- Amend a commit that's been pushed to a shared branch without
  understanding the consequences.
- Bundle multiple unrelated fixes into one amend. If the fix is
  non-trivial, make a new commit.
- Amend to add a `Co-Authored-By` trailer or any agent
  attribution unless explicitly requested.

## Output shape

```
Amended: <sha-before> → <sha-after>
Changes: <message-only | files added | both>
Pushed: <yes / no / required force-with-lease>
```
