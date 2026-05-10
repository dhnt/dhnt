---
name: triage-incident
description: Triage a live production incident — assess severity, contain the blast, hand off cleanly
phase: operate
user_invocable: true
aliases: [incident, oncall]
origin: new
---

# /triage-incident — triage a production incident

Something is on fire. Your job in the first 30 minutes is to
contain blast radius, not to fix the bug. Save fixing for after.

## Steps

1. **Acknowledge.** If there's an alert, ack it. If there's a
   user report, reply that you're looking. Silence makes
   incidents feel worse than they are.
2. **Classify severity.** Common bands:
   - **Sev1**: customer-facing outage, data loss, security
     breach. Page leadership.
   - **Sev2**: substantial degradation, partial outage. Drop
     other work.
   - **Sev3**: minor impact, single feature broken. Address in
     business hours.
   - **Sev4**: cosmetic. File a ticket; no urgent action.
   Match the project's documented severity scale.
3. **Open the incident channel / ticket** per the project's
   convention. Pin the timeline; everything you learn goes here.
4. **Assess blast radius.** How many users / tenants / regions
   are affected? Is it growing or stable?
5. **Mitigate before fixing.** Mitigations preserve service
   while you investigate:
   - Roll back the recent deploy (`/rollback`).
   - Scale up the failing service if it's overloaded.
   - Drain traffic to a healthy region.
   - Disable the broken feature with a flag.
   - Rate-limit the harmful traffic.
   The right mitigation depends on the symptom; pick the
   smallest one that restores service.
6. **Communicate** every 15–30 minutes during an active sev1/2:
   what's known, what's unclear, what you're trying.
7. **Hand off cleanly** if you can't continue: timeline +
   current state + what you've ruled out + suspicious leads.
8. **Schedule the post-mortem** before the incident ends. The
   meeting itself happens later, but the calendar slot prevents
   "we'll get to it" from becoming "we never got to it."

## What NOT to do

- Investigate root cause before mitigating. Mitigation first;
  diagnosis second.
- Push a hotfix during the incident without a tested rollback
  path. Untested hotfixes cause sev2-becomes-sev1.
- Restart things without recording state. You may need that
  state for the post-mortem.
- Communicate only good news. The honest update is the right
  update — overconfident "we've got it" messages followed by
  silence erode trust.
- Skip the post-mortem. Incidents you don't learn from come
  back as worse incidents.
- Page everyone for a sev3. Severity drift is real; it makes
  on-call worse.

## Output shape (running incident timeline)

```
## Incident <id>: <one-line>
- Sev: <1-4>
- Started: <timestamp>
- First responder: <name>

## Timeline
- HH:MM <ack>
- HH:MM <observation>
- HH:MM <mitigation attempted: X>
- HH:MM <mitigation result>
- HH:MM <handed off to / resolved>

## Current state
<what's true right now>

## Known
<facts established>

## Unclear
<questions still open>

## Next
<what's being tried>
```
