---
name: audit-dependencies
description: Audit dependencies for security advisories, license compatibility, abandoned packages, and supply-chain risk
phase: maintain
user_invocable: true
aliases: [dep-audit, supply-chain-audit]
origin: new
---

# /audit-dependencies — dependency audit

Take stock of the project's dependency health. Surface security
issues, licensing concerns, abandoned packages, and supply-chain
risks.

## Audit dimensions

### Security

- Run the ecosystem's vulnerability scanner:
  - `npm audit`, `pnpm audit`
  - `govulncheck ./...`
  - `cargo audit`
  - `pip-audit`, `safety check`
- For each finding: severity, affected version range, whether
  the project actually exercises the vulnerable code path, fixed
  version available?

### License compatibility

- Run a license scanner:
  - `npx license-checker`
  - `go-licenses report ./...`
  - `cargo deny check licenses`
  - `pip-licenses`
- Flag any non-permissive licenses (GPL, AGPL, SSPL, CPAL, BSL)
  if the project's policy excludes them. Match the project's
  documented license policy — many ship a list of allowed
  licenses.

### Maintenance / abandonment

- Check each dependency's last release date and activity:
  - More than 18 months without a release on a non-stable
    package is a yellow flag.
  - Archived / deprecated upstreams are red flags; plan
    migration.
  - Single-maintainer + low activity = supply-chain risk.

### Supply-chain hygiene

- **Pin lockfile is committed** (`package-lock.json`, `Cargo.lock`,
  `go.sum`, `poetry.lock`).
- **Subresource integrity** if the project loads scripts from a
  CDN.
- **Provenance signatures** (npm provenance, sigstore) verified
  in CI.
- **Dependency freshness** — too-new packages are unverified;
  too-old packages may have unfixed CVEs. Both are risks.

## Process

1. Run all the scanners above. Capture raw output.
2. Triage findings by severity / impact / actionability.
3. Open issues / tickets for items that need action; note items
   that are "known and accepted."
4. Recommend a remediation plan with priority order (critical
   security → license blockers → abandonment migrations →
   freshness).

## What NOT to do

- Treat every CVE as an emergency. Most don't affect the project
  in the way the advisory describes.
- Auto-update everything `audit fix` suggests. Some fixes
  introduce breaking changes; review before applying.
- Ignore findings that "look like noise." False positives are
  real but should be marked as suppressed-with-reason, not just
  ignored.
- Skip licenses because "we'll deal with it later." License
  surprises kill product launches.
- Audit once and call it done. Run on a cadence (monthly
  minimum, or in CI).

## Output shape

```
## Security
- <CVE-id> · <package>@<version> · severity: <…> · path: <reachable / unreachable> · fix: <version available>
- … (N total)

## License
- <package>@<version> · license: <SPDX> · status: <allowed / disallowed / review>

## Maintenance
- <package> · last-release: <date> · status: <active / stale / archived>

## Supply chain hygiene
- Lockfile: <committed / missing>
- Provenance: <verified / N packages without>
- Critical findings: <…>

## Recommendation (priority order)
1. <action>
2. <action>
```
