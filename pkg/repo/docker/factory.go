package docker

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/davidjspooner/go-http-client/pkg/client"
	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// factory implements the repo.Factory interface for Docker repositories.
type factory struct {
	muxOnetimeDone bool
}

// init registers the Docker factory.
func init() {
	repo.MustRegisterFactory("docker", &factory{})
}

// Ensure factory implements repo.Factory.
var _ repo.Factory = (*factory)(nil)

// NewRepo creates a new Docker repository instance.
func (f *factory) NewRepo(ctx context.Context, config *repo.Config) (repo.Instance, error) {
	instance := &dockerInstance{
		factory: f,
		config:  *config,
	}
	instance.pipeline = append(instance.pipeline, client.WithAuthentication(instance))
	return instance, nil
}

// AddToMux registers HTTP handlers for Docker endpoints on the mux.
func (f *factory) AddToMux(mux *mux.ServeMux) error {
	// API Root
	mux.HandleFunc("GET /v2/", f.HandleV2)

	// Catalog
	mux.HandleFunc("GET /v2/_catalog", f.HandleV2Catalog)

	//tags
	mux.HandleFunc("GET /v2/{name...}/tags/list", f.HandleV2Tags)

	//manifests
	mux.HandleFunc("GET|PUT|DELETE /v2/{name...}/manifests/{tag}", f.HandleV2Manifest) //note tag may also match manifest

	//blobs
	mux.HandleFunc("POST /v2/{name...}/blobs/uploads/", f.HandleV2BlobUpload)
	mux.HandleFunc("PATCH|PUT|DELETE /v2/{name...}/blobs/uploads/{uuid}", f.HandleV2BlobUID)
	mux.HandleFunc("GET|HEAD|DELETE /v2/{name...}/blobs/{digest}", f.HandleV2BlobDigest)
	return nil
}

// lookupParam extracts the Docker repository instance from the request path.
// This is a placeholder implementation.
func (f *factory) lookupParam(r *http.Request) (*dockerInstance, *param) {
	param := &param{
		name:   r.PathValue("name"),
		tag:    r.PathValue("tag"),
		uuid:   r.PathValue("uuid"),
		digest: r.PathValue("digest"),
	}
	return nil, param
}

// HandleV2Catalog handles requests to the Docker v2 catalog endpoint.
func (f *factory) HandleV2Catalog(w http.ResponseWriter, r *http.Request) {
	// TODO : Implement the logic to handle the request for the Docker v2 catalog
	slog.DebugContext(r.Context(), "TODO: Implement the logic to handle the request for the Docker v2 catalog")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("V2 Catalog handler"))
}

// HandleV2 handles requests to the Docker v2 API root endpoint.
func (f *factory) HandleV2(w http.ResponseWriter, r *http.Request) {
	// TODO : Implement the logic to handle the request for the Docker v2 API root
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("V2 handler"))
}

// HandleV2Tags handles requests to the Docker v2 tags endpoint.
func (f *factory) HandleV2Tags(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV2Tags(param, w, r)
}

// HandleV2Manifest handles requests to the Docker v2 manifest endpoint.
func (f *factory) HandleV2Manifest(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV2Manifest(param, w, r)
}

// HandleV2BlobUpload handles requests to the Docker v2 blob upload endpoint.
func (f *factory) HandleV2BlobUpload(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV2BlobUpload(param, w, r)
}

// HandleV2BlobUID handles requests to the Docker v2 blob UID endpoint.
func (f *factory) HandleV2BlobUID(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV2BlobUID(param, w, r)
}

// HandleV2BlobDigest handles requests to the Docker v2 blob digest endpoint.
func (f *factory) HandleV2BlobDigest(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV2BlobDigest(param, w, r)
}

func (f *factory) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	// Handle not found for Docker endpoints
	http.Error(w, "Repository Not Found", http.StatusNotFound)
}
