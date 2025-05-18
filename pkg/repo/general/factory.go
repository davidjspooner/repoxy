package general

import "github.com/davidjspooner/repoxy/pkg/repo"

type factory struct {
}

func init() {
	repo.MustRegisterFactory("general", &factory{})
}

var _ repo.Factory = (*factory)(nil)

func (f *factory) NewRepo(config map[string]string) (repo.Instance, error) {
	return &generalInstance{}, nil
}
