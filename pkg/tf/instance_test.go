package tf

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/davidjspooner/go-fs/pkg/storage"
	_ "github.com/davidjspooner/go-fs/pkg/storage/mem"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

func newTFInstanceForTest(t *testing.T, upstream string) *tfInstance {
	t.Helper()
	ctx := context.Background()
	fsRO, err := storage.OpenFileSystemFromString(ctx, "mem://", storage.Config{})
	if err != nil {
		t.Fatalf("failed to create mem fs: %v", err)
	}
	root, ok := fsRO.(storage.WritableFS)
	if !ok {
		t.Fatalf("fs not writable")
	}
	proxyFS, err := root.EnsureSub(ctx, "proxies/test")
	if err != nil {
		t.Fatalf("EnsureSub proxies: %v", err)
	}
	refsFS, err := proxyFS.EnsureSub(ctx, "refs")
	if err != nil {
		t.Fatalf("EnsureSub refs: %v", err)
	}
	refsHelper, err := repo.NewStorageHelper(refsFS)
	if err != nil {
		t.Fatalf("NewStorageHelper: %v", err)
	}
	inst, err := NewInstance(&repo.Repo{
		Name: "test",
		Upstream: repo.Upstream{
			URL: upstream,
		},
	}, proxyFS, refsHelper)
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	return inst
}

func TestHandleV1VersionCachesManifest(t *testing.T) {
	t.Parallel()

	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"version": "1.2.3",
		})
	}))
	defer srv.Close()

	inst := newTFInstanceForTest(t, srv.URL)
	p := &param{namespace: "hashicorp", name: "aws", version: "1.2.3"}

	req := httptest.NewRequest(http.MethodGet, "/v1/providers/hashicorp/aws/1.2.3", nil)
	rr := httptest.NewRecorder()
	inst.HandleV1Version(p, rr, req)
	if hits != 1 {
		t.Fatalf("expected 1 upstream hit, got %d", hits)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/providers/hashicorp/aws/1.2.3", nil)
	rr2 := httptest.NewRecorder()
	inst.HandleV1Version(p, rr2, req2)
	if hits != 1 {
		t.Fatalf("expected cached response without upstream hit, got %d hits", hits)
	}
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}

	reader, err := inst.refs.Open(context.Background(), path.Join("providers", "hashicorp", "aws", "1.2.3.json"))
	if err != nil {
		t.Fatalf("expected manifest stored locally: %v", err)
	}
	reader.Close()
}

func TestHandleV1VersionListAlwaysHitsUpstream(t *testing.T) {
	t.Parallel()
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"versions": []string{"1.0.0"},
		})
	}))
	defer srv.Close()

	inst := newTFInstanceForTest(t, srv.URL)
	p := &param{namespace: "hashicorp", name: "aws"}

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/providers/hashicorp/aws/versions", nil)
		rr := httptest.NewRecorder()
		inst.HandleV1VersionList(p, rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	}
	if hits != 2 {
		t.Fatalf("expected upstream hit each time, got %d", hits)
	}
}
