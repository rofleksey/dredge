package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rofleksey/dredge/internal/repository"
)

type Repository struct {
	pool *pgxpool.Pool
	obs  repository.Instrumentation
}
