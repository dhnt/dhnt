---
name: rollback
description: Roll a deployment back to the previous good version when a release is misbehaving
phase: deploy
user_invocable: true
aliases: [revert-deploy, undeploy]
origin: new
---

# /rollback — roll back a bad deploy

The current deploy is misbehaving and the right move is to
revert to the previous good version. Don't try to fix forward
under pressure — restore service first, debug second.

## Pre-flight

1. **Confirm the symptom** is a release regression, not an
   infrastructure incident or a transient downstream failure.
   Rolling back fixes the first; rollback won't help the others.
2. **Identify the previous good version** — the deploy log /
   release page / `kubectl rollout history` / `helm history`
   should show it.
3. **Identify the rollback mechanism** the project uses:
   - `kubectl rollout undo deployment/<name>`
   - `helm rollback <release> <revision>`
   - re-deploy the previous tag
   - revert the database migration (if applicable — usually NOT
     part of rollback; see "data migrations" below).
4. **Confirm authorisation** — rollback to prod should be an
   explicit ask, even if the symptom is severe.

## Steps

1. **Communicate first.** Post in the deploy / incident channel
   that you're rolling back. The fastest way to make a bad
   incident worse is two engineers running conflicting commands.
2. **Execute the rollback.** Use the documented mechanism, not a
   custom recovery procedure invented under pressure.
3. **Wait for the rollback to land** — same health checks and
   readiness probes as deploy.
4. **Verify recovery**: error rate drops back to baseline, the
   symptom that triggered rollback is gone, no new symptoms
   appeared.
5. **Lock down the bad version** so no one redeploys it by
   accident: yank the tag from the registry if the project
   supports it, or mark it bad in your deploy tooling.
6. **Capture the timeline** for the post-mortem: when did the
   bad deploy go out, when was the symptom first observed, when
   was rollback initiated, when was service restored.

## Data migrations are special

Schema changes are usually one-way. If the bad deploy ran a
migration:

- A reverse migration may exist (`down`, `--reverse`). Test it
  in staging before applying to prod.
- If no reverse exists, you cannot rollback the schema. The
  application code can roll back, but the schema is what it is —
  fix-forward at the code layer to be compatible with the new
  schema.
- Discuss before reverting database state. The data is the part
  you can't redeploy.

## What NOT to do

- Roll forward under pressure. The bad deploy proves you got
  something wrong; another quick deploy will probably get
  something else wrong.
- Skip the communication step. Other engineers may be diagnosing
  the same symptom.
- Roll back a database migration without a tested down-migration
  path.
- Declare success based on "the rollback command exited 0." Wait
  for service to actually recover.
- Forget the post-mortem. The point of rollback is to buy time;
  the next step is understanding what went wrong.

## Output shape

```
Rolled back: <bad version> → <good version>
Target: <env>
Mechanism: <command>
Recovery time: <duration>
Health check: <ok>
Bad version locked: <yes / no — and how>
Post-mortem ticket: <link>
```
