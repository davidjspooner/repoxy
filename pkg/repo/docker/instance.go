package docker

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// dockerInstance implements the repo.Instance interface for Docker repositories.
type dockerInstance struct {
	factory      *factory
	config       repo.Config
	pipeline     client.MiddlewarePipeline
	nameMatchers repo.NameMatchers // Matchers for repository names
}

// NewDockerInstance creates a new Docker repository instance.
// It initializes the instance with the factory and configuration, and sets up the authentication middleware.
func NewDockerInstance(factory *factory, config *repo.Config) (*dockerInstance, error) {
	instance := &dockerInstance{
		factory: factory,
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
	d.proxyToUpstream(r.Context(), r)
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

// HandleV2BlobDigest handles Docker V2 blob digest requests. Returns a 405 for write operations.
func (d *dockerInstance) HandleV2BlobDigest(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	// TODO : Implement the logic to handle blob digest requests
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle blob digest requests")
	w.WriteHeader(http.StatusNotImplemented)
}

func (d *dockerInstance) Authenticate(response *http.Response) string {
	challenge := response.Header.Get("WWW-Authenticate")
	if challenge == "" {
		return ""
	}
	//eg "Bearer realm=\"Bearer realm=\"https://auth.docker.io/token\",service=\"registry.docker.io\",scope=\"repository:library/bash:pull\"\""
	challenges := client.ParseWWWAuthenticate(challenge)
	for _, challenge := range challenges {
		//TODO
		slog.DebugContext(response.Request.Context(), "TODO: Handle parsed WWW-Authenticate challenge")
		_ = challenge
	}
	return ""
}

func (d *dockerInstance) Config() repo.Config {
	return d.config
}

func (d *dockerInstance) proxyToUpstream(ctx context.Context, r *http.Request) (*http.Response, error) {
	u, err := url.Parse(d.config.Upstream.URL)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse upstream URL", "error", err)
		return nil, err
	}
	u.Path = r.URL.Path
	u.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(ctx, r.Method, u.String(), r.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create new request", "error", err)
		return nil, err
	}
	var c client.Interface = &http.Client{}
	c = d.pipeline.WrapClient(c)
	return c.Do(req)
}
