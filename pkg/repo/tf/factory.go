package tf

import "github.com/davidjspooner/repoxy/pkg/repo"

type factory struct {
	tofo bool
}

func init() {
	repo.MustRegisterFactory("terraform", &factory{tofo: false})
	repo.MustRegisterFactory("tofo", &factory{tofo: true})
}

var _ repo.Factory = (*factory)(nil)

func (f *factory) NewRepo(config map[string]string) (repo.Instance, error) {
	return &tfInstance{tofu: f.tofo}, nil
}
