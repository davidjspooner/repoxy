# Repoxy – Change Log

## Version 0.2 – Observability & Operations

- **HTTP logging & tracing**
  - Added request-tracing middleware that enforces/echoes `X-Request-ID` for every request and propagates the same ID to upstream registries so logs can be correlated end-to-end.
  - Structured slog output now always includes both `req_id` and `trace_id`, and `docs/observability.md` documents recommended log-shipping pipelines (journald → promtail/fluent-bit → Loki/ELK/etc.).
- **Metrics coverage**
  - New Prometheus metrics track cache hits/misses, bytes written/served, and upstream request counts/latency per repository (`repoxy_cache_*`, `repoxy_upstream_*`).
  - HTTP middleware metrics (`http_request_count`, `http_response_time_seconds`, etc.) are still exposed, and the observability guide lists KPI expectations plus example PromQL queries.

## Version 0.1 – Harden the Core Proxy

- **Terraform/OpenTofu mirror completion**
  - Serve `.well-known/terraform.json` and `/v1/providers/...` so Terraform/OpenTofu can automatically discover the mirror endpoint.
  - Cache provider manifests, version listings, and download metadata inside the shared `refs/` store.
  - Cache provider archives in `packages/` and rewrite download metadata so clients always fetch from Repoxy’s archive endpoints.

- **Docker read-only mirror completion**
  - Proxy Docker Hub and GHCR repositories in strict read-only mode, only caching immutable blobs/layers in the shared blob store.
  - Stream manifests, tags, and catalog data directly from the upstream registries to avoid persisting mutable metadata locally.

- **Upstream auth plumbing for containers**
  - Repositories can now declare an `upstream.auth` block to provide upstream credentials (Docker Hub, GHCR, AWS ECR).
  - The Docker client middleware exchanges those credentials for bearer/basic tokens, caches them per scope, and injects the appropriate `Authorization` headers on retries.
