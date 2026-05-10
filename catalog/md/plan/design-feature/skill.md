---
name: design-feature
description: Design a new feature before writing any code — API shape, data flow, edge cases, alternatives
phase: plan
user_invocable: true
aliases: [design, plan-feature]
origin: new
---

# /design-feature — design before coding

A new capability has been requested. Before opening any source
file to edit, do the design work.

## Steps

1. **Restate the requirement** in your own words. Surface gaps:
   - Who's the user? (end-user / API caller / agent / operator)
   - What outcome do they want? (in plain prose, no jargon)
   - What's explicitly out of scope?
   - What's the success criterion? (test, behaviour, metric)
2. **Identify affected components.** Walk the architecture brief
   (or generate one with `/summarize-architecture` if missing).
   Mark every component the change touches.
3. **Sketch the API shape.** For each affected component:
   - Public interface that changes (function signature, endpoint,
     CLI flag, struct field).
   - Inputs: types, validation, defaults.
   - Outputs: types, error cases.
   - Backwards compatibility: do existing callers break?
4. **Map the data flow.** Where does the new state come from, where
   does it go, what's authoritative? Update any persistence layer
   schemas you'll need.
5. **Walk through edge cases.** For each input dimension:
   - Empty / null / zero.
   - Maximum size / pagination boundaries.
   - Malformed / hostile input.
   - Concurrent calls / race conditions.
   - Failure of dependencies (DB down, network timeout).
   For each, write the expected behaviour in one sentence.
6. **Consider alternatives.** Even briefly. The "obvious" design is
   sometimes obviously wrong, and at least naming one alternative
   forces the question.
7. **Identify the smallest viable shape.** What's the minimum that
   delivers the user outcome? Iterate up from there if needed —
   don't iterate down from a bigger design.
8. **List risks and mitigations.** Performance hot paths, security
   surface area, breaking changes for downstream callers,
   observability holes.

## Output shape

```
## Feature: <name>

## Outcome
<one paragraph: what the user can now do>

## Out of scope
- <thing that might be expected but isn't here>

## API
<signature / endpoint / CLI flag / struct definition>

## Data flow
<how state moves; new tables/columns/keys if any>

## Edge cases
| Input shape | Expected behaviour |
|---|---|
| <…> | <…> |

## Alternatives considered
1. <alternative A> — <why not>
2. <alternative B> — <why not>

## Risks
- <risk> — <mitigation or "accepted">

## Affected components
- <path> — <what changes>
- <path> — <what changes>

## Testing strategy
- <test type>: <what it asserts>
```

## When to push back instead of designing

- The requirement is fundamentally ambiguous and the user can't
  resolve it in two clarifying questions. Write a one-pager
  explaining what's unclear; don't design around guesses.
- The feature breaks a documented invariant. Surface the conflict
  explicitly before designing.
- The smallest viable shape exposes a deeper issue (e.g. needs a
  schema migration that affects every consumer). Call that out as
  scope creep before committing.

## What NOT to do

- Design for hypothetical future requirements. Three concrete
  consumers is the bar for an abstraction; one isn't.
- Bury the API in implementation detail. The reader should be able
  to predict what callers write before they read the
  implementation.
- Skip alternatives because "the answer is obvious." The exercise
  catches assumptions you didn't know you had.
- Skip edge cases. Production bugs live there.
