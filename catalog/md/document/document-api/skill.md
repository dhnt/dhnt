---
name: document-api
description: Document a public API — function signatures, types, examples, error cases, stability
phase: document
user_invocable: true
aliases: [api-docs, godoc]
origin: new
---

# /document-api — document a public API

Write reference documentation for a public API surface. The
audience is a developer who already knows the language and
domain; they need to know what this specific API does.

## What good API documentation contains

For each public symbol (function, type, method, constant):

1. **What it does** — one sentence, present tense, descriptive
   not imperative ("Returns the …" not "Return the …" — match
   the language convention).
2. **Parameters** — name, type, what it represents, valid range
   / format.
3. **Returns** — type, meaning, possible values.
4. **Errors** — what conditions cause errors and which error
   types are returned.
5. **Side effects** — file I/O, network calls, mutation of
   inputs, time-dependent behaviour.
6. **Concurrency** — is it safe to call concurrently? Does it
   block?
7. **Examples** — at least one runnable example for non-trivial
   APIs.
8. **Stability** — alpha / beta / stable, deprecated, since /
   removed-in version.

## Format conventions

- **Go**: doc comments above the symbol. First sentence starts
  with the symbol name. Examples in `_test.go` files as
  `Example<Func>` functions.
- **Python**: docstrings (Google / NumPy / Sphinx style — match
  the project). Type annotations on signatures.
- **TypeScript / JS**: TSDoc / JSDoc above the export. Matches
  IDE tooltips.
- **Rust**: `///` doc comments. Examples in code blocks; tested
  by `cargo test --doc`.
- **OpenAPI / GraphQL**: schema-driven; descriptions on each
  operation and field.

Match the project's existing style. Don't introduce a new
convention.

## Steps

1. **Identify the public API surface.** What's exported? What's
   intended for callers vs internal use?
2. **For each public symbol**, draft the doc using the structure
   above. Skip obvious sections (a getter doesn't need a side-
   effects discussion).
3. **Cite the code.** Examples should be real, runnable, and
   verified to work. Tested examples (Go's `Example*`, Rust's
   doc tests) are best — they don't bit-rot.
4. **Document the invariants** — what the caller can rely on,
   what changes between versions.
5. **Document the gotchas** — non-obvious behaviour, performance
   characteristics, idiomatic vs unidiomatic uses.

## What NOT to do

- Document the implementation, not the contract. The user reads
  this to know how to call the API; how it's implemented may
  change.
- Restate the type signature in prose ("Takes a string and
  returns an int"). The signature already says that.
- Write doc comments that are just the symbol name spelled out.
  ("Get returns the value." for `func (m *Map) Get() T`.) Either
  add value or omit.
- Promise stability the project hasn't committed to.
- Document private symbols as if they're public. Comment them
  for the next maintainer, but don't surface them in API docs.

## Output shape

For each public symbol, a doc block matching the language's
convention:

```go
// Lookup returns the catalog entry registered under name. The lookup
// is case-insensitive and strips leading/trailing whitespace. The
// returned Skill is owned by the package — callers must not mutate.
//
// Returns ok=false if the name (or any of its aliases) does not match
// any catalog entry.
//
// Concurrency: safe for concurrent use.
//
// Example:
//
//	if s, ok := catalog.Lookup("commit"); ok {
//	    fmt.Println(s.Description)
//	}
func Lookup(name string) (Skill, bool)
```

Plus, if the project has a separate API reference:

```
## <Symbol>
<one-paragraph description>

### Parameters
- <name> (<type>): <meaning, range, format>

### Returns
- (<type>): <meaning>

### Errors
- <error type>: <condition>

### Stability
<alpha | beta | stable> since v<…>

### Example
<code block>
```
