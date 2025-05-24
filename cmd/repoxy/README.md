# repoxy (command)

This is the entry point for the Repoxy binary. It initializes the HTTP server and exposes metrics.

## Purpose

To run the Repoxy proxy service as a CLI application.

## Usage

```bash
go run ./cmd/repoxy
```

This launches the web service on port 8080 and exposes `/metrics`.

## Files

- `main.go`: starts the Repoxy server and Prometheus metrics
