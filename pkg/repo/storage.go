package repo

import (
	"context"

	"github.com/davidjspooner/go-fs/pkg/storage"
)

type StorageForest struct {
	root storage.WritableFS
}

func NewStorageForest(ctx context.Context, conf *Storage) (*StorageForest, error) {

	readableRS, err := storage.OpenFileSystemFromString(ctx, conf.URL, conf.Config)
	if err != nil {
		return nil, err
	}
	// Check if the ReadOnlyFS also implements WritableFS
	writableRS, ok := readableRS.(storage.WritableFS)
	if !ok {
		return nil, storage.Errorf(readableRS, "", "ECONFIG", nil).WithMessage("storage does not support writing")
	}
	return &StorageForest{
		root: writableRS,
	}, nil
}
