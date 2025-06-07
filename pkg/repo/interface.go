package repo

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/repoxy/internal/config"
)

var ErrInvalidRepoType = errors.New("invalid proxy type")
var ErrInvalidRepoConfig = errors.New("invalid proxy config")

type Instance interface {
	AddHandlersToMux(mux *http.ServeMux) error
	SetStorage(storage storage.ReadOnlyFS) error
}

type Factory interface {
	// Create creates a new proxy instance
	NewRepo(ctx context.Context, config config.Repo) (Instance, error)
}

var factories = make(map[string]Factory)
var factoryLock sync.RWMutex

func MustRegisterFactory(typeName string, factory Factory) {
	factoryLock.Lock()
	defer factoryLock.Unlock()
	if _, ok := factories[typeName]; ok {
		panic("factory already registered")
	}
	factories[typeName] = factory
}

func NewRepositoryFactory(ctx context.Context, config config.Repo) (Instance, error) {
	factoryLock.RLock()
	defer factoryLock.RUnlock()
	factory, ok := factories[config.Type]
	if !ok {
		return nil, ErrInvalidRepoType
	}
	repo, err := factory.NewRepo(ctx, config)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
