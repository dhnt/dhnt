---
name: diagnose-error
description: Diagnose an error reported in production — find the root cause from logs, traces, and code
phase: operate
user_invocable: true
aliases: [debug-prod, root-cause]
origin: new
---

# /diagnose-error — diagnose a production error

A production error was reported. Walk it back to root cause from
the limited evidence you have (logs, traces, metrics, the user's
report).

## Steps

1. **Capture the symptom precisely.**
   - Error message verbatim.
   - User's repro steps (if any).
   - When did it start? Did it correlate with a deploy?
   - Frequency — single occurrence, intermittent, sustained?
   - Blast radius — one user, one tenant, all users?
2. **Find the originating log line.** Use the project's log
   query tool (Grafana, Datadog, CloudWatch, OTEL collector, …)
   to locate the first occurrence. Capture the full stack trace
   if there is one.
3. **Trace the request** from entry point to error site. The
   project's distributed-tracing tool (if instrumented) shows the
   hop-by-hop path; otherwise correlate by request ID across
   logs.
4. **Read the code at the failing line.** Understand:
   - What invariant is violated?
   - What input shape would cause this?
   - Is it a precondition we're failing to validate, a downstream
     dependency that's misbehaving, or a logic bug?
5. **Correlate with recent changes.** `git log -5 -- <file>` —
   was this code recently modified? What about its dependencies?
6. **Form a hypothesis** and test it:
   - If "X always reproduces it" — write a test, run it.
   - If "X happens occasionally" — it's likely a race or a
     stateful preconditional.
   - If "X happens after Y minutes of uptime" — leak / state
     accumulation / time-based logic.
7. **Capture the diagnosis** — even before fixing — so the next
   on-call has the trail.

## Output shape

```
## Symptom
- Error: <verbatim>
- Frequency: <one-time | intermittent | sustained>
- Affected: <users / tenants / surfaces>
- First seen: <timestamp>
- Correlates with: <deploy / event / nothing yet>

## Trace
- Entry point: <handler / endpoint>
- Failure site: <file:line>
- Stack: <key frames>

## Root cause
<one paragraph: the actual underlying issue>

## Reproduction
- <steps that trigger it>

## Recommendation
- Immediate: <hotfix / rollback / mitigation>
- Permanent: <what to fix and where>
- Prevention: <test or assertion to add>
```

## What NOT to do

- Restart the service and call it fixed. Restarts mask
  state-accumulation bugs and rob you of evidence.
- Speculate without checking the logs. "Probably X" without
  evidence is how on-call rotations stay miserable.
- Stop at the first plausible cause. Confirm by reproducing or
  by reading the code path end-to-end.
- Push a hot patch without a regression test. The next deploy
  will reintroduce the bug.
- Skip the root-cause capture because "we just need to ship the
  fix." The capture is what prevents the *next* iteration of
  this bug.
