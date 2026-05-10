---
name: setup-dev-env
description: Set up a working development environment for an unfamiliar repository — install, verify, ready to contribute
phase: onboard
user_invocable: true
aliases: [setup, dev-setup]
origin: new
---

# /setup-dev-env — set up the dev environment

You've cloned the repo. Make it build, run tests locally, and
have a sensible editor / linter setup. Goal: from `git clone` to
"I can make a one-line change and verify it" in under 30 minutes.

## Steps

1. **Read the project's setup docs first**: `README.md`
   (Installation / Setup / Getting Started section), `AGENTS.md`,
   `CONTRIBUTING.md`. Most projects have a documented happy path.
   Use it. Don't invent.
2. **Check prerequisites**: language toolchain version
   (`go.mod` `go 1.X`, `package.json` `engines.node`,
   `pyproject.toml` `requires-python`, etc.). Install or
   activate (`nvm use`, `rustup default`, `pyenv shell`).
3. **Install dependencies** with the project's preferred tool:
   - Node: `npm install` / `pnpm install` / `yarn install`
     (look at the lockfile to know which).
   - Go: `go mod download`.
   - Rust: handled by `cargo build`.
   - Python: `pip install -e .` / `poetry install` / `uv sync`.
4. **Run the build** to surface env issues early. The first
   build often fails on missing system deps (gcc, openssl, llvm,
   protoc); read the error, install, retry.
5. **Run the tests**. A green test suite is the strongest
   "environment is set up correctly" signal. Note long /
   integration tests; default to the unit tier first.
6. **Set up the editor / IDE**:
   - Language server for the project's languages.
   - Formatter on save (prettier, gofmt, rustfmt, black, ruff).
   - Linter integration (eslint, golangci-lint, clippy, mypy).
   - Whatever else `.editorconfig` / `.vscode/settings.json` /
     project docs specify.
7. **Verify with a no-op change** — change a comment, rebuild,
   re-run tests. Confirms the loop works end-to-end.
8. **Note open env issues** the user will hit on day 2: docker
   needed for integration tests, a vendor account needed for
   some service, a `direnv` setup for env vars.

## What NOT to do

- Skip the project's setup docs and improvise. The docs encode
  things you can't easily infer.
- Install global tools when local-to-the-project versions are
  expected (`npx`, `mise`, `asdf`, `nix-shell`).
- Modify project config to "fix" your local env. The right move
  is to fix your env, not commit a workaround.
- Declare done at "build passes" without running the tests.
  Build passing means it compiles, not that it works.
- Skip the no-op verification. The setup that works for setup
  isn't always the setup that works for development.

## Output shape

```
## Toolchain
- <language>: <version detected | installed>

## Dependencies
- Installed via: <command>
- Time: <duration>

## Build
- Command: <…>
- Result: <ok>

## Tests (unit)
- Command: <…>
- Result: <N pass / M fail>

## Editor
- LSP: <set up>
- Formatter: <set up>
- Linter: <set up>

## Open issues
- <thing the user will hit later>
```
