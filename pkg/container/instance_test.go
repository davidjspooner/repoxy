package container

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	awsecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/davidjspooner/go-fs/pkg/storage"
	_ "github.com/davidjspooner/go-fs/pkg/storage/mem"
	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

func newContainerInstanceForTest(t *testing.T, upstream string) *containerInstance {
	cfg := &repo.Repo{
		Name: "mirror",
		Type: "container",
		Upstream: repo.Upstream{
			URL: upstream,
		},
		Mappings: []string{"library/*"},
	}
	return newContainerInstanceFromConfig(t, cfg)
}

func newContainerInstanceFromConfig(t *testing.T, cfg *repo.Repo) *containerInstance {
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
	typeFS, err := root.EnsureSub(ctx, "type/container")
	if err != nil {
		t.Fatalf("ensure type fs: %v", err)
	}
	f := &factory{}
	common, err := repo.NewCommonStorageWithLabels(typeFS, "container", "container")
	if err != nil {
		t.Fatalf("common storage: %v", err)
	}
	if err := f.Initialize(ctx, "container", mux.NewServeMux()); err != nil {
		t.Fatalf("factory init: %v", err)
	}
	inst, err := f.NewRepository(ctx, common, cfg)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	containerInst, ok := inst.(*containerInstance)
	if !ok {
		t.Fatalf("instance type mismatch")
	}
	return containerInst
}

func newContainerClientFactory(handler func(req *http.Request) (*http.Response, error)) func() client.Interface {
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

func TestContainerManifestAlwaysProxies(t *testing.T) {
	t.Parallel()
	hits := 0
	inst := newContainerInstanceForTest(t, "https://registry.test")
	inst.httpClientFactory = newContainerClientFactory(func(req *http.Request) (*http.Response, error) {
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

func TestContainerTagsAlwaysProxy(t *testing.T) {
	t.Parallel()
	hits := 0
	inst := newContainerInstanceForTest(t, "https://registry.test")
	inst.httpClientFactory = newContainerClientFactory(func(req *http.Request) (*http.Response, error) {
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

func TestContainerManifestAuthWithBearerCredentials(t *testing.T) {
	t.Parallel()
	cfg := &repo.Repo{
		Name: "mirror",
		Type: "container",
		Upstream: repo.Upstream{
			URL: "https://registry.test",
			Auth: &repo.UpstreamAuth{
				Provider: "dockerhub",
				Config: map[string]string{
					"username": "demo",
					"password": "secret",
				},
			},
		},
		Mappings: []string{"library/*"},
	}
	inst := newContainerInstanceFromConfig(t, cfg)
	tokenRequests := 0
	inst.tokenHTTP = httpDoFunc(func(req *http.Request) (*http.Response, error) {
		tokenRequests++
		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("demo:secret"))
		if got := req.Header.Get("Authorization"); got != expected {
			t.Fatalf("token request authorization = %s, want %s", got, expected)
		}
		resp := httpResponse(http.StatusOK, map[string]string{
			"Content-Type": "application/json",
		}, []byte(`{"token":"test-token","expires_in":120}`))
		return resp, nil
	})
	auth, err := newContainerUpstreamAuth(inst.tokenHTTP, inst.config.Upstream)
	if err != nil {
		t.Fatalf("newContainerUpstreamAuth: %v", err)
	}
	inst.auth = auth
	upstreamHits := 0
	inst.httpClientFactory = newContainerClientFactory(func(req *http.Request) (*http.Response, error) {
		upstreamHits++
		switch upstreamHits {
		case 1:
			resp := httpResponse(http.StatusUnauthorized, map[string]string{
				"WWW-Authenticate": `Bearer realm="https://auth.registry.test/token",service="registry.test",scope="repository:library/alpine:pull"`,
			}, nil)
			resp.Request = req
			return resp, nil
		case 2:
			if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Fatalf("upstream request Authorization = %s, want Bearer test-token", got)
			}
			resp := httpResponse(http.StatusOK, nil, []byte(`{"ok":true}`))
			resp.Request = req
			return resp, nil
		default:
			return httpResponse(http.StatusInternalServerError, nil, nil), nil
		}
	})
	req := httptest.NewRequest(http.MethodGet, "/v2/library/alpine/manifests/latest", nil)
	rr := httptest.NewRecorder()
	inst.HandleV2Manifest(&param{name: "library/alpine", tag: "latest"}, rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if upstreamHits != 2 {
		t.Fatalf("expected 2 upstream hits, got %d", upstreamHits)
	}
	if tokenRequests != 1 {
		t.Fatalf("expected single token request, got %d", tokenRequests)
	}
}

func TestContainerManifestAuthWithECR(t *testing.T) {
	t.Parallel()
	token := base64.StdEncoding.EncodeToString([]byte("AWS:password"))
	expires := time.Now().Add(time.Hour)
	fakeClient := &fakeECRClient{
		token:   token,
		expires: expires,
	}
	orig := newECRClient
	newECRClient = func(cfg aws.Config) ecrAPI {
		return fakeClient
	}
	defer func() {
		newECRClient = orig
	}()
	cfg := &repo.Repo{
		Name: "mirror",
		Type: "container",
		Upstream: repo.Upstream{
			URL: "https://123456789012.dkr.ecr.us-east-1.amazonaws.com",
			Auth: &repo.UpstreamAuth{
				Provider: "ecr",
				Config: map[string]string{
					"region":            "us-east-1",
					"access_key_id":     "AKIA",
					"secret_access_key": "SECRET",
					"registry_id":       "123456789012",
				},
			},
		},
		Mappings: []string{"library/*"},
	}
	inst := newContainerInstanceFromConfig(t, cfg)
	upstreamHits := 0
	inst.httpClientFactory = newContainerClientFactory(func(req *http.Request) (*http.Response, error) {
		upstreamHits++
		switch upstreamHits {
		case 1:
			resp := httpResponse(http.StatusUnauthorized, map[string]string{
				"WWW-Authenticate": `Basic realm="https://ecr.amazonaws.com"`,
			}, nil)
			resp.Request = req
			return resp, nil
		case 2:
			if got := req.Header.Get("Authorization"); got != "Basic "+token {
				t.Fatalf("upstream Authorization = %s, want Basic %s", got, token)
			}
			resp := httpResponse(http.StatusOK, nil, []byte(`{"ok":true}`))
			resp.Request = req
			return resp, nil
		default:
			return httpResponse(http.StatusInternalServerError, nil, nil), nil
		}
	})
	req := httptest.NewRequest(http.MethodGet, "/v2/library/alpine/manifests/latest", nil)
	rr := httptest.NewRecorder()
	inst.HandleV2Manifest(&param{name: "library/alpine", tag: "latest"}, rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if upstreamHits != 2 {
		t.Fatalf("expected 2 upstream hits, got %d", upstreamHits)
	}
	if fakeClient.calls != 1 {
		t.Fatalf("expected single ECR token fetch, got %d", fakeClient.calls)
	}
}

func TestContainerBlobCaching(t *testing.T) {
	t.Parallel()
	layer := []byte("layer-data")
	sum := sha256.Sum256(layer)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	hits := 0
	inst := newContainerInstanceForTest(t, "https://registry.test")
	inst.httpClientFactory = newContainerClientFactory(func(req *http.Request) (*http.Response, error) {
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

type fakeECRClient struct {
	token   string
	expires time.Time
	calls   int
}

func (f *fakeECRClient) GetAuthorizationToken(ctx context.Context, params *awsecr.GetAuthorizationTokenInput, optFns ...func(*awsecr.Options)) (*awsecr.GetAuthorizationTokenOutput, error) {
	f.calls++
	return &awsecr.GetAuthorizationTokenOutput{
		AuthorizationData: []awsecrtypes.AuthorizationData{
			{
				AuthorizationToken: aws.String(f.token),
				ExpiresAt:          aws.Time(f.expires),
			},
		},
	}, nil
}
