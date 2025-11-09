# pkg/repo

This package defines the interfaces and registration helpers for repository types (Docker, Terraform, etc.) used by Repoxy.

## Purpose

Provide a single place where repository types register themselves, receive the shared storage root, wire their HTTP routes, and create
repository instances.

## Interfaces

- `Instance` – implemented by any concrete repository implementation.
- `Type` – exposes:
  - `Initialize(ctx, typeName, fs, mux)` – register HTTP handlers and prepare any type-level storage subtrees.
  - `NewRepository(ctx, config)` – construct an `Instance` for a specific repo configuration.

## Usage

```go
type dockerType struct{
    fs storage.WritableFS
}

func init() {
    repo.MustRegisterType("docker", &dockerType{})
}

func (d *dockerType) Initialize(ctx context.Context, typeName string, fs storage.WritableFS, mux *mux.ServeMux) error {
    d.fs = fs
    mux.HandleFunc("GET /v2/", d.handlePing)
    return nil
}

func (d *dockerType) NewRepository(ctx context.Context, cfg *repo.Repo) (repo.Instance, error) {
    if d.fs == nil {
        return nil, fmt.Errorf("docker type not initialized")
    }
    return NewDockerInstance(cfg, d.fs), nil
}
```

At startup the server calls:

```go
fs, _ := repo.NewStorageRoot(ctx, cfg.Storage)
_ = repo.Initialize(ctx, fs, mux)
instance, _ := repo.NewRepository(ctx, repoConfig)
```
