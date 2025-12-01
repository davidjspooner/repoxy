# Registry Abstraction Design

This document describes the proposed **registry abstraction** that sits on top of a `go-fs` filesystem and provides a uniform model for naming, metadata, and blob storage across multiple artifact types (Container, Terraform/OpenTofu, Debian, PyPI, etc.).

The goal is to:
- Centralise how hosts, names, versions, labels, and files are represented.
- Provide a stable, reusable API that proxy implementations can depend on, instead of each proxy hand-rolling its own layout.
- Keep storage backend–agnostic by building on the existing `storage.WritableFS` interface.

---

## 1. Conceptual Model

The registry abstraction exposes a uniform model of artifacts:

- **Host** – top-level namespace (e.g. `docker.io`, `registry.terraform.io`, `pypi.org`).
- **Name** – repository/package identifier, potentially hierarchical (e.g. `library/nginx`, `hashicorp/aws`, `myorg/mypkg`).
- **Version** – an immutable snapshot of the package contents, identified by a UUID (v1) string. These UUIDs are minted **only** by Repoxy; they never reuse upstream identifiers or digests even if two registries carry the same payload. Repoxy is currently a single-instance system, so it can guarantee unique IDs without distributed coordination.
- **Label** – a human-friendly alias that points to a specific version (e.g. `latest`, `stable`, `1.2`, `v1`). Multiple labels may point to the same version.
- **Files** – concrete artifacts belonging to a version (manifests, zips, .deb files, JSON descriptors, signatures, etc.).

Two physical areas exist in storage:

1. **Metadata/index area** – all naming, versioning, and file-list information.
2. **Blob area** – content-addressable storage for actual file payloads, keyed by cryptographic hashes.

Repositories only interact with this abstraction; they do not directly choose path layouts in the underlying filesystem.

---

## 2. Filesystem Layout

### 2.1 Top-Level Layout

The registry abstraction assumes a single writable filesystem root (backed by any `go-fs` backend).

At that root, we introduce two main directories:

```text
<root>/
├── metadata/      # all index + version metadata
└── blobs/         # shared content-addressable storage for file payloads
```

The existing `type/*/proxies/*` layout can be adapted to delegate into this abstraction, or this abstraction can be mounted under a known prefix (e.g. `type/generic/registry/`). 

**Per-instance rooting:** Each repository instance must receive its own isolated filesystem subtree (e.g. `type/<type>/<repoName>/`) before constructing a `CommonStorage`. Instances must not share a CommonStorage root across repos; the layout above lives inside that per-instance root.

### 2.2 Metadata Layout

Metadata is organised hierarchically by host and name. Under each name we store:

- A set of **version metadata files**, one per immutable version (keyed by UUID).
- A **labels index**, mapping labels (tags) to version IDs.

```text
metadata/
└── index/
    └── <host>/
        └── <name-path>/
            ├── labels.json
            └── versions/
                ├── <version-uuid-1>.json
                ├── <version-uuid-2>.json
                └── ...
```

Where:

- `<host>` is whatever canonical form the **adapter** chooses (e.g. `docker.io`, `registry.terraform.io`). Each adapter applies the normalization rules appropriate for its protocol; `CommonStorage` treats the string as opaque.
- `<name-path>` is the repository/package path, split on `/` according to adapter-specific rules. For example:
  - `library/nginx`
  - `hashicorp/aws`
  - `simplejson`
- `labels.json` maps text labels to version UUIDs.
- Each `versions/<uuid>.json` file describes **one immutable version**, including its file list and references into the `blobs/` area.

#### 2.2.1 `labels.json`

Per host+name, `labels.json` is a single small JSON document:

```jsonc
{
  "kind": "registry.labels",
  "host": "docker.io",
  "name": "library/nginx",
  "labels": {
    "latest": "c0b2c020-6d1d-11ef-9b3f-0242ac120003",
    "1.25":  "bcd61f3e-6d1d-11ef-9b3f-0242ac120003",
    "stable": "bcd61f3e-6d1d-11ef-9b3f-0242ac120003"
  },
  "updatedAt": "2025-11-20T12:34:56Z"
}
```

Notes:

- Labels are **not** stored in `versions/<uuid>.json`; moving a label only requires editing `labels.json`.
- Multiple labels may point at the same version UUID.
- The file remains small even for repos with many versions.
- Repoxy runs as a single instance, so `CommonStorage` can serialize updates to this file locally without cross-node consensus.

#### 2.2.2 `versions/<uuid>.json`

Each version file contains the full file list and any additional metadata for that immutable version. The `versionId` value is purely internal to Repoxy; adapters must not infer any relationship between a Repoxy version ID and upstream tags, digests, or SemVer strings. Adapters **may embed canonical descriptors** (e.g., a container manifest JSON) directly in the version metadata while also recording the blob location for the same content; this keeps the manifest available even if the blob is pruned or unavailable.

```jsonc
{
  "kind": "registry.version",
  "versionId": "bcd61f3e-6d1d-11ef-9b3f-0242ac120003",
  "host": "docker.io",
  "name": "library/nginx",
  "createdAt": "2025-11-20T12:00:00Z",
  "manifest": "{...raw manifest JSON...}", // optional, canonical descriptor cached inline
  "files": [
    {
      "name": "manifest.json",
      "blobKey": "sha256:1c2b3d...",
      "size": 7432,
      "mediaType": "application/vnd.oci.image.manifest.v1+json"
    },
    {
      "name": "layer-0.tar.gz",
      "blobKey": "sha256:aa11bb...",
      "size": 20971520,
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip"
    },
    {
      "name": "layer-1.tar.gz",
      "blobKey": "sha256:cc33dd...",
      "size": 10485760,
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip"
    }
  ]
}
```

- `versionId` is a **UUID v1** (time + node ID) generated when the version is created.
- `blobKey` is an algorithm-prefixed digest string (e.g. `sha256:<hex>`).
- The version file is the single source of truth for the file list; labels simply point to `versionId`.

### 2.3 Blobs Layout

The `blobs/` tree is content-addressable, sharded by algorithm and digest prefix (similar to the existing design). 

```text
blobs/
└── sha256/
    ├── ab/cd/<full-digest>
    ├── ef/01/<full-digest>
    └── ...
```

Key points:

- **Immutability** – blobs are never modified once written.
- **Deduplication** – if two versions reference the same `blobKey`, the file is stored once and shared.
- **Backend neutrality** – the path layout works with any `go-fs` backend (local disk, S3, in-memory). 

---

## 3. CommonStorage API

The registry abstraction wraps a `storage.WritableFS` (from `go-fs/pkg/storage/interface.go`) and exposes a higher-level API in terms of hosts, names, versions, labels, and files. Rather than two interfaces, we provide a concrete helper that can be embedded or reused wherever a filesystem handle is available.

### 3.1 Core Parameter Structure

Most operations share the same address-like parameters. We capture that in a reusable struct:

```go
package registry

// Locator identifies a logical object within the registry.
//
// Fields are intentionally simple; absent/empty fields are ignored
// by operations that do not require them.
type Locator struct {
    Host     string // e.g. "docker.io", "registry.terraform.io"
    Name     string // repository/package path, e.g. "library/nginx", "hashicorp/aws"

    // Exactly one of Label or VersionID is typically used by a call.
    Label     string // human label, e.g. "latest", "stable", "1.2.3"
    VersionID string // UUID string that identifies an immutable version

    // Optional: identifies a single file within a version, when relevant.
    FileName string // e.g. "manifest.json", "package.whl", "provider.zip"
}
```

Suggested conventions:

- `Host` and `Name` are always required.
- Methods that operate on labels expect `Label` to be non-empty.
- Methods that operate on a specific version expect `VersionID` to be non-empty.
- Methods dealing with an individual file inside a version also use `FileName`.

### 3.2 Metadata Types

```go
package registry

// FileEntry describes one file belonging to a version.
type FileEntry struct {
    Name      string `json:"name"`      // logical name within the version
    BlobKey   string `json:"blobKey"`   // e.g. "sha256:<hex>"
    Size      int64  `json:"size"`      // bytes
    MediaType string `json:"mediaType"` // optional, but strongly recommended
}

// VersionMeta mirrors versions/<uuid>.json on disk.
type VersionMeta struct {
    Kind      string      `json:"kind"`      // "registry.version"
    VersionID string      `json:"versionId"`
    Host      string      `json:"host"`
    Name      string      `json:"name"`
    CreatedAt time.Time   `json:"createdAt"`
    Files     []FileEntry `json:"files"`
}

// LabelBindings mirrors labels.json on disk.
type LabelBindings struct {
    Kind      string            `json:"kind"`      // "registry.labels"
    Host      string            `json:"host"`
    Name      string            `json:"name"`
    Labels    map[string]string `json:"labels"`    // label -> versionId
    UpdatedAt time.Time         `json:"updatedAt"`
}

// VersionSummary is a lightweight view for listing.
type VersionSummary struct {
    VersionID string    `json:"versionId"`
    CreatedAt time.Time `json:"createdAt"`
    // Optional: derived info like number of files, total size, etc.
}
```

### 3.3 CommonStorage Structure

```go
package registry

import (
    "io"
    "go-fs/pkg/storage"
)

// CommonStorage owns the filesystem handles used for metadata + blob access.
// It can be constructed once per proxy and reused across incoming requests.
type CommonStorage struct {
    fs storage.WritableFS
}

func NewCommonStorage(fs storage.WritableFS) *CommonStorage {
    return &CommonStorage{fs: fs}
}
```

`CommonStorage` encapsulates all filesystem concerns (paths, atomic updates, sharding) and exposes higher-level helpers that the protocol adapters consume. Adapters may implement caches or full repositories; in either case they rely on the same object for reads and for mutations such as cache refreshes.

### 3.4 CommonStorage Methods

All methods accept a `context.Context` for cancellation/timeouts. The same methods cover both read-only and writable semantics; repositories that only require reads simply ignore the mutation helpers.

```go
// Discovery --------------------------------------------------------------

// ListHosts enumerates every <host> folder under metadata/index/.
// Results are returned in ascending string sort order and always include every host.
func (s *CommonStorage) ListHosts(ctx context.Context) ([]string, error)

// ListNamesForHost enumerates <name-path> entries for a given host.
// Results are sorted lexicographically; no server-side filtering or pagination is applied.
func (s *CommonStorage) ListNamesForHost(ctx context.Context, host string) ([]string, error)

// Label & Version Resolution --------------------------------------------

func (s *CommonStorage) ResolveLabel(ctx context.Context, loc Locator) (Locator, error)
func (s *CommonStorage) GetLabels(ctx context.Context, loc Locator) (*LabelBindings, error)

// Version Metadata -------------------------------------------------------

func (s *CommonStorage) GetVersionMeta(ctx context.Context, loc Locator) (*VersionMeta, error)
func (s *CommonStorage) ListVersions(ctx context.Context, loc Locator) ([]VersionSummary, error)

// Blob Access ------------------------------------------------------------

func (s *CommonStorage) OpenBlob(ctx context.Context, blobKey string) (io.ReadCloser, error)
func (s *CommonStorage) PutBlob(ctx context.Context, blobKey string, r io.Reader) error

// Mutations --------------------------------------------------------------

func (s *CommonStorage) CreateVersion(ctx context.Context, loc Locator, meta *VersionMeta) (Locator, error)
func (s *CommonStorage) SetLabel(ctx context.Context, loc Locator) error
func (s *CommonStorage) DeleteLabel(ctx context.Context, loc Locator) error
func (s *CommonStorage) DeleteVersion(ctx context.Context, loc Locator) error
```

Implementation details:

- `NewCommonStorage` always receives a `storage.WritableFS`. Even adapters that only read data obtain the same handle; they simply avoid invoking mutation helpers.
- The discovery helpers return lexicographically sorted slices with no server-side filtering or pagination; clients perform any additional filtering in their own layer. This satisfies MVP requirements for listing endpoints (PyPI simple, Terraform provider listings, etc.) without leaking filesystem layouts.
- Each mutation helper is responsible for generating new Repoxy-specific version IDs, writing metadata through temporary files + renames, and coordinating label updates atomically.
- Because Repoxy runs as a single instance, CommonStorage can guarantee consistent version-ID allocation and serialized `labels.json` writes without distributed locking.
- Implementations should guard metadata mutations with a process-wide `sync.RWMutex` (or similar) to ensure atomic read/write behavior even when multiple adapters share a CommonStorage instance.

---

## 4. Mapping to Concrete Artifact Types

This abstraction is intentionally generic enough to support multiple backends:

- **Docker / OCI images Container registry**  
  - Host: `registry-1.docker.io`  
  - Name: `library/nginx`  
  - Files per version: one manifest JSON, N layer blobs, optional config JSON.
- **Terraform / OpenTofu providers**  
  - Host: `registry.terraform.io`  
  - Name: `hashicorp/aws`  
  - Files per version: zipped provider binary (`provider_<os>_<arch>.zip`), `.sha256sum`, optional signatures.
- **Debian repositories**  
  - Host: `deb.example.com`  
  - Name: `pool/main/n/nginx` (or a logical package name)  
  - Files per version: `.deb` file(s), `Packages` index snapshots if desired.
- **PyPI packages**  
  - Host: `pypi.org`  
  - Name: `simplejson`  
  - Files per version: sdist `.tar.gz`, wheel `.whl`, metadata JSON.

Each adapter (Container, Terraform, Debian, PyPI, etc.):

1. Parses incoming HTTP requests into a `(Host, Name, Label/VersionID, FileName)` triple.
2. Uses `CommonStorage` helpers to resolve labels, fetch `VersionMeta`, and open blobs.
3. Handles protocol-specific concerns (response headers, auth, range requests, etc.) without knowing the underlying path layout.

---

## 5. Summary

- A **hierarchical metadata layout** under `metadata/index/<host>/<name>/` provides clear separation of hosts, names, versions, and labels.
- **Per-version JSON files** (`versions/<uuid>.json`) describe immutable versions, while `labels.json` maps labels to versions.
- A shared **content-addressable blob store** under `blobs/<algo>/<shard>/<digest>` deduplicates file payloads.
- The concrete **`CommonStorage` helper** exposes this model in terms of `Locator`, `VersionMeta`, blob keys, and discovery helpers, hiding the filesystem details behind `go-fs`.

This design allows you to implement the abstraction once and reuse it across all repository types, while staying compatible with the existing storage forest and backends. 
