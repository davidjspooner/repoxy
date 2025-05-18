package tf

import "github.com/davidjspooner/repoxy/pkg/repo"

type tfInstance struct {
	tofu bool
}

var _ repo.Instance = (*tfInstance)(nil)
