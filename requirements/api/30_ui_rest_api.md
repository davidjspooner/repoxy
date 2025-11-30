# UI REST API (Draft)

This document captures the initial REST endpoints that power the Repoxy UI. The scope covers read-only navigation (types → repositories → items → versions → files) and excludes live-update/watch behaviour (covered separately).

Base URL: `/api/ui/v1`

Common response conventions:
- All responses use `application/json`.
- Errors follow `{ "error": { "code": "NOT_FOUND", "message": "..." } }`.
- Timestamps are ISO-8601 UTC strings.
- Field names are lower-case with underscores where needed (e.g., `id`, `label`, `description`, `type_id`, `count_repositories`); Go structs should carry explicit `json` tags to enforce this.

### ID Conventions (MVP)

UI-facing IDs are deterministic and reversible to the underlying storage locators:

- **Item ID**: `{repoId}:{host}/{name}`. Example: `dockerhub:docker.io/library/nginx`.
- **Version ID**: `{repoId}:{host}/{name}:{versionLabel}` where `versionLabel` is the friendly tag/label exposed to users (internal UUID remains server-only). Example: `dockerhub:docker.io/library/nginx:v1.27.0`.
- **File ID**: `{repoId}:{host}/{name}:{versionLabel}:{fileName}` where `fileName` is the logical entry name in the version’s metadata. Example: `dockerhub:docker.io/library/nginx:v1.27.0:manifest.json`.

Reverse mapping to CommonStorage:

- Split the ID to recover `repoId`, `host`, `name`, `versionLabel`, and `fileName` (for file IDs).
- Resolve `repoId` to the repository instance.
- Resolve `versionLabel` to the internal version UUID via the instance’s label mapping (e.g., `labels.json` in CommonStorage).
- Use `host/name` and the resolved version ID to load `VersionMeta` via `CommonStorage`.
- Use `fileName` (when present) to locate the file entry within `VersionMeta.Files` and its `blobKey` for content.

Escaping of delimiters is deferred post-MVP; repo IDs and names must avoid `:` in MVP.

---

## 1. Repository Types

`GET /api/ui/v1/repository-types`

Returns the set of repository categories configured on the backend.

```json
{
  "types": [
    {
      "id": "containers",
      "label": "Containers",
      "description": "Container images served via pull-through caches of upstream or private registries",
      "count_repositories": 2
    },
    {
      "id": "terraform",
      "label": "Terraform",
      "description": "Providers delivered from pull-through caches of registry.terraform.io or private catalogs",
      "count_repositories": 1
    },
    {
      "id": "opentofu",
      "label": "OpenTofu",
      "description": "Providers delivered from pull-through caches of registry.opentofu.org or private catalogs",
      "count_repositories": 1
    }
  ]
}
```

---

## 2. Repository Instances per Type

`GET /api/ui/v1/repository-types/{typeId}/repositories`

Returns repositories configured under the specified type.

```json
{
  "type": {
    "id": "containers",
    "label": "Containers"
  },
  "repositories": [
    {
      "id": "dockerhub",
      "label": "Docker Hub",
      "description": "Pull-through cache for registry-1.docker.io",
      "host_count": 1,
      "item_count": 12
    },
    {
      "id": "ghcr",
      "label": "GitHub CR",
      "description": "Mirrors ghcr.io/davidjspooner/*",
      "host_count": 1,
      "item_count": 4
    }
  ]
}
```

---

## 3. Items within a Repository

`GET /api/ui/v1/repositories/{repoId}/items?limit=50&cursor=...`

Lists logical items/containers/packages exposed by the repository. Supports pagination via `cursor`.

```json
{
  "repository": {
    "id": "dockerhub",
    "label": "Docker Hub"
  },
  "items": [
    {
      "id": "dockerhub:library/nginx",
      "label": "library/nginx",
      "detail": "registry-1.docker.io • container image",
      "last_updated": "2025-05-18T10:10:45Z",
      "version_count": 3
    },
    {
      "id": "dockerhub:redis/alpine",
      "label": "redis/alpine",
      "detail": "registry-1.docker.io • container image",
      "last_updated": "2025-05-17T19:29:10Z",
      "version_count": 1
    }
  ],
  "next_cursor": "eyJwYWdlIjoyfQ=="
}
```

---

## 4. Versions for an Item

`GET /api/ui/v1/items/{itemId}/versions?limit=50&cursor=...`

Returns available versions/tags for a single item.

```json
{
  "item": {
    "id": "dockerhub:library/nginx",
    "label": "library/nginx"
  },
  "versions": [
    {
      "id": "dockerhub:library/nginx:v1.27.0",
      "label": "v1.27.0",
      "detail": "Published 2025-05-18",
      "file_count": 2
    },
    {
      "id": "dockerhub:library/nginx:v1.26.2",
      "label": "v1.26.2",
      "detail": "Published 2025-05-05",
      "file_count": 2
    }
  ],
  "next_cursor": null
}
```

---

## 5. Files per Version

`GET /api/ui/v1/versions/{versionId}/files`

Lists files/assets belonging to a specific version (e.g., manifests, metadata, archives).

```json
{
  "version": {
    "id": "dockerhub:library/nginx:v1.27.0",
    "label": "v1.27.0"
  },
  "files": [
    {
      "id": "file-1111",
      "name": "manifest.json",
      "path": "container/dockerhub/library/nginx/v1.27.0/manifest.json",
      "content_type": "application/vnd.oci.image.manifest.v1+json",
      "size_bytes": 812,
      "modified": "2025-05-18T10:10:44Z"
    },
    {
      "id": "file-2222",
      "name": "config.json",
      "path": "container/dockerhub/library/nginx/v1.27.0/config.json",
      "content_type": "application/json",
      "size_bytes": 2048,
      "modified": "2025-05-18T10:10:40Z"
    }
  ]
}
```

---

## 6. File Details

`GET /api/ui/v1/files/{fileId}`

Detailed metadata for a single file reference.

```json
{
  "file": {
    "id": "file-1111",
    "name": "manifest.json",
    "path": "container/dockerhub/library/nginx/v1.27.0/manifest.json",
    "repository_type": "Containers",
    "repository_name": "Docker Hub",
    "item_label": "library/nginx",
    "version_label": "v1.27.0",
    "content_type": "application/vnd.oci.image.manifest.v1+json",
    "size_bytes": 812,
    "checksum": {
      "algorithm": "sha256",
      "value": "1111111111111111111111111111111111111111111111111111111111111111"
    },
    "download_count": 42,
    "last_accessed": "2025-05-18T10:55:00Z"
  }
}
```

---

## 7. Search Filter Helper (Optional)

`GET /api/ui/v1/search/suggestions?q=nginx`

Returns quick suggestions for cross-panel search/filter boxes. This endpoint is optional; UI can filter client-side without it.

```json
{
  "query": "nginx",
  "suggestions": {
    "items": [
      { "id": "dockerhub:library/nginx", "label": "library/nginx (Docker Hub)" }
    ],
    "versions": [
      { "id": "dockerhub:library/nginx:v1.27.0", "label": "v1.27.0 • library/nginx" }
    ],
    "files": [
      { "id": "file-1111", "label": "manifest.json • library/nginx v1.27.0" }
    ]
  }
}
```

---

## Notes & Future Work

- Live update / watch endpoints are out of scope for this draft.
- Authentication/authorization is not specified here; assume existing Repoxy auth applies globally.
- Pagination for item/version endpoints should support both `limit` and `cursor` to keep responses lightweight.
- Consider including `total_count` metadata on list responses once backend supports counting cheaply.
