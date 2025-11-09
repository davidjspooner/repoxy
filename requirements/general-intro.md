# Repoxy – Project Introduction & Guardrails

This document gives coding agents the context needed to evolve Repoxy confidently. Keep it open while working on any task; it explains the
project’s purpose, key architecture, and the rules that must not be broken.

---

## 1. Purpose & Vision

Repoxy is a Go service that fronts multiple artifact repositories (Docker, Terraform, future formats) and proxies their upstream sources.
It strives to:

1. **Provide one proxy endpoint** for many registries with unified auth, logging, and metrics.
2. **Share an extensible storage layer** (via go-fs) for caching immutable artifacts.
3. **Offer a plug-in model** so new repository types can be added without touching core code.

Success means smooth onboarding of additional repository types while maintaining observability and storage hygiene.

---

## 2. Vocabulary

| Term              | Meaning                                                                                  |
|-------------------|------------------------------------------------------------------------------------------|
| Repository Type   | A plug-in (e.g., `docker`) registered via `pkg/repo.Type`.                                |
| Storage Root      | The shared writable `go-fs` filesystem handed to every repository.                       |
| Upstream          | The remote artifact service being proxied (Docker Hub, GHCR, registry.terraform.io).     |
| Mapping           | A glob-style rule that determines which requests a repository instance handles.          |

---

## 3. Architecture Snapshot

1. **CLI Entry Point:** `cmd/repoxy` (`serve` command) loads YAML config, mounts the shared storage root, registers HTTP middleware,
   and builds each repository via the factory registry.
2. **Repository Core (`pkg/repo`):** Handles config loading/validation, factory registration, and the storage root helper
   (`NewStorageRoot`).
3. **Backends (`pkg/docker`, future packages):** Implement repo-specific logic but must only rely on interfaces provided by `pkg/repo`
   and `go-fs`.
4. **Caching Layer (`pkg/cache`):** Implements `middleware.CacheImpl` to store HTTP responses inside the shared storage root.
5. **Storage (go-fs):** Provides `ReadOnlyFS`/`WritableFS` implementations for file/S3/GitHub/memory backends with consistent error
   semantics. See `go-fs/pkg/storage/README.md` for details.

The current storage layout and invariants are defined in `requirements/framework/storage-heirachy.md`. Refer to it whenever writing to
the filesystem.

---

## 4. Design Guardrails

1. **Repository Types Everywhere:** All repository types must register via `repo.MustRegisterType`. Never instantiate repositories directly.
2. **Shared Storage Only:** Repository code should always use the provided `storage.WritableFS` root (backed by go-fs) or sub-filesystems derived from it. No direct `os` calls in
   proxy code unless they’re guarded by the storage abstraction.
3. **Fail Fast Configs:** Configuration parsing must reject unknown fields or duplicated sections. When adding new YAML keys, update both
   the schema and documentation.
4. **Consistent Error Codes:** Use `storage.Errorf(...).WithMessage(...)` or well-defined error enums so clients and tests can detect
   failure reasons (`EOPEN`, `ECONFIG`, `ECREATE`, etc.).
5. **Logging & Metrics:** All HTTP handlers must run through the middleware stack (metrics, structured logging, recovery). When adding
   new HTTP paths, register them on the mux provided by `cmd/repoxy`.
6. **Streaming by Default:** When moving large blobs, use `storage.Writer` or another streaming approach instead of buffering entire
   files in memory.
7. **Path Safety:** Always resolve user paths via `storage.FullPath` (or the equivalent helper) to avoid directory traversal or
   symlink-escape issues.
8. **Deterministic I/O:** `ReadDir`, `Glob`, and similar operations must return sorted results where the calling code relies on order.

---

## 5. Coding Expectations for Agents

- **Tests:** Run `go test ./...` inside `repoxy` (and `go-fs` when editing storage code). Add tests whenever practical.
- **Docs:** Update `README.md` or `requirements/` docs when changing architecture, storage layout, or configuration contracts.
- **Error Handling:** Wrap errors with context (`fmt.Errorf("...: %w", err)`). Surface actionable messages to operators.
- **Concurrency:** Protect shared maps (e.g., factory registries) with mutexes. Avoid global state unless protected.
- **Imports:** For storage backends or repo types that depend on side effects, import them using blank identifiers in the appropriate
  package (`_ "github.com/.../storage/localfile"`).
- **Security:** Never allow repositories to escape their assigned storage sandbox or to bypass authentication hooks once implemented.

---

## 6. Non-Goals (Guardrail)

- Building a full UI or admin console inside this repo (future project).
- Re-implementing artifact formats (OCI, Terraform modules) from scratch—Repoxy proxies existing upstreams.
- Introducing ad-hoc filesystem access: all persistent storage must flow through go-fs.

---

## 7. References for Further Detail

- `repoxy/README.md` – high-level project overview, configuration, and extension steps.
- `requirements/framework/storage-heirachy.md` – canonical storage layout and lifecycle semantics.
- `go-fs/pkg/storage/README.md` – storage backends, interfaces, and design principles.
- `pkg/repo` package docs – interfaces and registration rules.

Keep this document updated whenever large-scale decisions change. It is the contract that future coding agents rely on when modifying
Repoxy. Failure to keep it in sync will slow development and introduce regressions.
