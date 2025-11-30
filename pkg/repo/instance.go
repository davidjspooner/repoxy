package repo

import (
	"errors"
)

var ErrUIAPINotImplemented = errors.New("repository does not implement UI navigation APIs")

type Instance interface {
	// GetMatchWeight reports how well the instance's mappings match the provided artifact path segments.
	GetMatchWeight(name []string) int
	// DescribeRepository returns read-only metadata used by UI clients to label the repository tile.
	Describe() InstanceMeta
}

// InstanceMeta describes a repository instance for UI display.
type InstanceMeta struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	TypeID      string `json:"type_id"`
}
