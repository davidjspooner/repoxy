package tf

import (
	"fmt"
	"net/http"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/repoxy/pkg/repo"
)

type tfInstance struct {
	tofu    bool
	factory *factory
}

var _ repo.Instance = (*tfInstance)(nil)

func (d *tfInstance) AddHandlersToMux(mux *http.ServeMux) error {
	d.factory.addHandlersOnce(mux, d)
	return nil
}

func (d *tfInstance) Handle(w http.ResponseWriter, r *http.Request) {
	// Handle the request here
	// For example, you can write a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("TF instance handler"))
}

func (d *tfInstance) SetStorage(fs storage.ReadOnlyFS) error {
	return fmt.Errorf("SetStorage not implemented")
}
