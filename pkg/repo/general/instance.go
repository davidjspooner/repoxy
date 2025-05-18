package general

import "github.com/davidjspooner/repoxy/pkg/repo"

type generalInstance struct {
}

var _ repo.Instance = (*generalInstance)(nil)
