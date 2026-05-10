---
name: explain-changes
description: Walk through recent changes (a commit, a diff, a session) and explain what they do and why
phase: build
user_invocable: true
aliases: [explain, walk-diff]
origin: priorart/cline
---

# /explain-changes — explain a diff or commit history

The user wants to understand a chunk of recent work. They might be
reviewing your own session output, or onboarding to a teammate's
PR, or trying to remember what they did last week.

## Steps

1. **Establish the scope.** Default to the diff between the
   current branch and the base branch. If the user named a
   commit, range, or session, target that.
2. **Read the diff once end-to-end** before forming a narrative.
   Note the magnitude (files, ±lines, commits if a range).
3. **Group changes by intent**, not by file. Common groupings:
   - "Fixes": each grouped with the symptom it addresses.
   - "Refactors": each grouped with the smell it removes.
   - "New features": each grouped with the user-visible change.
   - "Test additions": noted alongside the code they cover.
   - "Docs / chore": brief mention, not a section.
4. **For each group, explain in plain prose**: what the change
   does, why it's needed, and any non-obvious mechanics.
5. **Surface anything that surprised you** — code that's
   suspiciously commented out, a test that was disabled, a
   dependency added without obvious motivation, a public API
   that broke.

## Output shape

```
## Overview
<2-3 sentences: total magnitude, overall direction>

## What changed
### Fixes (N)
- <one or two-line description, with file pointer>

### Refactors (N)
- <description, with file pointer>

### Features (N)
- <description, with file pointer>

### Tests (N)
- <description>

## Worth a closer look
- <surprising or risky thing — file:line>
```

## What NOT to do

- Restate the diff verbatim. Summarise.
- Editorialise ("this is great work" / "this is sloppy"). Stick
  to facts.
- Skip the "Worth a closer look" section if there's something
  worth flagging — the user is relying on you to notice.
- Pretend to understand a change you can't follow. Say "this
  block is unclear; would benefit from author explanation."
