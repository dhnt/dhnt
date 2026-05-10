---
name: open-pr
description: Open a pull request for the current branch with a clear title, summary, and test plan
phase: integrate
user_invocable: true
aliases: [pr, pull-request]
origin: new
---

# /open-pr — open a pull request

Open a PR / MR for the current branch. The author of a good PR
description respects the reviewer's time: the reviewer should
understand *why*, not just *what*, in 30 seconds.

## Steps

1. **Confirm the branch is ready.** Working tree clean, all
   intended commits present, branch pushed to the remote, no
   debug code or commented-out blocks left.
2. **Pick the base branch.** Default to `main` / `master` /
   `develop` per the project's convention.
3. **Read the diff against base** end-to-end. The PR description
   summarises this — you can't summarise what you haven't read.
4. **Title** — short, imperative, no trailing period. ≤72 chars.
   Match the project's prefix convention (`feat:`, `fix:`, …) if
   commits do.
5. **Body** with these sections:
   - **Summary** — 1–3 bullets: what changed, why.
   - **Test plan** — bulleted checklist of how this was verified.
   - Optional: **Screenshots** for UI changes; **Migration notes**
     for breaking changes; **Closes #N** to link issues.
6. **Open via the project's preferred mechanism**: `gh pr create`,
   `glab mr create`, the project's CLI wrapper, or the web UI if
   automation isn't available. Use a HEREDOC for the body to avoid
   shell-quote issues.
7. **Verify** the PR opened, the base branch is correct, the
   reviewer requests are appropriate, and CI started.

## Output shape

The PR URL plus a brief recap:

```
PR opened: <url>
Base: <branch>
Reviewers: <…>
CI: <pending|running>
```

## What NOT to do

- Add a `🤖 Generated with` trailer or any agent-attribution
  unless the user explicitly asked for one.
- Force-push to `main` / a shared branch.
- Open the PR while CI is failing locally — fix first.
- Stuff unrelated commits into one PR. Split if the diff covers
  multiple intents.
- Skip the test plan section because "the diff is small."
- Restate the diff in the body. Summarise — the reviewer reads
  the diff.
