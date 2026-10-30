# Production Observability Checklist

> Sources: https://pkg.go.dev/log/slog; https://prometheus.io/docs/practices/histograms/; https://opentelemetry.io/docs/languages/go/
> Authority: advisory
> Minimum Go: `cmp.Or` 1.22; `http.Request.Pattern` 1.23; `slog.NewMultiHandler` 1.26
> Last verified: 2026-10-30

> **Advisory**: for a production service, in the stack the project already
> runs. The metric shapes below are Prometheus terms because that is the
> common case — translate them for OpenTelemetry or a vendor agent rather
> than adding a second stack, and skip the checklist for a library or CLI.

A feature in a service is not done until an operator can see it fail:

- **Metrics** — counters for operations and errors, histograms for latency
  (histograms aggregate across instances; summaries do not), buckets picked
  from the operation's expected latency (`prometheus.ExponentialBuckets`):
  `DefBuckets` start at 5ms and flatten a sub-millisecond p99. A counter ends
  in `_total`, a unit in `_seconds` or `_bytes`. Exemplars reach Prometheus
  only over OpenMetrics (`promhttp.HandlerOpts{EnableOpenMetrics: true}`).
  Keep the query that reads a metric next to its declaration.
- **Cardinality** — label values stay bounded (method, route pattern, status);
  never user IDs, full URLs, or request bodies. The route label is the
  pattern the mux will match, read before `next` with
  `_, route := mux.Handler(r)` (`""` for a 404 or 405), labelled
  `cmp.Or(route, "unmatched")` (Go 1.22+). The `Pattern` field (Go 1.23+) of
  a request the middleware passed on is set only when `next` is the mux
  itself: a wrapper in between that passes `r.WithContext(…)`, like
  go-security's auth wrapper, leaves it empty.
- **Logs** — structured key-value records carrying the request or trace ID,
  which needs the enriched logger or context handler described in
  [Request-Scoped Logging](../SKILL.md#request-scoped-logging). In an
  OpenTelemetry stack that handler reads `sc := trace.SpanContextFromContext(ctx)`
  and, when `sc.IsValid()`, adds `trace_id`/`span_id` from
  `sc.TraceID().String()`/`sc.SpanID().String()`, never a key of its own, so
  logs join the traces. `otelslog.NewHandler` writes to the global OTel
  `LoggerProvider`, a no-op until `global.SetLoggerProvider` installs one with
  an exporter: records sent only there vanish, so keep the stdout handler
  beside it with `slog.NewMultiHandler`.
- **Traces** — in an OpenTelemetry stack, `otel.SetTextMapPropagator` with
  `propagation.TraceContext{}` beside `otel.SetTracerProvider`; with baggage,
  one call with `propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})`,
  since a second `Set…` call replaces the first. The global propagator is a no-op, so
  without it `otelhttp.NewHandler` and `otelhttp.NewTransport` neither read
  nor send `traceparent` and each service starts its own trace. A failed span
  gets `span.RecordError(err)` and `span.SetStatus(codes.Error, …)`, since
  `RecordError` alone leaves it unset, and `tp.Shutdown(ctx)` at exit flushes
  the last batch. A ratio sampler is wrapped,
  `sdktrace.ParentBased(sdktrace.TraceIDRatioBased(r))`: a bare one decides
  again in every service and splits the trace.
- **Dashboards and alerts** — a metric nobody queries is not observability:
  land each one in the project's dashboards and alert rules, or say plainly
  that it shipped unwired.
- **Profiles** — mount `net/http/pprof` on a separate internal listener,
  never on the public mux ([go-security](../../go-security/SKILL.md#http-surface)
  owns the rule); [go-troubleshooting](../../go-troubleshooting/SKILL.md) owns
  reading them.
