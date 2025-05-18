package docker

import "github.com/davidjspooner/repoxy/pkg/repo"

type factory struct {
}

func init() {
	repo.MustRegisterFactory("docker", &factory{})
}

var _ repo.Factory = (*factory)(nil)

func (f *factory) NewRepo(config map[string]string) (repo.Instance, error) {
	return &dockerInstance{}, nil
}
