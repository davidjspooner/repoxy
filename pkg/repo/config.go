package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidjspooner/go-fs/pkg/storage"
	"github.com/davidjspooner/repoxy/pkg/listener"
	"gopkg.in/yaml.v3"
)

// Upstream represents an upstream service configuration.
// It includes the type, URL of the service and optional authentication details.
// examples could be a Git repository URL, a Docker registry URL, etc.
type Upstream struct {
	// URL is the URL of the upstream service . eg https://index.docker.io/
	URL    string            `yaml:"url"`
	Config map[string]string `yaml:"config"`
}

// Repo represents a repository configuration.
// It includes the name of the repository, its type (e.g., git, docker), upstream service details,
// and a list of mappings that define how requests to this repository should be handled.
type Repo struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Upstream Upstream `yaml:"upstream"`
	Mappings []string `yaml:"mappings"`
}

// Storage represents the storage configuration for the proxy.
// It includes the URL of the storage service and optional authentication details.
// This is used for caching immutable artifacts fetched from upstream services.
// Examples could be a file storage service, an object storage service, etc.
type Storage struct {
	URL    string         `yaml:"url"`
	Config storage.Config `yaml:"config"`
}

// File represents the overall configuration for repoxy
type File struct {
	Server       *listener.Group `yaml:"server"`
	Storage      *Storage        `yaml:"storage"`
	Repositories []*Repo         `yaml:"repos"`
}

func loadConfig(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", filename, err)
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	d.KnownFields(true) // Ensure unknown fields are not allowed
	cfg := &File{}
	if err := d.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file %s: %w", filename, err)
	}
	return cfg, nil
}

// LoadConfigs loads the configuration from a set of YAML file.
func LoadConfigs(fileglobs ...string) (*File, error) {
	mergedConfig := &File{}
	for _, fileglob := range fileglobs {
		if fileglob == "" {
			continue
		}
		filenames, err := filepath.Glob(fileglob)
		if err != nil {
			return nil, fmt.Errorf("failed to glob config files %s: %w", fileglob, err)
		}
		for _, filename := range filenames {
			if filename == "" {
				continue
			}
			cfg, err := loadConfig(filename)
			if err != nil {
				return nil, err
			}
			if cfg == nil {
				continue // skip empty configs
			}
			if cfg.Server != nil {
				if mergedConfig.Server != nil {
					return nil, fmt.Errorf("multiple server configurations found")
				}
				mergedConfig.Server = cfg.Server // shallow copy
			}
			if cfg.Storage != nil {
				if mergedConfig.Storage != nil {
					return nil, fmt.Errorf("multiple storage configurations found")
				}
				mergedConfig.Storage = cfg.Storage // shallow copy
			}
			mergedConfig.Repositories = append(mergedConfig.Repositories, cfg.Repositories...)
		}
	}
	if len(mergedConfig.Repositories) == 0 {
		return nil, fmt.Errorf("no repositories found in configuration files")
	}
	if mergedConfig.Server == nil {
		return nil, fmt.Errorf("no server configuration found in configuration files")
	}
	if mergedConfig.Storage == nil {
		return nil, fmt.Errorf("no storage configuration found in configuration files	")
	}
	return mergedConfig, nil
}
