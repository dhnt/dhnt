---
name: deploy
description: Deploy the current build to a target environment using the project's documented mechanism
phase: deploy
user_invocable: true
aliases: [ship, push-to-env]
origin: new
---

# /deploy — deploy to an environment

Push a build to dev / staging / prod using whatever mechanism the
project documents. Don't invent one.

## Pre-flight

1. **Identify the target environment** the user wants. Default
   to `dev` if unstated; never default to `prod`.
2. **Confirm the deploy mechanism**: `make deploy`, a CI/CD
   pipeline trigger, `kubectl apply`, `terraform apply`, a
   project-specific script. Read the project's deploy docs first.
3. **Confirm the artifact** to deploy — usually the current
   commit, possibly a specific tag. Verify it exists and CI is
   green for it.
4. **For staging / prod**: confirm the user has authorisation. If
   the deploy is irreversible (DB migration, schema change, paid
   API call), explicitly ask before proceeding.

## Steps

1. **Run a dry-run if available** (`terraform plan`,
   `kubectl diff`, `helm install --dry-run`). Read the output;
   surface anything unexpected.
2. **Apply** the deploy. Stream output so the user sees progress.
3. **Wait for the deploy to land** — health check, readiness
   probe, smoke test. Do not declare success until the target
   environment is actually serving the new version.
4. **Verify post-deploy**: hit the health endpoint, run the
   smoke-test command, check for new error rate spikes in the
   monitoring dashboard.
5. **Update the deploy log / changelog / Slack channel** per the
   project's convention.

## Deploy targets a coding agent commonly meets

- **Local / dev**: usually a `make dev` / `npm run dev` /
  `cargo run` for live development.
- **Staging**: typically requires a CI artifact + an apply step.
  Often gated behind a manual approval.
- **Production**: usually only via CI, never directly. If your
  deploy command pushes straight to prod from a dev machine, push
  back and ask whether that's really the right path.

## What NOT to do

- Deploy from a dirty working tree.
- Skip the dry-run on staging or prod.
- Deploy from an unmerged feature branch unless the project's
  workflow expects it.
- Continue past a failed health check and "let it warm up." Roll
  back, diagnose, redeploy.
- Hide errors from the user during streaming. Surface them
  clearly.
- Run `terraform apply` / `kubectl apply` with `--auto-approve`
  on production state without explicit user instruction.

## Output shape

```
Deployed: <artifact / version>
Target: <env>
Mechanism: <command run>
Health check: <ok / failed>
Verification: <smoke test result>
Log entry: <where the deploy was recorded>
```
