package docker

import "github.com/davidjspooner/repoxy/pkg/repo"

type dockerInstance struct {
}

var _ repo.Instance = (*dockerInstance)(nil)
