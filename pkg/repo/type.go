package repo

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/go-http-server/pkg/mux"
)

var ErrInvalidRepoType = errors.New("invalid proxy type")
var ErrInvalidRepoConfig = errors.New("invalid proxy config")

type Instance interface {
	GetMatchWeight(name []string) int //weight
}

type Type interface {
	// Initialize allows a repository type to set up HTTP routes, storage subtrees, etc.
	Initialize(ctx context.Context, typeName string, fs storage.WritableFS, mux *mux.ServeMux) error
	// NewRepository constructs an Instance for the given logical repository configuration.
	NewRepository(ctx context.Context, config *Repo) (Instance, error)
}

type TypeDetails struct {
	rType Type
	ready bool
}

var (
	rTypeDetails = make(map[string]*TypeDetails)
	rTypeLock    sync.RWMutex
)

func MustRegisterType(typeNames string, rType Type) {
	rTypeLock.Lock()
	defer rTypeLock.Unlock()
	for _, typeName := range strings.Split(typeNames, "|") {
		if _, ok := rTypeDetails[typeName]; ok {
			panic("repository type already registered")
		}
		rTypeDetails[typeName] = &TypeDetails{
			rType: rType,
		}
	}
}

func Initialize(ctx context.Context, fs storage.WritableFS, mux *mux.ServeMux) error {
	rTypeLock.Lock()
	defer rTypeLock.Unlock()

	for typeName, rTypeDetail := range rTypeDetails {
		typeFS, err := fs.EnsureSub(ctx, "type/"+typeName)
		if err != nil {
			return err
		}
		if err := rTypeDetail.rType.Initialize(ctx, typeName, typeFS, mux); err != nil {
			return err
		}
		rTypeDetail.ready = true
	}
	return nil
}

func NewRepository(ctx context.Context, config *Repo) (Instance, error) {
	rTypeLock.RLock()
	defer rTypeLock.RUnlock()
	rTypeDetail, ok := rTypeDetails[config.Type]
	if !ok {
		return nil, ErrInvalidRepoType
	}
	if !rTypeDetail.ready {
		return nil, storage.Errorf(nil, "", "EINIT", nil).WithMessage("repository type %q not initialized", config.Type)
	}
	repo, err := rTypeDetail.rType.NewRepository(ctx, config)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func NewStorageRoot(ctx context.Context, conf *Storage) (storage.WritableFS, error) {

	readableRS, err := storage.OpenFileSystemFromString(ctx, conf.URL, conf.Config)
	if err != nil {
		return nil, err
	}
	// Check if the ReadOnlyFS also implements WritableFS
	writableFS, ok := readableRS.(storage.WritableFS)
	if !ok {
		return nil, storage.Errorf(readableRS, "", "ECONFIG", nil).WithMessage("storage does not support writing")
	}
	return writableFS, nil
}
