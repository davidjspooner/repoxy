package config

// Auth represents authentication details for accessing upstream services or storage.
// It includes the type of authentication, an ID, and a secret.
// The ID/Secret values can be inline values or ${ENV_VAR} references.
type Auth struct {
	Type   string `yaml:"type"`
	ID     string `yaml:"id"`
	Secret string `yaml:"secret"`
}

// Upstream represents an upstream service configuration.
// It includes the URL of the service and optional authentication details.
// examples could be a Git repository URL, a Docker registry URL, etc.
type Upstream struct {
	URL  string `yaml:"url"`
	Auth Auth   `yaml:"auth"`
}

// Mapping represents a mapping between source and destination paths.
// This is used to define how paths in requests to the proxies should be mapped to upstream service paths.
type Mapping struct {
	Source string `yaml:"source"`
	Dest   string `yaml:"dest"`
}

// Repo represents a repository configuration.
// It includes the name of the repository, its type (e.g., git, docker), upstream service details,
// and a list of mappings that define how requests to this repository should be handled.
type Repo struct {
	Name     string    `yaml:"name"`
	Type     string    `yaml:"type"`
	Upstream Upstream  `yaml:"upstream"`
	Mappings []Mapping `yaml:"mappings"`
}

// Storage represents the storage configuration for the proxy.
// It includes the URL of the storage service and optional authentication details.
// This is used for caching immutable artifacts fetched from upstream services.
// Examples could be a file storage service, an object storage service, etc.
type Storage struct {
	URL  string `yaml:"path"`
	Auth Auth   `yaml:"auth"`
}

type Server struct {
	// TODO: Add server configuration options
	// For example, server address, port, TLS settings, etc.
}

// File represents the overall configuration for repoxy
type File struct {
	Storage Storage `yaml:"storage"`
	Repos   []Repo  `yaml:"repos"`
}
