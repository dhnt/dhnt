---
name: query-metrics
description: Query production metrics to answer a quantitative question about service behaviour
phase: operate
user_invocable: true
aliases: [metrics, dashboard]
origin: new
---

# /query-metrics — answer a question with metrics

Logs tell you what happened to one request; metrics tell you what
happened to all of them. Use this skill to answer "how many", "how
fast", "how often", "what changed" questions.

## Steps

1. **State the question quantitatively.** Vague: "is the service
   slow?" Specific: "is the p95 latency for /search higher this
   week than last week?" The query you write depends on the
   specific form.
2. **Identify the right metric.** Common shapes:
   - **Rate / counter**: `requests_total`, `errors_total`. Query
     as `rate(...[5m])`.
   - **Gauge**: `cpu_usage`, `memory_resident`, `queue_depth`.
     Query as the raw value or with `avg_over_time`.
   - **Histogram**: latency, sizes. Query as
     `histogram_quantile(0.95, ...)` for p95.
   - **Summary**: pre-computed quantiles; less flexible than
     histograms but lighter.
3. **Pick the right time window** for the question. Drift /
   regression questions: weeks. Incident questions: minutes-hours.
4. **Aggregate across the right dimensions.** `sum by(route)`,
   `avg by(instance)`. Wrong aggregation hides the signal — e.g.
   averaging error rate across all routes can mask a route that's
   100% failing because the others are at 0%.
5. **Compare to a baseline.** A number alone tells you nothing;
   compared to a week ago, an hour ago, or a SLO threshold it
   tells you something. `_offset 1w` (Prometheus) or analogous.
6. **Read the dashboard** — many questions are answered by a
   panel someone already built. Check the project's standard
   dashboards before composing a new query.

## Common metric tools

- **Prometheus / Mimir / VictoriaMetrics**: PromQL. Most agentic
  coding workflows touch this stack.
- **Grafana**: the dashboard / explore frontend over the above.
- **Datadog / New Relic / CloudWatch**: vendor variants;
  syntax differs but shapes match.
- **OTEL collector** + Prometheus exporter: increasingly the
  default for new projects.

## What NOT to do

- Eyeball a graph and conclude without checking the y-axis.
  Linear vs log scales mislead at a glance.
- Aggregate in a way that hides the signal (averaging when you
  should sum, or vice versa).
- Compare metrics from different time zones or different sample
  windows.
- Trust a single sample. Latency p95 at one moment isn't a SLO
  breach — sustained breach over a window is.
- Build a custom query when the project has a standard panel
  for it. Cross-team consistency matters in incidents.

## Output shape

```
## Question
<quantitative form>

## Query
<actual PromQL / SQL / vendor query>

## Result
<the number or graph summary>

## Compared to baseline
<offset comparison>

## Conclusion
<what the data shows, in one paragraph>

## Caveats
<sampling window, missing data, known biases>
```
