package docker

import (
	"log/slog"
	"net/http"

	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// dockerInstance implements the repo.Instance interface for Docker repositories.
type dockerInstance struct {
	factory  *factory
	config   repo.Config
	pipeline client.MiddlewarePipeline
}

// Ensure dockerInstance implements the repo.Instance , client.Authenticator, client.Cache interfaces.
var _ repo.Instance = (*dockerInstance)(nil)
var _ client.Authenticator = (*dockerInstance)(nil)

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
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// HandleV2Manifest handles Docker V2 manifest requests. Returns a 405 for write operations.
func (d *dockerInstance) HandleV2Manifest(param *param, w http.ResponseWriter, r *http.Request) {
	if d.HandledWriteMethodForReadOnlyRepo(w, r) {
		return
	}
	// TODO : Implement the logic to handle manifest requests
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle manifest requests")
	w.WriteHeader(http.StatusNotImplemented)
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
	challenges := client.ParseWWWAuthenticate(challenge)
	for _, challenge := range challenges {
		//TODO
		slog.DebugContext(response.Request.Context(), "TODO: Handle parsed WWW-Authenticate challenge")
		_ = challenge
	}
	return ""
}

func (d *dockerInstance) Client() client.Interface {
	c := &http.Client{}
	return d.pipeline.WrapClient(c)
}

func (d *dockerInstance) Config() repo.Config {
	return d.config
}
