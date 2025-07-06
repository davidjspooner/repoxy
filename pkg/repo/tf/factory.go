package tf

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// factory implements the repo.Factory interface for Terraform and Tofu providers.
type factory struct {
	muxOnetimeDone bool // Used to ensure the mux is only set up once
}

// init registers the Terraform and Tofu factories.
func init() {
	repo.MustRegisterFactory("terraform|tofo", &factory{})
}

// Ensure factory implements repo.Factory.
var _ repo.Factory = (*factory)(nil)

// NewRepo creates a new Terraform or Tofu repository instance.
func (f *factory) NewRepo(ctx context.Context, config *repo.Config) (repo.Instance, error) {
	instance := &tfInstance{
		config:  *config,
		factory: f,
	}
	return instance, nil
}

// AddToMux registers HTTP handlers for the Terraform/Tofu endpoints on the mux.
func (f *factory) AddToMux(mux *mux.ServeMux) error {
	if f.muxOnetimeDone {
		return nil
	}
	f.muxOnetimeDone = true
	mux.HandleFunc("GET /.well-known/terraform.json", f.HandleWellKnownTerraform)
	mux.HandleFunc("GET /v1/providers/{namespace}/{name}/versions", f.HandleV1VersionList)
	mux.HandleFunc("GET /v1/providers/{namespace}/{name}/{version}", f.HandleV1Version)
	mux.HandleFunc("GET /v1/providers/{namespace}/{name}/{version}/{tail...}", f.HandleV1VersionDownload)
	return nil
}

// HandleWellKnownTerraform handles requests to the .well-known/terraform.json endpoint.
func (f *factory) HandleWellKnownTerraform(w http.ResponseWriter, r *http.Request) {
	//TODO: implement the logic to handle the request for the .well-known/terraform.json
	slog.DebugContext(r.Context(), "TODO: implement the logic to handle the request for the .well-known/terraform.json")

	// 	{
	// 	  "providers.v1": "https://registry.opentofu.org/v1/providers/",
	// 	}
	w.WriteHeader(http.StatusNotImplemented)
}

// lookupParam extracts the provider reference from the request path.
func (f *factory) lookupParam(r *http.Request) (*tfInstance, *param) {
	ref := &param{
		namespace: r.PathValue("namespace"),
		name:      r.PathValue("name"),
		version:   r.PathValue("version"),
		tail:      r.PathValue("tail"),
	}
	slog.DebugContext(r.Context(), "Terraform/Tofu endpoint not found", slog.String("path", r.URL.Path))
	return nil, ref
}

// HandleV1VersionList handles requests for the list of provider versions.
func (f *factory) HandleV1VersionList(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV1VersionList(param, w, r)
}

// HandleV1Version handles requests for a specific provider version.
func (f *factory) HandleV1Version(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV1Version(param, w, r)
}

func (f *factory) HandleV1VersionDownload(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV1VersionDownload(param, w, r)
}

// HandleNotFound handles requests to endpoints that are not found.
func (f *factory) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Repository Not Found", http.StatusNotFound)
}
