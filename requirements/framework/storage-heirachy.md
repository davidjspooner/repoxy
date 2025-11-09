# Repoxy Storage Layout

This document describes the centralized storage design for Repoxy, built on top of the **go-fs** abstraction.
It defines the filesystem hierarchy, pathing rules, and storage guarantees for all proxied artifact types (Containers, Terraform, OpenTofu, Debian, etc.).

---

## 1. Overview

All repositories share a **single writable filesystem** (the shared storage root) mounted at a configured root.
That root may be backed by any `go-fs` backend such as `file://`, `s3://`, or `mem://`.
Each proxy operates within a sandboxed subdirectory under this root and cannot escape its assigned area.

### Supported Modes

Repoxy eventually needs to support two repository modes:

1. **Read-only pull-through cache** – proxy forwards requests to an upstream, optionally persisting immutable blobs/refs locally. This is the current and **only** mode implemented in the MVP.
2. **Writable local origin** – proxy accepts client writes and serves content solely from local storage with no upstream.

While the storage layout already reserves space for metadata/cache/tmp directories that a future writable mode could reuse, **do not implement write paths yet**. All code should assume pull-through caching semantics until the project formally graduates writable repos from the backlog.

---

## 2. Top-Level Directory Structure

```
<root>/
├── config/                # global configuration (placeholder, optional)
└── type/
    ├── container/         # container artifacts (Docker, GHCR, ECR, etc.)
    │   ├── config/        # per-type configuration and metadata
    │   ├── blobs/         # shared content-addressable layers
    │   └── proxies/       # per-proxy caches and references
    ├── tf/                # Terraform / OpenTofu artifacts
    │   ├── config/
    │   ├── blobs/
    │   └── proxies/
    ├── debian/            # (potential future support for Debian packages)
    └── generic/           # other artifact types
```

Each artifact type (e.g., `container`, `tf`) maintains:

* A **`config/`** directory for type-specific configuration or metadata.
* A **`blobs/`** directory for shared, content-addressable storage.
* A **`proxies/`** directory for individual proxy instances.

---

## 3. Proxy-Specific Hierarchy

Each proxy has its own sandboxed directory under `type/<type>/proxies/<proxy-name>/`.

Examples of proxy names might use internal or code designations rather than backends. The following uses Marvel superhero codenames as examples:

```
type/container/proxies/
├── ironman/
│   ├── refs/              # human-readable references (e.g., tags, manifests)
│   ├── meta/              # metadata and descriptors
│   ├── cache/             # cached HTTP responses
│   └── tmp/               # temp uploads and staging areas
├── blackwidow/
│   ├── refs/
│   ├── meta/
│   ├── cache/
│   └── tmp/
└── thor/
    ├── refs/
    ├── meta/
    ├── cache/
    └── tmp/
```

Different proxies may connect to the same backend with separate authentication tokens or access scopes.
For example, multiple proxies might each map to Docker Hub or GHCR while using distinct credentials.

All proxies of a given artifact type share that type’s blob store (`type/<type>/blobs/`) to enable deduplication and efficient reuse.

---

## 4. Blob Layout

Blobs are content-addressable and organized using digest-based sharding:

```
type/<type>/blobs/
└── sha256/
    ├── ab/cd/<digest>       # individual content-addressable blobs
    └── ...
```

* **Sharding** prevents performance issues from large, flat directories.
* **Immutability:** blobs are never modified after writing.
* **Atomic writes:** uploads complete via a temporary staging path before being finalized.

---

## 5. Reference Files

Small JSON files act as pointers from logical identifiers (tags, versions, etc.) to immutable blobs.

### Example: Container Tag Reference

`refs/<namespace>/<name>/tags/<tag>.json`

```json
{
  "kind": "container.tag",
  "targetDigest": "sha256:…",
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "size": 7432,
  "upstream": "<registry-host>",
  "fetchedAt": "2025-11-09T12:34:56Z",
  "ttlSeconds": 604800
}
```

### Example: Terraform Provider Reference

`refs/providers/<namespace>/<name>/<platform>/<version>.json`

```json
{
  "kind": "tf.provider",
  "zipDigest": "sha256:…",
  "size": 12345678,
  "upstream": "registry.terraform.io",
  "fetchedAt": "2025-11-09T12:34:56Z",
  "ttlSeconds": 604800
}
```

These reference files are compact, human-readable, and may be managed through the administrative UI.

---

## 6. Guarantees

* **Namespace isolation:** Each proxy only accesses its sandboxed directory.
* **Type-level deduplication:** Blobs are shared within each artifact type.
* **Safe path handling:** All operations use `FullPath` normalization to prevent escapes.
* **Backend neutrality:** Works identically across file, S3, and memory-backed storage.
* **Observability:** Storage operations can be consistently logged and instrumented.

---

## 7. Future Enhancements after Minimum Viable Product

* **Atomicity:** Uploads occur in `tmp/uploads/<uuid>` and are committed to their final location only after validation.
* **Concurrency:** Lock files under `/locks/<scoped-key>` prevent concurrent conflicting writes.
* **Garbage Collection:**

  1. Collect all reachable digests from `refs/`.
  2. Remove expired references.
  3. Delete unreferenced blobs after a grace period.
* **Metrics:** Cache hits, misses, bytes written, and cleanup stats are exposed through the metrics system.

---
