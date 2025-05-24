# pkg/repo

This package defines interfaces and registration for proxy repository backends.

## Purpose

To support pluggable backend repository types (e.g., Docker, General, Terraform) using a global factory model.

## Interfaces

- `Instance` – implemented by any repository proxy
- `Factory` – creates `Instance` objects using `NewRepo(ctx, config)`

## Usage

```go
repo.MustRegisterFactory("docker", docker.Factory{})
instance, err := repo.NewRepositoryFactory(ctx, "docker", config)
```

## Files

- `interface.go` – shared interfaces and global factory registry
