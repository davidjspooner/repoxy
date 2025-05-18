package s3bucket

import (
	"io/fs"

	"github.com/davidjspooner/repoxy/pkg/storage"
)

type s3BucketFS struct {
	//TODO: implement s3bucketFS
}

func news3bucketFS(typ string, config storage.StringMap) (storage.ReadOnlyFS, error) {
	return nil, fs.ErrInvalid // TODO: implement
}

func init() {
	fn := func(typ string, config storage.StringMap) (storage.ReadOnlyFS, error) {
		return s3BucketFS{}, nil
	}
	storage.MustRegisterFactoryFunc("s3", fn)
	storage.MustRegisterFactoryFunc("bucket", fn)
}

var _ storage.WritableFS = (*s3BucketFS)(nil)
