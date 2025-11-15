package docker

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// dockerInstance implements the repo.Instance interface for Docker repositories.
type dockerInstance struct {
	factory           *factory
	fs                storage.WritableFS
	blobs             *repo.BlobHelper
	config            repo.Repo
	pipeline          client.MiddlewarePipeline
	nameMatchers      repo.NameMatchers // Matchers for repository names
	httpClientFactory func() client.Interface
	tokenHTTP         httpDoer
	auth              *dockerUpstreamAuth
}

// NewDockerInstance creates a new Docker repository instance.
// It initializes the instance with the factory and configuration, and sets up the authentication middleware.
func NewDockerInstance(factory *factory, fs storage.WritableFS, blobs *repo.BlobHelper, config *repo.Repo) (*dockerInstance, error) {
	instance := &dockerInstance{
		factory: factory,
		fs:      fs,
		blobs:   blobs,
		config:  *config,
	}
	instance.nameMatchers.Set(config.Mappings)
	instance.pipeline = append(instance.pipeline, client.WithAuthentication(instance))
	instance.httpClientFactory = func() client.Interface {
		return &http.Client{}
	}
	instance.tokenHTTP = &http.Client{}
	auth, err := newDockerUpstreamAuth(instance.tokenHTTP, config.Upstream)
	if err != nil {
		return nil, err
	}
	instance.auth = auth
	return instance, nil
}

// Ensure dockerInstance implements the repo.Instance , client.Authenticator, client.Cache interfaces.
var _ repo.Instance = (*dockerInstance)(nil)
var _ client.Authenticator = (*dockerInstance)(nil)

func (d *dockerInstance) GetMatchWeight(name []string) int {
	return d.nameMatchers.GetMatchWeight(name)
}

// HandledWriteMethodForReadOnlyRepo checks if the request is a write operation and returns a 405 if so.
// Returns true if the request was handled (i.e., is not allowed), false otherwise.
func (d *dockerInstance) HandledWriteMethodForReadOnlyRepo(w http.ResponseWriter, r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		return false // Read operations are allowed
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("This repo is read-only, no write operations allowed"))
		return true
	}
}

// HandleV2Tags handles Docker V2 tags requests. Returns a 405 for write operations.
func (d *dockerInstance) HandleV2Tags(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	if err := d.proxyToUpstream(r.Context(), w, r); err != nil {
		slog.ErrorContext(r.Context(), "failed to proxy docker tags request", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

// HandleV2Manifest handles Docker V2 manifest requests. Returns a 405 for write operations.
func (d *dockerInstance) HandleV2Manifest(param *param, w http.ResponseWriter, r *http.Request) {
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

// HandleV2BlobUpload handles Docker V2 blob upload requests. Returns a 405 for write operations.
func (d *dockerInstance) HandleV2BlobUpload(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	// TODO : Implement the logic to handle blob upload requests
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle blob upload requests")
	w.WriteHeader(http.StatusNotImplemented)
}

// HandleV2BlobUID handles Docker V2 blob UID requests. Returns a 405 for write operations.
func (d *dockerInstance) HandleV2BlobUID(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	// TODO : Implement the logic to handle blob UID requests
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle blob UID requests")
	w.WriteHeader(http.StatusNotImplemented)
}

// HandleV2BlobByDigest handles Docker V2 blob digest requests. Returns a 405 for write operations.
func (d *dockerInstance) HandleV2BlobByDigest(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	if param == nil || param.digest == "" || d.blobs == nil {
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

func (d *dockerInstance) Authenticate(response *http.Response) string {
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

func (d *dockerInstance) Config() repo.Repo {
	return d.config
}

func (d *dockerInstance) proxyToUpstream(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

func (d *dockerInstance) roundTripUpstream(ctx context.Context, r *http.Request) (*http.Response, error) {
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
	httpClient := d.httpClientFactory
	var base client.Interface
	if httpClient != nil {
		base = httpClient()
	} else {
		base = &http.Client{}
	}
	base = d.pipeline.WrapClient(base)
	resp, err := base.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *dockerInstance) writeHeadersFromResponse(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
}

func (d *dockerInstance) serveLocalBlob(param *param, w http.ResponseWriter, r *http.Request) bool {
	reader, err := d.blobs.Open(r.Context(), param.digest)
	if err != nil {
		return false
	}
	defer reader.Close()
	if info, err := d.blobs.Stat(r.Context(), param.digest); err == nil {
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	}
	w.Header().Set("Docker-Content-Digest", param.digest)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return true
	}
	if _, err := io.Copy(w, reader); err != nil {
		slog.ErrorContext(r.Context(), "failed to stream cached blob", "error", err)
	}
	return true
}

func (d *dockerInstance) fetchAndStoreBlob(param *param, w http.ResponseWriter, r *http.Request) error {
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

	tee := io.TeeReader(resp.Body, w)
	_, err = d.blobs.Store(ctx, param.digest, tee)
	return err
}
