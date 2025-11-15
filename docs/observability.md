# Observability Guide

Repoxy exposes structured logs and Prometheus-compatible metrics out of the box so operators can trace requests across proxy instances and
monitor cache behaviour. This document explains the signals that Version 0.2 introduced and how to consume them.

## HTTP Logging & Tracing

- Every inbound HTTP request passes through the request tracing middleware. The middleware ensures a correlation ID exists, stores it in
  the request context, and echoes it back via the `X-Request-ID` response header.
- Clients may provide their own `X-Request-ID` header (for example from an upstream reverse proxy). Repoxy will adopt that value so logs
  across the stack use the same identifier.
- The correlation ID is appended to all structured slog entries as both `req_id` and `trace_id`, making it trivial to filter logs for a
  specific request.
- When Repoxy calls upstream registries (Docker Hub, GHCR, registry.terraform.io, etc.) it forwards the same `X-Request-ID` header so
  vendor logs can be correlated with local ones.

### Shipping logs

Repoxy writes structured slog output to stdout/stderr. Recommended pipeline:

1. Run Repoxy under systemd so logs flow into `journald`.
2. Configure `promtail`, `fluent-bit`, or `vector` to tail the journald unit and ship JSON logs to your aggregation layer (Loki, ELK,
   CloudWatch, etc.).
3. Filter on `trace_id`/`req_id` to rebuild end-to-end traces. The structured fields also include `route`, `method`, `status`, and byte
   counts, which makes it straightforward to build dashboards or alerts directly from logs.

## Metrics

Hit `GET /metrics` to scrape Repoxy’s Prometheus/OpenMetrics output. Key series introduced in Version 0.2:

| Metric | Labels | Description / KPI |
| ------ | ------ | ----------------- |
| `repoxy_cache_events_total` | `type`, `repo`, `cache`, `result` | Cache hits/misses/errors for refs, packages, and Docker blobs. Track hit ratio per repo. |
| `repoxy_cache_bytes_total` | `type`, `repo`, `cache`, `action` (`serve`/`store`) | Bytes served from caches vs. bytes written to them. Useful for sizing storage. |
| `repoxy_upstream_requests_total` | `type`, `repo`, `target`, `status` | Counts upstream round trips and failures. Alert on growing `status="error"` counts. |
| `repoxy_upstream_request_duration_seconds` | same labels | Histogram of upstream latency; build SLOs per registry. |
| `http_request_count`, `http_response_time_seconds`, etc. | `method`, `status_code`, `route` | Automatic HTTP middleware metrics for every handler. |
| `repoxy_storage_operations_total`, `repoxy_storage_bytes_total` | `type`, `repo`, `op`, `result` | Low-level storage helper counters (already present before v0.2). |

Example PromQL snippets:

- Cache hit ratio per repo: `sum(rate(repoxy_cache_events_total{result="hit"}[5m])) / sum(rate(repoxy_cache_events_total{result=~"hit|miss"}[5m]))`.
- Upstream error rate: `rate(repoxy_upstream_requests_total{status="error"}[5m])`.
- Bytes served from Terraform package cache: `increase(repoxy_cache_bytes_total{cache="packages",action="serve"}[1h])`.

Document typical KPI expectations (hit ratio > 0.9 for Terraform refs, low upstream error rate) in your operational runbooks.

## Accessing Metrics

The `/metrics` endpoint is mounted on the same host/port as the proxy. Example:

```bash
curl -s https://repoxy.example.com/metrics | head
```

Protect this endpoint with your standard ingress controls (mutual TLS, IP allowlists, etc.) if it should not be publicly accessible.
