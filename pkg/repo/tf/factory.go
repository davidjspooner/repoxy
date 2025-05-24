package tf

import (
	"context"
	"net/http"

	"github.com/davidjspooner/repoxy/pkg/repo"
)

type factory struct {
	muxOnetimeDone bool
	tofo           bool
}

func init() {
	repo.MustRegisterFactory("terraform", &factory{tofo: false})
	repo.MustRegisterFactory("tofo", &factory{tofo: true})
}

var _ repo.Factory = (*factory)(nil)

func (f *factory) NewRepo(ctx context.Context, config map[string]string) (repo.Instance, error) {
	return &tfInstance{tofu: f.tofo}, nil
}

func (f *factory) addHandlersOnce(mux *http.ServeMux, instance *tfInstance) error {
	if !f.muxOnetimeDone {
		f.muxOnetimeDone = true
		mux.HandleFunc(".well-known/terraform.json", f.HandleWellKnownTerraform)
		mux.HandleFunc("GET /v1/providers/<namespace>/<name>/versions", f.HandleV1VersionList)
		mux.HandleFunc("GET /v1/providers/<namespace>/<name>/<version>", f.HandleV1Version)
		mux.HandleFunc("GET /v1/providers/<namespace>/<name>/<version>/<tail...>", f.HandleV1VersionDownload)
	}
	return nil
}
func (f *factory) HandleWellKnownTerraform(w http.ResponseWriter, r *http.Request) {
	// 	{
	// 	  "providers.v1": "https://registry.opentofu.org/v1/providers/",
	// 	}
	w.WriteHeader(http.StatusNotImplemented)
}

func (f *factory) HandleV1VersionList(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// /v1/providers/<namespace>/<name>/versions                              [GET]
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")
	_ = namespace
	_ = name
	w.WriteHeader(http.StatusNotImplemented)
}
func (f *factory) HandleV1Version(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// /v1/providers/<namespace>/<name>/<version>                             [GET]
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")
	version := r.PathValue("version")
	_ = namespace
	_ = name
	_ = version
	w.WriteHeader(http.StatusNotImplemented)
}
func (f *factory) HandleV1VersionDownload(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// /v1/providers/<namespace>/<name>/<version>/download/<os>/<arch>        [GET]
	// /v1/providers/<namespace>/<name>/<version>/download                    [GET]
	// /v1/providers/<namespace>/<name>/<version>/shasums                     [GET]
	// /v1/providers/<namespace>/<name>/<version>/shasums.sig                 [GET]
	// /v1/providers/<namespace>/<name>/<version>/manifest.json               [GET]
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")
	version := r.PathValue("version")
	tail := r.PathValue("tail")
	_ = namespace
	_ = name
	_ = version
	_ = tail
	w.WriteHeader(http.StatusNotImplemented)
}
