package tf

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"path"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// tfType implements the repo.Type interface for Terraform and Tofu providers.
type tfType struct {
	muxOnetimeDone bool          // Used to ensure the mux is only set up once
	instances      []*tfInstance // List of registered Terraform/Tofu instances
	fs             storage.WritableFS
}

// init registers the Terraform and Tofu factories.
func init() {
	repo.MustRegisterType("terraform|tofo", &tfType{})
}

// Ensure factory implements repo.Type.
var _ repo.Type = (*tfType)(nil)

// NewRepository creates a new Terraform or Tofu repository instance.
func (f *tfType) NewRepository(ctx context.Context, config *repo.Repo) (repo.Instance, error) {
	if f.fs == nil {
		return nil, errors.New("tf type not initialized")
	}
	proxyFS, err := f.fs.EnsureSub(ctx, path.Join("proxies", config.Name))
	if err != nil {
		return nil, err
	}
	refsFS, err := proxyFS.EnsureSub(ctx, "refs")
	if err != nil {
		return nil, err
	}
	refsHelper, err := repo.NewStorageHelper(refsFS)
	if err != nil {
		return nil, err
	}
	instance, err := NewInstance(config, proxyFS, refsHelper)
	if err != nil {
		return nil, err
	}
	f.instances = append(f.instances, instance)
	if config.Type == "tofo" {
		instance.tofu = true // Set tofu flag for Tofu instances
	} else {
		instance.tofu = false // Set tofu flag for Terraform instances
	}
	return instance, nil
}

// Initialize registers HTTP handlers for the Terraform/Tofu endpoints on the mux.
func (f *tfType) Initialize(ctx context.Context, typeName string, fs storage.WritableFS, mux *mux.ServeMux) error {
	f.fs = fs
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
func (f *tfType) HandleWellKnownTerraform(w http.ResponseWriter, r *http.Request) {
	//TODO: implement the logic to handle the request for the .well-known/terraform.json
	slog.DebugContext(r.Context(), "TODO: implement the logic to handle the request for the .well-known/terraform.json")

	// 	{
	// 	  "providers.v1": "https://registry.opentofu.org/v1/providers/",
	// 	}
	w.WriteHeader(http.StatusNotImplemented)
}

// lookupParam extracts the provider reference from the request path.
func (f *tfType) lookupParam(r *http.Request) (*tfInstance, *param) {
	ref := &param{
		namespace: r.PathValue("namespace"),
		name:      r.PathValue("name"),
		version:   r.PathValue("version"),
		tail:      r.PathValue("tail"),
	}
	var bestInstance *tfInstance
	var bestScore int
	nameParts := []string{ref.namespace, ref.name}
	for _, instance := range f.instances {
		score := instance.GetMatchWeight(nameParts)
		if score > bestScore {
			bestScore = score
			bestInstance = instance
		}
	}
	return bestInstance, ref
}

// HandleV1VersionList handles requests for the list of provider versions.
func (f *tfType) HandleV1VersionList(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV1VersionList(param, w, r)
}

// HandleV1Version handles requests for a specific provider version.
func (f *tfType) HandleV1Version(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV1Version(param, w, r)
}

func (f *tfType) HandleV1VersionDownload(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV1VersionDownload(param, w, r)
}

// HandleNotFound handles requests to endpoints that are not found.
func (f *tfType) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Repository Not Found", http.StatusNotFound)
}
