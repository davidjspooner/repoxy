package repo

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/davidjspooner/go-http-server/pkg/mux"
)

var ErrInvalidRepoType = errors.New("invalid proxy type")
var ErrInvalidRepoConfig = errors.New("invalid proxy config")

type Instance interface {
	Config() Config
}

type Factory interface {
	// Create creates a new proxy instance
	NewRepo(ctx context.Context, config *Config) (Instance, error)
	AddToMux(mux *mux.ServeMux) error
}

var factories = make(map[string]Factory)
var factoryLock sync.RWMutex

func MustRegisterFactory(typeNames string, factory Factory) {
	factoryLock.Lock()
	defer factoryLock.Unlock()
	for _, typeName := range strings.Split(typeNames, "|") {
		if _, ok := factories[typeName]; ok {
			panic("factory already registered")
		}
		factories[typeName] = factory
	}
}

func AddAllToMux(mux *mux.ServeMux) error {
	factoryLock.RLock()
	defer factoryLock.RUnlock()
	for _, factory := range factories {
		if err := factory.AddToMux(mux); err != nil {
			return err
		}
	}
	return nil
}

func NewRepository(ctx context.Context, config *Config) (Instance, error) {
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
