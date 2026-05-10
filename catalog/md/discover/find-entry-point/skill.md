---
name: find-entry-point
description: Locate where execution starts in an unfamiliar codebase, across language and project shapes
phase: discover
user_invocable: true
aliases: [entrypoint, where-it-starts]
origin: new
---

# /find-entry-point — locate where execution starts

Before reading code, find the door. Different ecosystems put it in
different places.

## Where to look

### Go
- `cmd/<name>/main.go` — the canonical place; `func main()` is the
  entry. Multiple subdirs of `cmd/` mean multiple binaries.
- A top-level `main.go` for single-binary projects.
- Library-only modules have no `main` — look at `doc.go` or the
  package with the largest public surface.

### Node / TypeScript
- `package.json` `"main"` field → CommonJS entry.
- `package.json` `"exports"` block → ESM entry per import path.
- `package.json` `"bin"` → CLI entry points.
- For frameworks: `next` → `pages/` or `app/`; `vite` → `index.html`
  + the script tag's src; `nest` → `src/main.ts`.
- Rollup / webpack: check `rollup.config.*` / `webpack.config.*`
  `input` / `entry` keys.

### Python
- `if __name__ == "__main__":` in any module — the script-style
  entry.
- `pyproject.toml` `[project.scripts]` → installed CLI entries.
- Web apps: look for `app.py`, `wsgi.py`, `asgi.py`, or the framework
  pattern (`from fastapi import FastAPI; app = FastAPI()`).
- `setup.py` `entry_points` for older packages.

### Rust
- `src/main.rs` → binary crate entry (`fn main()`).
- `src/lib.rs` → library entry.
- Workspaces: `Cargo.toml` `[workspace] members = [...]`; each
  member crate has its own entry.
- Multiple binaries: `[[bin]]` blocks in `Cargo.toml` or a
  `src/bin/` directory.

### JVM (Java / Kotlin / Scala)
- Class with `public static void main(String[] args)` (Java) /
  `fun main()` (Kotlin top-level) / `def main(args: Array[String])`
  (Scala).
- Spring Boot: `@SpringBootApplication` annotated class.
- `pom.xml` / `build.gradle` → look for `<mainClass>` or
  `application { mainClass = … }`.

### Ruby
- `bin/<name>` executable scripts.
- `Gemfile` and `lib/<gem>/version.rb` for gems.
- Rails: `config.ru` (Rack), `bin/rails`.

### Shell scripts / Makefile-driven projects
- `Makefile` `default` or first target.
- `scripts/run.sh`, `scripts/start.sh`.
- `Procfile` (Heroku-style) → declares processes.

### Containerised / orchestrated
- `Dockerfile` `CMD` / `ENTRYPOINT`.
- `docker-compose.yml` `services.<name>.command`.
- `Procfile` declares process types.

## Process

1. **Check the project's stated entry** in README / AGENTS.md
   first. Many projects document it.
2. **Manifest first**: `package.json` / `Cargo.toml` / `go.mod`
   etc. usually points the way.
3. **Convention next**: `cmd/`, `src/`, `bin/` are the usual homes.
4. **Grep for the language's `main` idiom** if all else fails.
5. **Trace ~50 lines** from the entry to confirm — does this code
   actually look like a startup path? If not, you've found a
   library or test entry; keep looking.

## Output shape

```
## Entry point
- File: <path:line>
- Idiom: <e.g. Go func main(), Python __main__, Spring Boot>
- Binary count: <if multi>

## What it does in the first 50 lines
<one paragraph: arg parsing, config load, dependency wiring,
launching the main loop / server / handler>

## Other entry points (if any)
- <path> — <purpose>
```

## When there isn't a single entry

Some projects are libraries, frameworks, or extensions with no main
function. Say so explicitly and identify what consumers call:

- Library → the public API in `doc.go` / `index.ts` / `lib.rs`.
- Framework → the integration points users wire up.
- Plugin / extension → the manifest and the activation hook.
