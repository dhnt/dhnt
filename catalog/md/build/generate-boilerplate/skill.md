---
name: generate-boilerplate
description: Generate the boilerplate scaffold for a new module / component / package, matching the project's conventions
phase: build
user_invocable: true
aliases: [scaffold, new-module]
origin: new
---

# /generate-boilerplate — scaffold a new module

A new component / package / module needs a starting skeleton.
Generate it matching the project's existing conventions, not
generic templates.

## Steps

1. **Identify the convention** by reading 2–3 existing peers in
   the same directory:
   - File naming (`<name>.go` vs `<name>/<name>.go` vs
     `<name>/index.ts`).
   - Test file colocation and naming.
   - Documentation file presence (`doc.go` for Go, `__init__.py`
     for Python, README per package).
   - Public API surface — types, constants, factory functions.
2. **List the files** the new module needs:
   - Source file(s).
   - Test file(s).
   - Doc file (`doc.go` / `README.md`).
   - Anything else the project's peers ship (fixtures, mocks,
     example).
3. **Generate the minimum viable contents**, not stubs. Each file
   should compile and pass minimal tests on creation:
   - Source: package declaration + a single placeholder
     function with a doc comment.
   - Test: package declaration + a single passing test
     exercising the placeholder function.
   - Doc: package overview matching peers' tone and length.
4. **Wire the new module into the rest of the project** if
   needed — the parent package's README, an export index, a
   build manifest.
5. **Run the build + tests** to confirm the scaffold is valid.

## What NOT to do

- Generate from a generic template that doesn't match the
  project. The new module looking like an outsider hurts review
  and adoption.
- Stub out function bodies with `panic("not implemented")` /
  `throw new Error("TODO")` — those break the test suite when
  they accidentally run. Use a real (trivial) implementation.
- Pre-build abstractions you don't need yet. The scaffold should
  be the smallest unit that compiles and tests; expand later.
- Commit the boilerplate before any actual logic. A "scaffold-
  only" commit is fine if labelled clearly; otherwise, ship the
  scaffold + first real content together.

## Output shape

```
Generated module: <name>
Location: <path>
Files:
- <path/to/source>
- <path/to/test>
- <path/to/doc>
Wired into:
- <where it's referenced>
Build: <ok>
Tests: <ok / N pass>
```
