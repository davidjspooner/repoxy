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

type Type interface {
	// Initialize allows a repository type to set up HTTP routes, storage subtrees, etc.
	Initialize(ctx context.Context, typeName string, mux *mux.ServeMux) error
	// NewRepository constructs an Instance for the given logical repository configuration.
	NewRepository(ctx context.Context, common CommonStorage, config *Repo) (Instance, error)
	// Meta returns human-friendly labels for UI display.
	Meta() TypeMeta
}

type TypeDetails struct {
	rType     Type
	ready     bool
	fs        storage.WritableFS // type-level filesystem
	typeName  string
	instances map[string]*InstanceDetails // per-instance state by repo name
}

type InstanceDetails struct {
	name     string
	common   CommonStorage
	instance Instance
}

// TypeMeta captures human-friendly labels for UI use.
type TypeMeta struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
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
			rType:     rType,
			typeName:  typeName,
			instances: map[string]*InstanceDetails{},
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
		rTypeDetail.fs = typeFS
		if err := rTypeDetail.rType.Initialize(ctx, typeName, mux); err != nil {
			return err
		}
		rTypeDetail.ready = true
	}
	return nil
}

func NewRepository(ctx context.Context, config *Repo) (Instance, error) {
	rTypeLock.Lock()
	defer rTypeLock.Unlock()
	rTypeDetail, ok := rTypeDetails[config.Type]
	if !ok {
		return nil, ErrInvalidRepoType
	}
	if !rTypeDetail.ready {
		return nil, storage.Errorf(nil, "", "EINIT", nil).WithMessage("repository type %q not initialized", config.Type)
	}
	repoName := config.Name
	if repoName == "" {
		repoName = "default"
	}
	existing, ok := rTypeDetail.instances[repoName]
	if ok && existing != nil && existing.instance != nil {
		return existing.instance, nil
	}
	typeFS := rTypeDetail.fs
	if typeFS == nil {
		return nil, storage.Errorf(nil, "", "EINIT", nil).WithMessage("repository type %q missing filesystem", config.Type)
	}
	repoFS, err := typeFS.EnsureSub(ctx, repoName)
	if err != nil {
		return nil, err
	}
	common, err := NewCommonStorageWithLabels(repoFS, config.Type, repoName)
	if err != nil {
		return nil, err
	}
	repo, err := rTypeDetail.rType.NewRepository(ctx, common, config)
	if err != nil {
		return nil, err
	}
	rTypeDetail.instances[repoName] = &InstanceDetails{
		name:     repoName,
		common:   common,
		instance: repo,
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
