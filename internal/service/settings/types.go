package settings

import (
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
)

type Service struct {
	repo repository.Store
	obs  *observability.Stack
}
