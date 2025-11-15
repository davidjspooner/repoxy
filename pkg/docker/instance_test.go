package docker

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidjspooner/go-fs/pkg/storage"
	_ "github.com/davidjspooner/go-fs/pkg/storage/mem"
	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

func newDockerInstanceForTest(t *testing.T, upstream string) *dockerInstance {
	t.Helper()
	ctx := context.Background()
	fsRO, err := storage.OpenFileSystemFromString(ctx, "mem://", storage.Config{})
	if err != nil {
		t.Fatalf("open mem fs: %v", err)
	}
	root, ok := fsRO.(storage.WritableFS)
	if !ok {
		t.Fatalf("fs not writable")
	}
	typeFS, err := root.EnsureSub(ctx, "type/docker")
	if err != nil {
		t.Fatalf("ensure type fs: %v", err)
	}
	f := &factory{}
	if err := f.Initialize(ctx, "docker", typeFS, mux.NewServeMux()); err != nil {
		t.Fatalf("factory init: %v", err)
	}
	inst, err := f.NewRepository(ctx, &repo.Repo{
		Name: "mirror",
		Type: "docker",
		Upstream: repo.Upstream{
			URL: upstream,
		},
		Mappings: []string{"library/*"},
	})
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	dockerInst, ok := inst.(*dockerInstance)
	if !ok {
		t.Fatalf("instance type mismatch")
	}
	return dockerInst
}

func newDockerClientFactory(handler func(req *http.Request) (*http.Response, error)) func() client.Interface {
	return func() client.Interface {
		return client.Func(func(req *http.Request) (*http.Response, error) {
			return handler(req)
		})
	}
}

func httpResponse(status int, headers map[string]string, body []byte) *http.Response {
	resp := &http.Response{
		StatusCode:    status,
		Status:        fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:        make(http.Header),
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	for k, v := range headers {
		resp.Header.Set(k, v)
	}
	return resp
}

func TestDockerManifestAlwaysProxies(t *testing.T) {
	t.Parallel()
	hits := 0
	inst := newDockerInstanceForTest(t, "https://registry.test")
	inst.httpClientFactory = newDockerClientFactory(func(req *http.Request) (*http.Response, error) {
		hits++
		body := []byte(fmt.Sprintf(`{"call":%d}`, hits))
		return httpResponse(http.StatusOK, map[string]string{
			"Content-Type":          "application/vnd.docker.distribution.manifest.v2+json",
			"Docker-Content-Digest": fmt.Sprintf("sha256:deadbeef%d", hits),
		}, body), nil
	})
	param := &param{name: "library/alpine", tag: "latest"}
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v2/library/alpine/manifests/latest", nil)
		rr := httptest.NewRecorder()
		inst.HandleV2Manifest(param, rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("iteration %d: expected 200, got %d", i, rr.Code)
		}
		expectedDigest := fmt.Sprintf("sha256:deadbeef%d", i+1)
		if got := rr.Header().Get("Docker-Content-Digest"); got != expectedDigest {
			t.Fatalf("iteration %d: expected digest %s, got %s", i, expectedDigest, got)
		}
		expectedBody := fmt.Sprintf(`{"call":%d}`, i+1)
		if got := rr.Body.String(); got != expectedBody {
			t.Fatalf("iteration %d: expected body %s, got %s", i, expectedBody, got)
		}
	}
	if hits != 2 {
		t.Fatalf("expected proxy to hit upstream twice, got %d hits", hits)
	}
}

func TestDockerTagsAlwaysProxy(t *testing.T) {
	t.Parallel()
	hits := 0
	inst := newDockerInstanceForTest(t, "https://registry.test")
	inst.httpClientFactory = newDockerClientFactory(func(req *http.Request) (*http.Response, error) {
		hits++
		body := []byte(fmt.Sprintf(`{"seq":%d}`, hits))
		return httpResponse(http.StatusOK, map[string]string{
			"Content-Type": "application/json",
		}, body), nil
	})
	param := &param{name: "library/alpine"}
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v2/library/alpine/tags/list", nil)
		rr := httptest.NewRecorder()
		inst.HandleV2Tags(param, rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("iteration %d: expected 200, got %d", i, rr.Code)
		}
		expectedBody := fmt.Sprintf(`{"seq":%d}`, i+1)
		if rr.Body.String() != expectedBody {
			t.Fatalf("iteration %d: expected %s, got %s", i, expectedBody, rr.Body.String())
		}
	}
	if hits != 2 {
		t.Fatalf("expected upstream tags hit each time, got %d hits", hits)
	}
}

func TestDockerBlobCaching(t *testing.T) {
	t.Parallel()
	layer := []byte("layer-data")
	sum := sha256.Sum256(layer)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	hits := 0
	inst := newDockerInstanceForTest(t, "https://registry.test")
	inst.httpClientFactory = newDockerClientFactory(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/v2/library/alpine/blobs/"+digest {
			hits++
			return httpResponse(http.StatusOK, nil, layer), nil
		}
		return httpResponse(http.StatusNotFound, nil, []byte("not found")), nil
	})
	blobParam := &param{name: "library/alpine", digest: digest}
	req := httptest.NewRequest(http.MethodGet, "/v2/library/alpine/blobs/"+digest, nil)
	rr := httptest.NewRecorder()
	inst.HandleV2BlobByDigest(blobParam, rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if hits != 1 {
		t.Fatalf("expected one upstream blob fetch, got %d", hits)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/v2/library/alpine/blobs/"+digest, nil)
	rr2 := httptest.NewRecorder()
	inst.HandleV2BlobByDigest(blobParam, rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("cached blob request failed: %d", rr2.Code)
	}
	if hits != 1 {
		t.Fatalf("expected cached blob, got %d hits", hits)
	}
	if rr2.Body.String() != string(layer) {
		t.Fatalf("unexpected cached blob body: %s", rr2.Body.String())
	}

	headReq := httptest.NewRequest(http.MethodHead, "/v2/library/alpine/blobs/"+digest, nil)
	headResp := httptest.NewRecorder()
	inst.HandleV2BlobByDigest(blobParam, headResp, headReq)
	if headResp.Code != http.StatusOK {
		t.Fatalf("HEAD request failed: %d", headResp.Code)
	}
	if headResp.Body.Len() != 0 {
		t.Fatalf("expected no body for HEAD, got %d bytes", headResp.Body.Len())
	}
	if hits != 1 {
		t.Fatalf("HEAD should be served from cache without upstream hit, got %d hits", hits)
	}
}
