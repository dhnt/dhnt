---
name: update-pr-description
description: Refresh the PR title and body to reflect the actual current state of the branch
phase: integrate
user_invocable: true
aliases: []
origin: priorart/openhands
---

# /update-pr-description — refresh the PR description

The branch evolved during review. The PR description on the web
might no longer match what's actually in the diff. Update it.

## When this is needed

- Significant code was added or removed during review.
- The scope changed (something planned was deferred, or something
  unplanned was bundled in).
- The test plan needs new items.
- The original description was thin and a substantive review
  surfaced what's really going on.
- A breaking change was introduced after the PR opened that
  wasn't in the original description.

## Steps

1. **Read the current PR description** on the platform.
2. **Read the full diff** against the base branch (not just the
   latest commits).
3. **Identify deltas** between the description and reality:
   - Bullet points that no longer match the code.
   - New behaviour the description doesn't mention.
   - Test-plan items that were added or skipped.
   - Scope that grew or shrunk.
4. **Rewrite, don't append.** A description with strikethrough
   sections is hard to read; a clean, current one isn't. Keep
   the structure (Summary / Test plan / etc.) but rewrite the
   contents.
5. **Update via the platform's API** (`gh pr edit <n> --body
   "$(cat <<'EOF'…)"`). HEREDOC the body to avoid quoting bugs.
6. **Reply on the PR** with a one-line note: "Description
   updated to reflect <what changed>." Reviewers may want to
   re-read.

## What NOT to do

- Hide changes by sneaking them into a description revision
  without flagging. The reviewers' first read is anchored to the
  original description; surprise them now and trust erodes.
- Strip sections (test plan, migration notes) because "they're
  obvious now." The reviewer might not have re-read the diff.
- Change the title casually — it shows up in commit logs once
  merged. Update it only when the change of scope warrants.

## Output shape

```
Updated <PR url>
Changes:
- <what shifted in the description>
- <what shifted in the test plan>
- <what shifted in scope>
```
