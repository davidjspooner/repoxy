# pkg/repo

This package defines the interfaces and registration helpers for repository types (Docker, Terraform, etc.) used by Repoxy.

## Purpose

Provide a single place where repository types register themselves, receive the shared storage root, wire their HTTP routes, and create
repository instances.

## Interfaces

- `Instance` – implemented by any concrete repository implementation.
- `Type` – exposes:
  - `Initialize(ctx, typeName, mux)` – register HTTP handlers and prepare any type-level storage subtrees.
  - `NewRepository(ctx, common, config)` – construct an `Instance` for a specific repo configuration.

## Usage

```go
type dockerType struct{}

func init() {
    repo.MustRegisterType("docker", &dockerType{})
}

func (d *dockerType) Initialize(ctx context.Context, typeName string, mux *mux.ServeMux) error {
    mux.HandleFunc("GET /v2/", d.handlePing)
    return nil
}

func (d *dockerType) NewRepository(ctx context.Context, common repo.CommonStorage, cfg *repo.Repo) (repo.Instance, error) {
    if common == nil {
        return nil, fmt.Errorf("docker type not initialized")
    }
    return NewDockerInstance(cfg, common), nil
}
```

At startup the server calls:

```go
fs, _ := repo.NewStorageRoot(ctx, cfg.Storage)
_ = repo.Initialize(ctx, fs, mux)
instance, _ := repo.NewRepository(ctx, repoConfig)
```
