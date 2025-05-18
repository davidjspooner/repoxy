package github

import (
	"github.com/davidjspooner/repoxy/pkg/storage"
)

type gitHubFS struct {
	//TODO: implement GitHubFS
}

var _ storage.ReadOnlyFS = (*gitHubFS)(nil)

func init() {
	fn := func(typ string, config storage.StringMap) (storage.ReadOnlyFS, error) {
		return gitHubFS{}, nil
	}
	storage.MustRegisterFactoryFunc("github", fn)
}
