package container

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/observability"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// containerInstance implements the repo.Instance interface for Container repositories.
type containerInstance struct {
	factory           *factory
	storage           repo.CommonStorage
	config            repo.Repo
	pipeline          client.MiddlewarePipeline
	nameMatchers      repo.NameMatchers // Matchers for repository names
	httpClientFactory func() client.Interface
	tokenHTTP         httpDoer
	auth              *containerUpstreamAuth
}

// NewContainerInstance creates a new Container repository instance.
// It initializes the instance with the factory and configuration, and sets up the authentication middleware.
func NewContainerInstance(factory *factory, storage repo.CommonStorage, config *repo.Repo) (*containerInstance, error) {
	if storage == nil {
		return nil, fmt.Errorf("docker instance missing storage")
	}
	instance := &containerInstance{
		factory: factory,
		storage: storage,
		config:  *config,
	}
	instance.nameMatchers.Set(config.Mappings)
	instance.pipeline = append(instance.pipeline, client.WithAuthentication(instance))
	instance.httpClientFactory = func() client.Interface {
		return &http.Client{}
	}
	instance.tokenHTTP = &http.Client{}
	auth, err := newContainerUpstreamAuth(instance.tokenHTTP, config.Upstream)
	if err != nil {
		return nil, err
	}
	instance.auth = auth
	return instance, nil
}

// Ensure dockerInstance implements the repo.Instance , client.Authenticator, client.Cache interfaces.
var _ repo.Instance = (*containerInstance)(nil)
var _ client.Authenticator = (*containerInstance)(nil)

func (d *containerInstance) GetMatchWeight(name []string) int {
	return d.nameMatchers.GetMatchWeight(name)
}

func (d *containerInstance) Describe() repo.InstanceMeta {
	label := d.config.Name
	if label == "" {
		label = "containers"
	}
	typeID := d.config.Type
	if typeID == "container" {
		typeID = "containers"
	}
	return repo.InstanceMeta{
		ID:          d.config.Name,
		Label:       label,
		Description: d.config.Description,
		TypeID:      typeID,
	}
}

// HandledWriteMethodForReadOnlyRepo checks if the request is a write operation and returns a 405 if so.
// Returns true if the request was handled (i.e., is not allowed), false otherwise.
func (d *containerInstance) HandledWriteMethodForReadOnlyRepo(w http.ResponseWriter, r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		return false // Read operations are allowed
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("This repo is read-only, no write operations allowed"))
		return true
	}
}

// HandleV2Tags handles container V2 tags requests. Returns a 405 for write operations.
func (d *containerInstance) HandleV2Tags(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	if err := d.proxyToUpstream(r.Context(), w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to proxy docker tags request", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

// HandleV2Manifest handles container V2 manifest requests. Returns a 405 for write operations.
func (d *containerInstance) HandleV2Manifest(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	err := d.proxyToUpstream(r.Context(), w, r)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to proxy request to upstream", "error", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

// HandleV2BlobUpload handles container V2 blob upload requests. Returns a 405 for write operations.
func (d *containerInstance) HandleV2BlobUpload(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	http.Error(w, "Repository is read-only; uploads are not supported", http.StatusMethodNotAllowed)
}

// HandleV2BlobUID handles container V2 blob UID requests. Returns a 405 for write operations.
func (d *containerInstance) HandleV2BlobUID(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	http.Error(w, "Repository is read-only; uploads are not supported", http.StatusMethodNotAllowed)
}

// HandleV2BlobByDigest handles container V2 blob digest requests. Returns a 405 for write operations.
func (d *containerInstance) HandleV2BlobByDigest(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	if param == nil || param.digest == "" || d.storage == nil {
		_ = d.proxyToUpstream(r.Context(), w, r)
		return
	}
	if d.serveLocalBlob(param, w, r) {
		return
	}
	if err := d.fetchAndStoreBlob(param, w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to proxy blob request", "error", err)
	}
}

func (d *containerInstance) Authenticate(response *http.Response) string {
	if response == nil || d.auth == nil {
		return ""
	}
	header, err := d.auth.authorization(response)
	if err != nil {
		slog.ErrorContext(response.Request.Context(), "failed to build upstream auth header", "error", err)
		return ""
	}
	return header
}

func (d *containerInstance) Config() repo.Repo {
	return d.config
}

func (d *containerInstance) proxyToUpstream(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	resp, err := d.roundTripUpstream(ctx, r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	d.writeHeadersFromResponse(w, resp)
	w.WriteHeader(resp.StatusCode)
	if resp.Body == nil {
		return nil
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

func (d *containerInstance) roundTripUpstream(ctx context.Context, r *http.Request) (*http.Response, error) {
	u, err := url.Parse(d.config.Upstream.URL)
	if err != nil {
		return nil, err
	}
	u.Path = r.URL.Path
	u.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(ctx, r.Method, u.String(), r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header.Clone()
	observability.ApplyRequestIDHeader(req, observability.RequestIDFromRequest(r))
	httpClient := d.httpClientFactory
	var base client.Interface
	if httpClient != nil {
		base = httpClient()
	} else {
		base = &http.Client{}
	}
	base = d.pipeline.WrapClient(base)
	start := time.Now()
	resp, err := base.Do(req)
	elapsed := time.Since(start)
	repoType, repoName := d.repoLabels()
	if err != nil {
		observability.ObserveUpstreamRequest(repoType, repoName, u.Host, 0, err, elapsed)
		return nil, err
	}
	observability.ObserveUpstreamRequest(repoType, repoName, u.Host, resp.StatusCode, nil, elapsed)
	return resp, nil
}

func (d *containerInstance) writeHeadersFromResponse(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
}

func (d *containerInstance) serveLocalBlob(param *param, w http.ResponseWriter, r *http.Request) bool {
	reader, err := d.storage.OpenBlob(r.Context(), param.digest)
	if err != nil {
		d.recordCacheMiss(observability.CacheBlobs)
		return false
	}
	d.recordCacheHit(observability.CacheBlobs)
	defer reader.Close()
	if info, err := d.storage.StatBlob(r.Context(), param.digest); err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		d.recordCacheBytes(observability.CacheBlobs, "serve", info.Size())
	}
	w.Header().Set("Docker-Content-Digest", param.digest)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return true
	}
	if _, err := io.Copy(w, reader); err != nil {
		slog.ErrorContext(r.Context(), "failed to stream cached blob", "error", err)
		d.recordCacheError(observability.CacheBlobs)
	}
	return true
}

func (d *containerInstance) fetchAndStoreBlob(param *param, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	resp, err := d.roundTripUpstream(ctx, r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	d.writeHeadersFromResponse(w, resp)
	w.Header().Set("Docker-Content-Digest", param.digest)
	w.WriteHeader(resp.StatusCode)

	if resp.Body == nil || r.Method == http.MethodHead {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		_, err = io.Copy(w, resp.Body)
		return err
	}

	counter := &writeCounter{w: w}
	tee := io.TeeReader(resp.Body, counter)
	n, err := d.storage.PutBlob(ctx, param.digest, tee)
	if err != nil {
		d.recordCacheError(observability.CacheBlobs)
		return err
	}
	if n == 0 {
		n = counter.n
	}
	d.recordCacheBytes(observability.CacheBlobs, "store", n)
	return nil
}

func (d *containerInstance) repoLabels() (string, string) {
	repoType := d.config.Type
	if repoType == "" {
		repoType = "container"
	}
	repoName := d.config.Name
	if repoName == "" {
		repoName = "default"
	}
	return repoType, repoName
}

func (d *containerInstance) recordCacheHit(cache string) {
	repoType, repoName := d.repoLabels()
	observability.RecordCacheHit(repoType, repoName, cache)
}

func (d *containerInstance) recordCacheMiss(cache string) {
	repoType, repoName := d.repoLabels()
	observability.RecordCacheMiss(repoType, repoName, cache)
}

func (d *containerInstance) recordCacheError(cache string) {
	repoType, repoName := d.repoLabels()
	observability.RecordCacheError(repoType, repoName, cache)
}

func (d *containerInstance) recordCacheBytes(cache, action string, n int64) {
	if n <= 0 {
		return
	}
	repoType, repoName := d.repoLabels()
	observability.RecordCacheBytes(repoType, repoName, cache, action, n)
}

type writeCounter struct {
	w io.Writer
	n int64
}

func (c *writeCounter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}
