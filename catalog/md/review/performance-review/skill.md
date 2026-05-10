---
name: performance-review
description: Review pending changes specifically for performance regressions — allocations, complexity, hot paths
phase: review
user_invocable: true
aliases: [perf-review]
origin: new
---

# /performance-review — performance audit

A code review with a performance lens. A general /code-review
asks "is this correct?"; a /performance-review asks "is this
still fast?"

## What to look for

- **Algorithmic complexity changes.** Did O(n) become O(n²)? A
  nested loop where there used to be a hash map?
- **Hot-path allocations.** New allocations inside loops, maps
  resized repeatedly, slices growing without preallocation,
  string concatenation in a tight loop.
- **N+1 patterns.** A loop that issues one DB / RPC / file-read
  per iteration when a batch call would do.
- **Synchronous work inside request handlers** that could be
  async (logging to disk synchronously, heavy serialisation).
- **Lock contention.** New mutexes around hot data structures,
  expanding the critical section, holding locks across I/O.
- **Cache invalidation.** New writes that invalidate a hot cache
  on every call.
- **Memory leaks.** Long-lived structures collecting entries
  without bound, contexts holding references after request
  completion.
- **Regex / parser hot paths.** Regex compilation per call instead
  of once, repeated parsing of the same data.
- **Cold-start regressions.** Heavy initialisation, large embedded
  data, blocking I/O on startup.

## Process

1. Read the diff once for surface area.
2. For every loop in the diff, ask: "what's the inner cost?"
3. For every new function, ask: "what's the call frequency?"
4. For every new abstraction (interface, generic, dynamic
   dispatch), ask: "does this go through a hot path?"
5. Run the project's benchmarks if it has them
   (`go test -bench`, `cargo bench`, `pytest --benchmark`,
   k6 / wrk for HTTP). Compare to baseline.
6. Profile if anything is suspicious (`pprof`, `perf record`,
   browser devtools profiler). Reading code is fine for finding
   issues; profiling confirms or refutes the hypothesis.

## What NOT to do

- Speculate about cost without measuring or citing the code.
  "This looks slow" without evidence isn't actionable.
- Suggest micro-optimisations the compiler already does (most
  runtimes constant-fold, inline trivially-callable functions,
  CSE common subexpressions).
- Block a PR for "could be faster" when the path isn't actually
  hot. Optimise where the profiler points; ignore where it
  doesn't.
- Recommend a complete rewrite. Surgical suggestions land; "redo
  this" doesn't.

## Output shape

```
## Summary
<verdict in 1-2 sentences>

## Concerns (N)
- <file:line> — <issue> — <expected impact: order of magnitude>

## Hardening opportunities (N)
- <file:line> — <improvement>

## Benchmarks
- Suite: <ran / didn't run>
- Result: <delta from baseline if measured>
```

## Severity bands

- **Regression** — measured slower than the baseline (or strongly
  predicted to be on a hot path). Block merge until addressed.
- **Concern** — looks slow on a hot path; ask for benchmark data
  to confirm or refute.
- **Hardening** — defensible micro-improvement; non-blocking,
  optional.
