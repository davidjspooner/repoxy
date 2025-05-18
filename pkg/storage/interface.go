package storage

import (
	"fmt"
	"io/fs"
	"sync"
)

type FileMetaData interface {
	fs.FileInfo
	Metadata(key string) string
}

type WritableFileMetaData interface {
	fs.FileInfo
	SetMetadata(key, value string)
	Clone() FileMetaData
	GetReadOnlyWrapper() FileMetaData
}

type ReadOnlyFS interface {
	fs.FS
	fs.GlobFS
	ReadDir(name string) ([]FileMetaData, error)
	Stat(name string) (FileMetaData, error)
	Sub(name string) (ReadOnlyFS, error)
}

type WritableFS interface {
	ReadOnlyFS
	CreateFrom(path string, r fs.File, fileMetaData FileMetaData) error
	Delete(path string) error
}

type StringMap map[string]string

type Factory interface {
	OpenFileSystem(typ string, config StringMap) (ReadOnlyFS, error)
}
type FactoryFunc func(typ string, config StringMap) (ReadOnlyFS, error)

func (f FactoryFunc) OpenFileSystem(typ string, config StringMap) (ReadOnlyFS, error) {
	return f(typ, config)
}

var factoryMap = make(map[string]Factory)
var factoryLock = sync.RWMutex{}

func MustRegisterFactoryFunc(typ string, factory FactoryFunc) {
	MustRegisterFactory(typ, factory)
}

func MustRegisterFactory(typ string, factory Factory) {
	factoryLock.Lock()
	defer factoryLock.Unlock()
	if _, ok := factoryMap[typ]; ok {
		panic("factory already registered")
	}
	factoryMap[typ] = factory
}

func OpenFileSystem(typ string, config StringMap) (ReadOnlyFS, error) {
	factoryLock.RLock()
	defer factoryLock.RUnlock()
	factory, ok := factoryMap[typ]
	if !ok {
		return nil, fmt.Errorf("invalid filesystem type: %s", typ)
	}
	return factory.OpenFileSystem(typ, config)
}

func init() {
}
