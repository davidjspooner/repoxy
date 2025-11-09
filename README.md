# Repoxy

Repoxy is a Go service that fronts multiple artifact repositories (Docker, Terraform, future formats) and proxies their upstream sources
through one observable, cache-aware endpoint. Every repository implementation plugs into a shared storage root (powered by `go-fs`) and
inherits the same HTTP middleware (metrics, logging, recovery) so new backends can be added with minimal boilerplate.

This README gives a high-level orientation. Architectural details and guardrails now live in the `requirements/` folder so they can evolve
independently without bloating this overview.

## Quick Start

```bash
cd repoxy
go run ./cmd/repoxy --config conf/repoxy.yaml
# or build
go build -o bin/repoxy ./cmd/repoxy
./bin/repoxy serve --config conf/repoxy.yaml
```

The server will parse the YAML config, connect to the shared storage root, instantiate each configured repository, and expose HTTP endpoints
plus `/metrics`.

## Repository Layout

```
repoxy/
├── cmd/repoxy         # CLI entry point
├── pkg/
│   ├── repo           # config loader, factory registry, storage root
│   ├── docker         # Docker proxy implementation
│   ├── cache          # HTTP response caching helpers
│   ├── listener       # listener configuration helpers
│   └── tf             # placeholder for Terraform/OpenTofu logic
├── conf/              # sample configuration
├── requirements/      # living design docs and guardrails
└── docs/              # additional background notes
```

## Key Documents

- `requirements/general-intro.md` – project mission, terminology, coding guardrails, and expectations for contributors/agents.
- `requirements/framework/storage-heirachy.md` – canonical layout for the shared storage root (blobs, refs, caching, GC rules).
- `go-fs/pkg/storage/README.md` – details of the filesystem abstraction used by Repoxy.

Consult those documents whenever you touch configuration contracts, storage behavior, or repository interfaces. They are the single source
of truth for design decisions and should be updated alongside any substantial change.

## Contributing

- Keep tests green (`go test ./...`).
- Update the relevant requirement doc when architectural or configuration semantics change.
- Register new repository types via `pkg/repo` factories and use the shared storage root.

Questions or proposals? Start by reviewing the requirements docs above; they outline the current direction and help keep the project
coherent as it grows.
