package settings

import (
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
)

func New(repo repository.Store, obs *observability.Stack) *Usecase {
	return &Usecase{repo: repo, obs: obs}
}

type Usecase struct {
	repo repository.Store
	obs  *observability.Stack
}
