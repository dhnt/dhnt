---
name: draft-issue
description: Draft a clear GitHub / GitLab / Linear issue from a bug report or feature request
phase: integrate
user_invocable: true
aliases: [issue, bug-report]
origin: priorart/continue
---

# /draft-issue — draft a tracker issue

Turn a fuzzy report ("X is broken", "we should do Y") into a
ticket someone can pick up.

## Steps

1. **Identify the type**: bug report, feature request, technical
   debt, question. Each has a different shape.
2. **For a bug report** include:
   - **Steps to reproduce** — the smallest sequence that triggers
     the issue.
   - **Expected** vs **actual** behaviour.
   - **Environment** — version, OS, runtime, anything relevant.
   - **Logs / screenshots** — only the relevant slices, not the
     whole file.
   - **Severity** — blocking / major / minor / trivial.
3. **For a feature request** include:
   - **User story** — "As <role>, I want <capability> so I can
     <outcome>."
   - **Motivation** — what's painful or impossible today.
   - **Proposed shape** — only if you have one; otherwise leave
     open.
   - **Acceptance criteria** — what would make this done.
4. **For technical debt** include:
   - **Symptom** — what's slow / brittle / confusing.
   - **Cost of inaction** — what gets harder as time passes.
   - **Proposed cleanup** — outline.
5. **Title**: short, imperative for features ("Add X"), past-tense
   for bugs ("X breaks when Y"), descriptive for debt ("Y is
   tightly coupled to Z"). ≤80 chars.
6. **Tags / labels** — match the project's conventions
   (`bug`, `feat`, `area:auth`, etc.).
7. **Open** via the project's CLI (`gh issue create`,
   `glab issue create`, Linear API) with a HEREDOC body.

## What NOT to do

- File one issue for many problems. One ticket = one outcome.
- Skip the reproduction steps because "it's obvious." It rarely
  is to whoever picks the ticket up.
- Speculate about the root cause in the issue body. If you know
  the cause, fix it; if you don't, describe symptoms.
- Set severity higher than the truth to get attention. The
  triagers will catch it and trust suffers.
- Include private data (customer names, credentials, internal
  hostnames) in a public issue.

## Output shape

```
Issue opened: <url>
Type: bug | feat | debt | question
Title: <…>
Labels: <…>
```
