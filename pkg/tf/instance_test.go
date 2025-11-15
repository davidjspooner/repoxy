package tf

import (
	"context"
	"encoding/json"
	"fmt"
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
	refsHelper, err := repo.NewStorageHelper(refsFS, "terraform", "test")
	if err != nil {
		t.Fatalf("NewStorageHelper: %v", err)
	}
	packagesFS, err := proxyFS.EnsureSub(ctx, "packages")
	if err != nil {
		t.Fatalf("EnsureSub packages: %v", err)
	}
	packagesHelper, err := repo.NewStorageHelper(packagesFS, "terraform", "test-packages")
	if err != nil {
		t.Fatalf("NewStorageHelper packages: %v", err)
	}
	inst, err := NewInstance(&repo.Repo{
		Name: "test",
		Upstream: repo.Upstream{
			URL: upstream,
		},
	}, proxyFS, refsHelper, packagesHelper)
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

func TestHandleV1VersionListCachesResult(t *testing.T) {
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
	if hits != 1 {
		t.Fatalf("expected upstream hit once, got %d", hits)
	}
}

func TestHandleV1VersionDownloadCachesPackage(t *testing.T) {
	t.Parallel()
	const (
		namespace = "hashicorp"
		name      = "aws"
		version   = "1.2.3"
		osName    = "linux"
		arch      = "amd64"
		filename  = "terraform-provider-aws_1.2.3_linux_amd64.zip"
	)
	packageHits := 0
	metadataHits := 0
	payload := []byte("zip-bytes")
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case fmt.Sprintf("/v1/providers/%s/%s/%s/download/%s/%s", namespace, name, version, osName, arch):
			metadataHits++
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"download_url": "%s/pkg.zip", "filename": "%s", "os":"%s","arch":"%s"}`, srv.URL, filename, osName, arch)
		case "/pkg.zip":
			packageHits++
			w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	inst := newTFInstanceForTest(t, srv.URL)
	metadataParam := &param{
		namespace: namespace,
		name:      name,
		version:   version,
		tail:      fmt.Sprintf("download/%s/%s", osName, arch),
	}
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/providers/%s/%s/%s/download/%s/%s", namespace, name, version, osName, arch), nil)
	req.Host = "proxy.test"
	rr := httptest.NewRecorder()

	inst.HandleV1VersionDownload(metadataParam, rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for metadata, got %d", rr.Code)
	}
	var metadata map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &metadata); err != nil {
		t.Fatalf("failed to decode metadata: %v", err)
	}
	expectedURL := fmt.Sprintf("http://proxy.test/v1/providers/%s/%s/%s/download/%s/%s/archive/%s", namespace, name, version, osName, arch, filename)
	if metadata["download_url"] != expectedURL {
		t.Fatalf("expected rewritten download url %s, got %s", expectedURL, metadata["download_url"])
	}
	if packageHits != 1 {
		t.Fatalf("expected single upstream package download, got %d", packageHits)
	}

	archiveParam := &param{
		namespace: namespace,
		name:      name,
		version:   version,
		tail:      fmt.Sprintf("download/%s/%s/archive/%s", osName, arch, filename),
	}
	archiveReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/providers/%s/%s/%s/download/%s/%s/archive/%s", namespace, name, version, osName, arch, filename), nil)
	archiveReq.Host = "proxy.test"
	archiveResp := httptest.NewRecorder()
	inst.HandleV1VersionDownload(archiveParam, archiveResp, archiveReq)
	if archiveResp.Code != http.StatusOK {
		t.Fatalf("expected 200 streaming archive, got %d", archiveResp.Code)
	}
	if packageHits != 1 {
		t.Fatalf("expected cached archive without extra upstream fetch, got %d hits", packageHits)
	}
	if got := archiveResp.Body.String(); got != string(payload) {
		t.Fatalf("expected archive payload %q, got %q", string(payload), got)
	}
	if metadataHits != 1 {
		t.Fatalf("expected single metadata fetch, got %d", metadataHits)
	}
}

func TestHandleV1DownloadMetadataCachesResponse(t *testing.T) {
	t.Parallel()
	const (
		namespace = "hashicorp"
		name      = "aws"
		version   = "4.0.0"
		osName    = "darwin"
		arch      = "arm64"
		filename  = "terraform-provider-aws_4.0.0_darwin_arm64.zip"
	)
	var (
		metadataHits int
		packageHits  int
		payload      = []byte("zip-bits")
	)
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case fmt.Sprintf("/v1/providers/%s/%s/%s/download/%s/%s", namespace, name, version, osName, arch):
			metadataHits++
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"download_url": "%s/artifact.zip","filename":"%s"}`, srv.URL, filename)
		case "/artifact.zip":
			packageHits++
			w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	inst := newTFInstanceForTest(t, srv.URL)
	param := &param{
		namespace: namespace,
		name:      name,
		version:   version,
		tail:      fmt.Sprintf("download/%s/%s", osName, arch),
	}

	for i := 0; i < 2; i++ {
		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/providers/%s/%s/%s/download/%s/%s", namespace, name, version, osName, arch), nil)
		request.Host = "proxy.test"
		rec := httptest.NewRecorder()
		inst.HandleV1VersionDownload(param, rec, request)
		if rec.Code != http.StatusOK {
			t.Fatalf("iteration %d: expected 200, got %d", i, rec.Code)
		}
		var metadata map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &metadata); err != nil {
			t.Fatalf("iteration %d: failed to decode metadata: %v", i, err)
		}
		expectedURL := fmt.Sprintf("http://proxy.test/v1/providers/%s/%s/%s/download/%s/%s/archive/%s", namespace, name, version, osName, arch, filename)
		if metadata["download_url"] != expectedURL {
			t.Fatalf("iteration %d: expected rewritten download url %s, got %s", i, expectedURL, metadata["download_url"])
		}
	}
	if metadataHits != 1 {
		t.Fatalf("expected metadata upstream once, got %d", metadataHits)
	}
	if packageHits != 1 {
		t.Fatalf("expected package upstream once, got %d", packageHits)
	}
}
