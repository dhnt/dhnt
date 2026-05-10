---
name: analyze-logs
description: Analyse logs to find a pattern, anomaly, or specific event — without drowning in noise
phase: operate
user_invocable: true
aliases: [log-search, log-investigate]
origin: new
---

# /analyze-logs — find signal in the logs

Production logs are noisy. The skill is finding the relevant
slice without reading the whole stream.

## Steps

1. **State the question.** Are you looking for:
   - A specific event ("when did X first happen?").
   - A pattern ("how often does Y occur?").
   - An anomaly ("why was traffic abnormal at 10:30?").
   - A correlation ("do A and B happen together?").
   The query shape depends on the question.
2. **Pick the time window.** Wider than you think for "first
   occurrence" questions; narrower for high-volume incidents.
   Don't query the last 30 days when 10 minutes will do.
3. **Filter aggressively** before searching for keywords:
   - **Service / component**: scope to the relevant binary.
   - **Severity**: ERROR / WARN are usually where investigations
     start; INFO drowns the signal.
   - **Trace / request ID**: if you have one, this is the most
     selective filter available.
   - **Tenant / user**: for user-specific incidents.
4. **Search for keywords** within the filtered slice:
   - Error message substrings.
   - Stack-trace function names.
   - Specific values from the user's report.
5. **Aggregate, don't read line-by-line.** Group-by counts
   (per-minute, per-status-code, per-route) often surface the
   pattern faster than scanning.
6. **Correlate across services** if the project has a
   distributed-tracing tool. Single-service log analysis misses
   most multi-hop bugs.
7. **Capture findings**: what you searched for, what you found,
   what slice supports each conclusion. Attach the query so
   someone else can re-run it.

## Common log query tools

- **Grafana Loki / Cloud Logs / Datadog Logs / Splunk**: LogQL /
  KQL / SPL syntax — learn the project's. Most support the same
  shape: `service=X severity>=ERROR | grep "<keyword>"`.
- **journalctl / `kubectl logs`**: for direct access; combine
  with `grep -A 5 -B 5` for context.
- **OTEL collector / Jaeger / Tempo**: trace-driven, follow a
  single request across services.

## What NOT to do

- Read the raw log stream when a query would work.
- Filter on a keyword and conclude based on counts without
  reading at least 5 sample messages — the keyword may match
  unrelated noise.
- Confuse correlation with causation. Two events happening
  together at the same time isn't a causal link.
- Forget the time zone. UTC vs local-time confusion has wasted
  many hours in incident response.
- Capture findings as "I found a few things" without the
  underlying queries. Reproducibility matters.

## Output shape

```
## Question
<what you were looking for>

## Window
<time range, time zone>

## Filters
<service, severity, trace ID, etc.>

## Findings
- <observation> — supporting query: `<query>`
- <observation> — supporting query: `<query>`

## Hypothesis
<what the data suggests, with confidence level>

## Next investigations
- <what to look at if the hypothesis doesn't hold>
```
