package docker

import (
	"context"
	"encoding/json"
	"fmt"
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
	factory      *factory
	fs           storage.WritableFS
	blobs        *repo.BlobHelper
	config       repo.Repo
	pipeline     client.MiddlewarePipeline
	nameMatchers repo.NameMatchers // Matchers for repository names
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
	// TODO : Implement the logic to handle tags requests
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle tags requests")
	w.WriteHeader(http.StatusNotImplemented)
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
	challenge := response.Header.Get("WWW-Authenticate")
	if challenge == "" {
		return ""
	}
	//eg "Bearer realm=\"Bearer realm=\"https://auth.docker.io/token\",service=\"registry.docker.io\",scope=\"repository:library/bash:pull\"\""
	challenges, err := client.ParseWWWAuthenticate(context.Background(), challenge)
	if err != nil {
		slog.ErrorContext(response.Request.Context(), "Failed to parse WWW-Authenticate challenge", "error", err)
		return ""
	}
	for _, challenge := range challenges {

		if challenge.Scheme != "Bearer" {
			continue
		}
		client := http.Client{}
		client.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/bash:pull")

		//challenge= {Scheme: "Bearer", Params: map[string]string ["realm": "https://auth.docker.io/token", "service": "registry.docker.io", "scope": "repository:library/bash:pull", ]}

		urlStr, ok := challenge.Params["realm"]
		if !ok {
			slog.ErrorContext(response.Request.Context(), "Failed to get realm from WWW-Authenticate challenge")
			return ""
		}
		service, ok := challenge.Params["service"]
		if !ok {
			slog.ErrorContext(response.Request.Context(), "Failed to get service from WWW-Authenticate challenge")
			return ""
		}
		scope, ok := challenge.Params["scope"]
		if !ok {
			slog.ErrorContext(response.Request.Context(), "Failed to get scope from WWW-Authenticate challenge")
			return ""
		}
		urlStr += fmt.Sprintf("?service=%s&scope=%s", service, scope)
		r, err := http.Get(urlStr)
		if err != nil {
			slog.ErrorContext(response.Request.Context(), "Failed to get token", "error", err)
			return ""
		}
		defer r.Body.Close()
		if r.StatusCode != http.StatusOK {
			slog.ErrorContext(response.Request.Context(), "Failed to get token", "status", r.Status)
			return ""
		}
		var tokenResp struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&tokenResp); err != nil {
			slog.ErrorContext(response.Request.Context(), "Failed to decode token response", "error", err)
			return ""
		}
		return "Bearer " + tokenResp.Token
	}
	return ""
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
	var c client.Interface = &http.Client{}
	c = d.pipeline.WrapClient(c)
	resp, err := c.Do(req)
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
