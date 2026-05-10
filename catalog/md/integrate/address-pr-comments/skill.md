---
name: address-pr-comments
description: Read PR review feedback and address each comment — fix, push back, or defer
phase: integrate
user_invocable: true
aliases: [pr-comments, respond-review]
origin: priorart/openhands
---

# /address-pr-comments — handle PR review feedback

A reviewer left comments on your PR. Each one needs a response —
either you change the code, or you explain why you won't.

## Steps

1. **Pull every comment** from the PR (use the project's CLI:
   `gh api repos/<o>/<r>/pulls/<n>/comments`, or equivalent).
   Don't rely on email notifications — they go stale fast.
2. **Categorise each comment** before you start fixing:
   - **Bug / correctness** → fix; reply with the commit that
     addresses it.
   - **Suggestion** → adopt if the reviewer is right; push back
     with reasoning if you disagree.
   - **Question** → answer in-line; if the answer reveals the
     code is unclear, also add a clarifying comment in the code.
   - **Style / convention** → adopt unless it conflicts with the
     project's documented style.
   - **Out of scope** → acknowledge, file a follow-up issue,
     reply with the link.
3. **Make the changes** in the smallest commits that address
   each comment. Don't bundle unrelated comments into one
   sweeping commit.
4. **Push** to the PR branch (regular push; never force unless
   the project's convention is force-push-on-rebase and the user
   asked).
5. **Reply to each comment thread** with the action taken:
   - "Fixed in <commit-sha>." — for code changes.
   - "Disagree because <reason>; happy to discuss." — for pushback.
   - "Filed as <issue-link> for separate work." — for deferrals.
6. **Re-request review** when all threads are addressed.

## Tone

Reviewers gave their time; respect it. Don't be defensive about
issues, don't dismiss suggestions without engaging, and don't
flood with thanks-emoji on every comment. Substantive replies are
the form of respect.

## What NOT to do

- Mark comments resolved without replying. The reviewer needs to
  know what changed.
- Make unrelated changes "while I'm in there." Stay focused on
  the comments.
- Push back on stylistic feedback that matches the project's
  convention. The reviewer is enforcing the rules everyone agreed
  to.
- Delete comment threads. They're part of the historical record
  of why the code is the way it is.
- Bundle "fix typo" with "rewrite the algorithm" in one commit.

## Output shape

```
Comments addressed: <N>
- Fixed: <N> (<commit-shas>)
- Pushed back: <N>
- Deferred: <N> (<issue-links>)
- Answered (no code change): <N>

Re-requested review: <reviewer>
```
