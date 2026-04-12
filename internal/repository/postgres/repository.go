package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rofleksey/dredge/internal/repository"
)

func New(pool *pgxpool.Pool, obs repository.Instrumentation) *Repository {
	return &Repository{pool: pool, obs: obs}
}

var _ repository.Store = (*Repository)(nil)
