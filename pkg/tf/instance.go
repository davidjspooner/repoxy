package tf

import (
	"log/slog"
	"net/http"

	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

type tfInstance struct {
	tofu         bool
	config       repo.Repo
	pipeline     client.MiddlewarePipeline
	nameMatchers repo.NameMatchers // Matchers for repository names
}

var _ repo.Instance = (*tfInstance)(nil)
var _ client.Authenticator = (*tfInstance)(nil)

func NewInstance(config *repo.Repo) (*tfInstance, error) {
	instance := &tfInstance{
		config: *config,
	}
	instance.nameMatchers.Set(config.Mappings)
	instance.pipeline = append(instance.pipeline, client.WithAuthentication(instance))
	return instance, nil
}

func (d *tfInstance) GetMatchWeight(name []string) int {
	return d.nameMatchers.GetMatchWeight(name)
}

func (d *tfInstance) HandleV1VersionList(param *param, w http.ResponseWriter, r *http.Request) {
	// TODO: Implement the logic to handle the request for a list of provider versions
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle the request for a list of provider versions")
	w.WriteHeader(http.StatusNotImplemented)
}
func (d *tfInstance) HandleV1Version(param *param, w http.ResponseWriter, r *http.Request) {
	// TODO: Implement the logic to handle the request for a specific provider version
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle the request for a specific provider version")
	w.WriteHeader(http.StatusNotImplemented)
}

// for actual download of the provider must return a redirect because the API is written that way
func (d *tfInstance) HandleV1VersionDownload(param *param, w http.ResponseWriter, r *http.Request) {
	// TODO Implement the logic to handle the request for downloading a specific provider version
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle the request for downloading a specific provider version")
	w.WriteHeader(http.StatusNotImplemented)
}

func (d *tfInstance) Authenticate(response *http.Response) string {
	// TODO : Implement the logic to authenticate the request according to selected upstream
	slog.DebugContext(response.Request.Context(), "TODO: Implement the logic to authenticate the request according to selected upstream")
	return ""
}
