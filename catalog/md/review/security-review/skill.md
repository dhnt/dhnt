---
name: security-review
description: Audit pending changes specifically for security regressions — auth, input handling, secrets, supply chain, cryptography
phase: review
user_invocable: true
aliases: [secreview, sec-review]
origin: new
---

# /security-review — security audit of pending changes

Examine the diff against the base branch with a security lens. A
*code-review* asks "is this correct?"; a *security-review* asks
"can this be abused?".

## Threat model the review addresses

- **Untrusted input** crossing trust boundaries (HTTP, RPC, CLI
  args, env vars, file uploads, message-bus messages).
- **Injection**: command, SQL, path-traversal, template, header,
  log, prompt injection (for LLM-adjacent code).
- **AuthN / AuthZ regressions**: weakened checks, missing checks,
  bypassable checks, principal confusion.
- **Secrets**: new secret material in code, logs, env files,
  fixtures, CI configs, error messages.
- **Crypto misuse**: rolling-your-own, weak primitives, missing
  authentication on encryption, hardcoded keys/IVs, CSPRNG vs
  PRNG.
- **Supply chain**: new dependencies, version pins removed, build
  pipeline changes, post-install hooks, lockfile integrity.
- **Resource exhaustion**: unbounded allocation from input,
  zip-bomb / decompression bomb, regex DOS, recursion without
  depth limit.
- **Insecure defaults**: TLS verification off, CORS wildcards,
  cookie flags missing, SSRF-prone HTTP clients.
- **Deserialization** of untrusted data into typed structures.

## Process

1. Read the diff. Identify every file that touches a trust
   boundary, an auth path, a secret-handling path, or a parser.
2. For each suspicious construct, read the surrounding code to
   confirm or rule out the concern.
3. Cross-reference with the language/framework's known
   anti-patterns (e.g. Go's `http.ServeMux` host header,
   Python's `pickle`, JS's `eval`, Rust's `unsafe`).
4. Check `go.mod` / `package.json` / `Cargo.toml` / equivalents
   for new dependencies — read their advisories if you have web
   access.

## Output shape

```
## Summary
<verdict in 1-2 sentences: clean / concerns / blockers>

## Vulnerabilities (N)
- <file:line> — <category> — <attack scenario> — <fix direction>

## Hardening opportunities (N)
- <file:line> — <improvement> — <why>

## Out-of-scope observations (≤3)
- <one-line — non-security but caught while reviewing>
```

A vulnerability is something an attacker could exercise. A
hardening opportunity is defence-in-depth. Don't conflate them.

## What NOT to do

- Speculate about CVEs without evidence.
- Block a PR for theoretical attacks that the threat model
  doesn't include (e.g. local-machine attacks against tooling
  that only runs on dev machines).
- Recommend "use a security library" without naming one.
- Suggest cryptography work without flagging that it should be
  reviewed by a cryptographer.
