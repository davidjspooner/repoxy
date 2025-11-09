package repo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigsErrorsWhenGlobHasNoMatches(t *testing.T) {
	t.Parallel()

	missing := filepath.Join(t.TempDir(), "missing.yaml")
	_, err := LoadConfigs(missing)
	if err == nil || !strings.Contains(err.Error(), missing) {
		t.Fatalf("expected missing file error mentioning %s, got %v", missing, err)
	}
}

func TestLoadConfigsParsesSingleFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filename := filepath.Join(dir, "repoxy.yaml")
	config := `
server:
  listeners:
    - url: http://127.0.0.1
      port: 8080
storage:
  url: mem://
  config: {}
repos:
  - name: alpine
    type: docker
    upstream:
      url: https://registry-1.docker.io
      config: {}
    mappings:
      - library/alpine
`
	if err := os.WriteFile(filename, []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := LoadConfigs(filename)
	if err != nil {
		t.Fatalf("LoadConfigs returned error: %v", err)
	}
	if cfg.Storage == nil || cfg.Storage.URL != "mem://" {
		t.Fatalf("unexpected storage config: %+v", cfg.Storage)
	}
	if len(cfg.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(cfg.Repositories))
	}
	if cfg.Repositories[0].Name != "alpine" {
		t.Fatalf("unexpected repo name %q", cfg.Repositories[0].Name)
	}
}
