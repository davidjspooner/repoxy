package general

import (
	"context"
	"net/http"

	"github.com/davidjspooner/repoxy/pkg/repo"
)

type factory struct {
	muxOnetimeDone bool
}

func init() {
	repo.MustRegisterFactory("general", &factory{})
}

var _ repo.Factory = (*factory)(nil)

func (f *factory) NewRepo(ctx context.Context, config map[string]string) (repo.Instance, error) {
	return &generalInstance{}, nil
}

func (f *factory) addHandlersOnce(mux *http.ServeMux, instance *generalInstance) error {
	if !f.muxOnetimeDone {
		f.muxOnetimeDone = true
	}
	return nil
}
