package tf

import (
	"net/http"

	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

type tfInstance struct {
	tofu     bool
	factory  *factory
	pipeline client.MiddlewarePipeline
}

var _ repo.Instance = (*tfInstance)(nil)
var _ client.Authenticator = (*tfInstance)(nil)

func (d *tfInstance) AddHandlersToMux(mux *http.ServeMux) error {
	d.factory.addHandlersOnce(mux)
	return nil
}

func (d *tfInstance) HandleV1VersionList(param *param, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func (d *tfInstance) HandleV1Version(param *param, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// for actual download of the provider must return a redirect because the API is written that way
func (d *tfInstance) HandleV1VersionDownload(param *param, w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	w.WriteHeader(http.StatusNotImplemented)
}

func (d *tfInstance) Authenticate(response *http.Response) string {
	// TODO Implement authentication logic here
	return ""
}
