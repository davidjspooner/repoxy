package local

import (
	"io/fs"

	"github.com/davidjspooner/repoxy/pkg/storage"
)

type localFS struct {
	//TODO: implement LocalFS
}

func newLocalFS(typ string, config storage.StringMap) (storage.ReadOnlyFS, error) {
	return nil, fs.ErrInvalid // TODO: implement
}

func init() {
	storage.MustRegisterFactoryFunc("local", func(typ string, config storage.StringMap) (storage.ReadOnlyFS, error) {
		return localFS{}, nil
	})
}

var _ storage.WritableFS = (*localFS)(nil)
