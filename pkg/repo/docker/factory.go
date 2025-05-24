package docker

import (
	"context"
	"net/http"
	"strings"

	"github.com/davidjspooner/repoxy/pkg/repo"
)

type factory struct {
	muxOnetimeDone bool
}

func init() {
	repo.MustRegisterFactory("docker", &factory{})
}

var _ repo.Factory = (*factory)(nil)

func (f *factory) NewRepo(ctx context.Context, config map[string]string) (repo.Instance, error) {
	return &dockerInstance{}, nil
}

func (f *factory) addHandlersOnce(mux *http.ServeMux, instance *dockerInstance) error {
	if !f.muxOnetimeDone {
		f.muxOnetimeDone = true
		mux.HandleFunc("GET /v2/$", f.HandleV2)
		mux.HandleFunc("GET /v2/_catalog$", f.HandleV2Catalog)
		mux.HandleFunc("GET /v2/<name>/<tail...>$", f.HandleV2Other)
	}
	return nil
}
func (f *factory) HandleV2Catalog(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// For example, you can write a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V2 Catalog handler"))
}
func (f *factory) HandleV2(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// For example, you can write a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V2 handler"))
}

func lastIndex(s []string, sep string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return i
		}
	}
	return -1
}
func (f *factory) HandleV2Other(w http.ResponseWriter, r *http.Request) {

	parts := []string{r.PathValue("name")}
	parts = append(parts, strings.Split(r.PathValue("tail"), "/")...)

	last := -1
	for _, key := range []string{"tags", "manifests", "blobs"} {
		last = max(last, lastIndex(parts, key))
	}
	if last == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	name := parts[:last]
	tail := parts[last:]

	r.SetPathValue("name", strings.Join(name, "/"))
	r.SetPathValue("tail", strings.Join(tail, "/"))
	var instance *dockerInstance
	if instance == nil {
		return
	}
	switch tail[0] {
	case "tags":
		instance.HandleV2Tags(w, r)
	case "manifests":
		instance.HandleV2Manifest(w, r)
	case "blobs":
		instance.HandleV2Blobs(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

}
