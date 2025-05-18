package repo

import (
	"errors"
	"sync"
)

var ErrInvalidRepoType = errors.New("invalid proxy type")
var ErrInvalidRepoConfig = errors.New("invalid proxy config")

type Instance interface {
}

type Factory interface {
	// Create creates a new proxy instance
	NewRepo(config map[string]string) (Instance, error)
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

func NewRepositoryFactory(typeName string, config map[string]string) (Instance, error) {
	factoryLock.RLock()
	defer factoryLock.RUnlock()
	factory, ok := factories[typeName]
	if !ok {
		return nil, ErrInvalidRepoType
	}
	repo, err := factory.NewRepo(config)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
