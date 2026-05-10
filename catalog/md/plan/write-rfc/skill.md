---
name: write-rfc
description: Author a request-for-comments document for a substantial proposal that needs review before implementation
phase: plan
user_invocable: true
aliases: [rfc, design-doc]
origin: new
---

# /write-rfc — author an RFC

An RFC (request for comments) is a written proposal for a
substantial change, circulated for review *before* implementation.
Use this skill when the change is big enough that getting buy-in on
the approach is worth more than racing to a working diff.

## When to write an RFC

- The change affects a public API, a data model, an external
  contract, or a multi-team interface.
- The implementation cost is more than a person-week.
- Reasonable engineers disagree about the right approach.
- The change is hard to reverse once shipped.
- A formal record of the decision is needed (audit, compliance,
  team onboarding).

If none of those apply, an RFC is overkill — write a smaller
design note via `/design-feature` instead.

## Sections an RFC must have

1. **Title** — short, descriptive. Not a clever phrase.
2. **Status** — Draft / In review / Accepted / Rejected /
   Superseded.
3. **Authors and reviewers** — who wrote it, who needs to weigh
   in.
4. **Problem statement** — what's broken or missing today, in
   prose. No jargon. Include the cost of *not* doing anything
   ("no-op cost").
5. **Goals and non-goals** — what success looks like; what's
   explicitly out of scope.
6. **Proposed solution** — the recommended approach, in enough
   detail that a reviewer can spot weaknesses. Include API shape,
   data flow, integration points.
7. **Alternatives considered** — at least two, each with a real
   reason for rejection. "We considered X but it's worse" without
   evidence is not a real alternative.
8. **Migration / rollout** — for changes that affect existing
   users or data: how do we get from here to there without
   breaking things?
9. **Security & privacy considerations** — even if just "no
   change to the threat model."
10. **Backwards compatibility** — what breaks, who's affected,
    how long is the deprecation window.
11. **Observability** — how will we know it's working in
    production (metrics, logs, alerts).
12. **Open questions** — explicit list of unresolved decisions
    that need reviewer input.
13. **Decision criteria** — what would change your recommendation
    or the choice between alternatives.

## Process

1. **Read existing RFCs in the project** before drafting. Match
   tone, depth, and structure conventions.
2. **Write the problem statement first** and circulate it
   informally. If reviewers don't agree there's a problem, the
   solution doesn't matter yet.
3. **Draft the solution and at least two alternatives** in
   parallel. The exercise of explaining alternatives often
   improves the recommended one.
4. **Self-review for the open questions section** — be honest
   about what you don't know. RFCs that pretend to have all the
   answers don't get useful comments.
5. **Send for review** with a deadline ("comments by Friday") and
   a tagged set of reviewers. Async review without deadlines
   stalls.
6. **Resolve every comment** before marking Accepted, even if the
   resolution is "noted, won't change." Silent dismissal makes
   reviewers stop reading future RFCs.

## Tone

Aim for "thoughtful colleague proposing something" — not "marketing
brochure" and not "academic paper." Be concrete, cite the code,
quantify when you can ("this query is N+1 at ~500 RPS"), and admit
uncertainty.

## What NOT to do

- Skip alternatives because the recommendation is "obvious." Write
  them anyway; the obvious one isn't always best.
- Write the RFC after starting the implementation. The RFC's job
  is to *prevent* expensive mistakes; doing it post-hoc just
  documents one.
- Overspecify the implementation. RFCs commit to the *approach*,
  not every line of code. Save details for the PR.
- Use an RFC as a writing exercise. If two paragraphs would do, do
  two paragraphs. RFCs that match the project's conventions get
  read; bespoke 30-page documents get skimmed.
- Pretend there are no open questions. Every interesting RFC has
  some.

## Length

A good RFC is as short as possible while complete. For most
projects, 1–4 pages of prose. The proposed solution and the
alternatives are usually the longest sections.
