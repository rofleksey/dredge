package settings

import (
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
)

func New(repo repository.Store, obs *observability.Stack) *Service {
	return &Service{repo: repo, obs: obs}
}

type Service struct {
	repo repository.Store
	obs  *observability.Stack
}
