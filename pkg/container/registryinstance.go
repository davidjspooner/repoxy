package container

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/observability"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// containerRegistryInstance implements the repo.Instance interface for Container registries.
type containerRegistryInstance struct {
	factory           *factory
	storage           repo.CommonStorage
	config            repo.Repo
	pipeline          client.MiddlewarePipeline
	nameMatchers      repo.NameMatchers // Matchers for repository names
	httpClientFactory func() client.Interface
	tokenHTTP         client.Interface
	auth              *containerUpstreamAuth
}

// newContainerRegistryInstance creates a new Container repository instance.
// It initializes the instance with the factory and configuration, and sets up the authentication middleware.
func newContainerRegistryInstance(factory *factory, storage repo.CommonStorage, config *repo.Repo) (*containerRegistryInstance, error) {
	if storage == nil {
		return nil, fmt.Errorf("docker instance missing storage")
	}
	instance := &containerRegistryInstance{
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
var _ repo.Instance = (*containerRegistryInstance)(nil)
var _ client.Authenticator = (*containerRegistryInstance)(nil)

func (d *containerRegistryInstance) GetMatchWeight(name []string) int {
	return d.nameMatchers.GetMatchWeight(name)
}

func (d *containerRegistryInstance) Describe() repo.InstanceMeta {
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
func (d *containerRegistryInstance) HandledWriteMethodForReadOnlyRepo(w http.ResponseWriter, r *http.Request) bool {
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
func (d *containerRegistryInstance) HandleV2Tags(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	if err := d.proxyToUpstream(r.Context(), w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to proxy docker tags request", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

// HandleV2Manifest handles container V2 manifest requests. Returns a 405 for write operations.
func (d *containerRegistryInstance) HandleV2Manifest(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	ctx := r.Context()
	resp, err := d.roundTripUpstream(ctx, r)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err == nil && resp != nil && resp.StatusCode >= http.StatusOK && resp.StatusCode < 300 && resp.Body != nil && r.Method == http.MethodGet {
		body, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			d.cacheManifest(ctx, param, resp.Header.Get("Docker-Content-Digest"), resp.Header.Get("Content-Type"), body)
			d.writeHeadersFromResponse(w, resp)
			w.WriteHeader(resp.StatusCode)
			_, _ = w.Write(body)
			return
		}
		err = readErr
	}
	if d.serveCachedManifest(param, w, r) {
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "Failed to proxy manifest request to upstream", "error", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	d.writeHeadersFromResponse(w, resp)
	w.WriteHeader(resp.StatusCode)
	if resp.Body != nil && r.Method != http.MethodHead {
		_, _ = io.Copy(w, resp.Body)
	}
}

// HandleV2BlobUpload handles container V2 blob upload requests. Returns a 405 for write operations.
func (d *containerRegistryInstance) HandleV2BlobUpload(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	http.Error(w, "Repository is read-only; uploads are not supported", http.StatusMethodNotAllowed)
}

// HandleV2BlobUID handles container V2 blob UID requests. Returns a 405 for write operations.
func (d *containerRegistryInstance) HandleV2BlobUID(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	http.Error(w, "Repository is read-only; uploads are not supported", http.StatusMethodNotAllowed)
}

// HandleV2BlobByDigest handles container V2 blob digest requests. Returns a 405 for write operations.
func (d *containerRegistryInstance) HandleV2BlobByDigest(param *param, w http.ResponseWriter, r *http.Request) {
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

func (d *containerRegistryInstance) Authenticate(response *http.Response) string {
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

func (d *containerRegistryInstance) Config() repo.Repo {
	return d.config
}

func (d *containerRegistryInstance) proxyToUpstream(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

func (d *containerRegistryInstance) roundTripUpstream(ctx context.Context, r *http.Request) (*http.Response, error) {
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

func (d *containerRegistryInstance) writeHeadersFromResponse(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
}

func (d *containerRegistryInstance) serveLocalBlob(param *param, w http.ResponseWriter, r *http.Request) bool {
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

func (d *containerRegistryInstance) fetchAndStoreBlob(param *param, w http.ResponseWriter, r *http.Request) error {
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

func (d *containerRegistryInstance) repoLabels() (string, string) {
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

func (d *containerRegistryInstance) recordCacheHit(cache string) {
	repoType, repoName := d.repoLabels()
	observability.RecordCacheHit(repoType, repoName, cache)
}

func (d *containerRegistryInstance) recordCacheMiss(cache string) {
	repoType, repoName := d.repoLabels()
	observability.RecordCacheMiss(repoType, repoName, cache)
}

func (d *containerRegistryInstance) recordCacheError(cache string) {
	repoType, repoName := d.repoLabels()
	observability.RecordCacheError(repoType, repoName, cache)
}

func (d *containerRegistryInstance) recordCacheBytes(cache, action string, n int64) {
	if n <= 0 {
		return
	}
	repoType, repoName := d.repoLabels()
	observability.RecordCacheBytes(repoType, repoName, cache, action, n)
}

func (d *containerRegistryInstance) cacheManifest(ctx context.Context, param *param, digest, mediaType string, body []byte) {
	if d.storage == nil || param == nil || len(body) == 0 {
		return
	}
	if digest == "" {
		sum := sha256.Sum256(body)
		digest = "sha256:" + hex.EncodeToString(sum[:])
	}
	if digest == "" {
		return
	}
	if _, err := d.storage.PutBlob(ctx, digest, bytes.NewReader(body)); err != nil {
		d.recordCacheError(observability.CacheManifests)
		return
	}
	d.recordCacheBytes(observability.CacheManifests, "store", int64(len(body)))

	loc := repo.Locator{
		Host:      d.upstreamHost(),
		Name:      param.name,
		VersionID: digest,
	}
	meta := &repo.VersionMeta{
		VersionID: digest,
		Files: []repo.FileEntry{{
			Name:      param.tag,
			BlobKey:   digest,
			Size:      int64(len(body)),
			MediaType: mediaType,
		}},
		Manifest: string(body),
	}
	loc, err := d.storage.CreateVersion(ctx, loc, meta)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		d.recordCacheError(observability.CacheManifests)
		return
	}
	loc.VersionID = digest
	loc.Label = param.tag
	if err := d.storage.SetLabel(ctx, loc); err != nil {
		d.recordCacheError(observability.CacheManifests)
		return
	}
}

func (d *containerRegistryInstance) serveCachedManifest(param *param, w http.ResponseWriter, r *http.Request) bool {
	if d.storage == nil || param == nil || param.tag == "" || param.name == "" {
		return false
	}
	ctx := r.Context()
	loc := repo.Locator{
		Host:  d.upstreamHost(),
		Name:  param.name,
		Label: param.tag,
	}
	loc, err := d.storage.ResolveLabel(ctx, loc)
	if err != nil {
		d.recordCacheMiss(observability.CacheManifests)
		return false
	}
	meta, err := d.storage.GetVersionMeta(ctx, loc)
	if err != nil || meta == nil || len(meta.Files) == 0 {
		d.recordCacheError(observability.CacheManifests)
		return false
	}
	file := meta.Files[0]
	manifest := []byte(meta.Manifest)
	if len(manifest) == 0 {
		reader, err := d.storage.OpenBlob(ctx, file.BlobKey)
		if err != nil {
			d.recordCacheError(observability.CacheManifests)
			return false
		}
		defer reader.Close()
		manifest, err = io.ReadAll(reader)
		if err != nil {
			d.recordCacheError(observability.CacheManifests)
			return false
		}
	}

	contentLength := int64(len(manifest))
	if contentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}
	if file.MediaType != "" {
		w.Header().Set("Content-Type", file.MediaType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	if file.BlobKey != "" {
		w.Header().Set("Docker-Content-Digest", file.BlobKey)
	}
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		d.recordCacheHit(observability.CacheManifests)
		return true
	}
	n, err := w.Write(manifest)
	if err != nil {
		d.recordCacheError(observability.CacheManifests)
		return true
	}
	d.recordCacheBytes(observability.CacheManifests, "serve", int64(n))
	d.recordCacheHit(observability.CacheManifests)
	return true
}

func (d *containerRegistryInstance) upstreamHost() string {
	u, err := url.Parse(d.config.Upstream.URL)
	if err != nil {
		return ""
	}
	return u.Host
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
