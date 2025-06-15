package gitrelease

import "fmt"

type ReleaseSource interface {
	ListRepos(prefix string) ([]string, error)
	ListVersions(repo string) ([]string, error)
	ListFiles(repo, version string) ([]string, error)
	GetFile(repo, version, file string) ([]byte, error)
}

type Authenticator interface {
	GetAuthToken() (string, error)
}

type Factory interface {
	NewReleaseSource() (ReleaseSource, error)
	NewAuthenticator() (Authenticator, error)
}

var factories = make(map[string]Factory)

func RegisterFactory(fType string, factory Factory) error {
	if factory == nil {
		fmt.Errorf("gitrelease: RegisterFactory called with nil factory for %s", fType)
	}
	if _, ok := factories[fType]; ok {
		return fmt.Errorf("gitrelease: RegisterFactory called twice for %s", fType)
	}
	factories[fType] = factory
	return nil
}

func MustRegisterFactory(fType string, factory Factory) {
	if err := RegisterFactory(fType, factory); err != nil {
		panic(err)
	}
}

func GetFactory(fType string) (Factory, error) {
	if factory, ok := factories[fType]; ok {
		return factory, nil
	}
	return nil, fmt.Errorf("gitrelease: unknown factory %s", fType)
}
