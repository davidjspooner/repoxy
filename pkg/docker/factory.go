package docker

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

// factory implements the repo.Factory interface for Docker repositories.
type factory struct {
	instances []*dockerInstance
}

// init registers the Docker factory.
func init() {
	repo.MustRegisterFactory("docker", &factory{})
}

// Ensure factory implements repo.Factory.
var _ repo.Factory = (*factory)(nil)

// NewRepo creates a new Docker repository instance.
func (f *factory) NewRepo(ctx context.Context, config *repo.Repo) (repo.Instance, error) {
	instance, err := NewDockerInstance(f, config)
	if err != nil {
		return nil, err
	}
	f.instances = append(f.instances, instance)
	return instance, nil
}

// AddToMux registers HTTP handlers for Docker endpoints on the mux.
func (f *factory) AddToMux(mux *mux.ServeMux) error {
	// API Root
	mux.HandleFunc("GET /v2/", f.HandleV2)

	// Catalog
	mux.HandleFunc("GET /v2/_catalog", f.HandleV2Catalog)

	//tags
	mux.HandleFunc("GET /v2/{name...}/tags/list", f.HandleV2Tags) //auto HEAD

	//manifests
	mux.HandleFunc("GET|PUT|DELETE /v2/{name...}/manifests/{tag}", f.HandleV2Manifest) //note tag may also match manifest

	//blobs
	mux.HandleFunc("POST /v2/{name...}/blobs/uploads/", f.HandleV2BlobUpload)
	mux.HandleFunc("PATCH|PUT|DELETE /v2/{name...}/blobs/uploads/{uuid}", f.HandleV2BlobUID)
	mux.HandleFunc("GET|DELETE /v2/{name...}/blobs/{digest}", f.HandleV2BlobByDigest) //auto HEAD
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
	var bestInstance *dockerInstance
	var bestScore int
	nameParts := strings.Split(param.name, "/")
	for _, instance := range f.instances {
		score := instance.GetMatchWeight(nameParts)
		if score > bestScore {
			bestScore = score
			bestInstance = instance
		}
	}
	return bestInstance, param
}

// HandleV2Catalog handles requests to the Docker v2 catalog endpoint.
func (f *factory) HandleV2Catalog(w http.ResponseWriter, r *http.Request) {
	// TODO : Implement the logic to handle the request for the Docker v2 catalog
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("V2 Catalog handler"))
}

// HandleV2 handles requests to the Docker v2 API root endpoint.
func (f *factory) HandleV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	w.Header().Set("Content-Length", "2")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Date", time.Now().Format(http.TimeFormat))

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("{}"))

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

// HandleV2BlobByDigest handles requests to the Docker v2 blob digest endpoint.
func (f *factory) HandleV2BlobByDigest(w http.ResponseWriter, r *http.Request) {
	instance, param := f.lookupParam(r)
	if instance == nil {
		f.HandleNotFound(w, r)
		return
	}
	instance.HandleV2BlobByDigest(param, w, r)
}

func (f *factory) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	// Handle not found for Docker endpoints
	http.Error(w, "Repository Not Found", http.StatusNotFound)
}
