package docker

import (
	"fmt"
	"net/http"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

type dockerInstance struct {
	factory *factory
}

var _ repo.Instance = (*dockerInstance)(nil)

func (d *dockerInstance) AddHandlersToMux(mux *http.ServeMux) error {
	d.factory.addHandlersOnce(mux, d)
	return nil
}

func (d *dockerInstance) Handle(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// For example, you can write a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Docker instance handler"))
}

func (d *dockerInstance) SetStorage(fs storage.ReadOnlyFS) error {
	return fmt.Errorf("SetStorage not implemented")
}

func (d *dockerInstance) HandleV2Tags(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// For example, you can write a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V2 Tags handler"))
}
func (d *dockerInstance) HandleV2Manifest(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// For example, you can write a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V2 Tags handler"))
}
func (d *dockerInstance) HandleV2Blobs(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// For example, you can write a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("V2 Tags handler"))
}
