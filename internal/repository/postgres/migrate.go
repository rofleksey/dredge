package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const ensureSchemaMigrations = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

// RunMigrations applies pending SQL files from migrations/ (lexicographic order by filename).
// Applied versions are recorded in schema_migrations.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, ensureSchemaMigrations); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	names, err := listMigrationFiles()
	if err != nil {
		return err
	}

	for _, name := range names {
		applied, err := isMigrationApplied(ctx, pool, name)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		b, err := migrationsFS.ReadFile(path.Join("migrations", name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		sql := strings.TrimSpace(string(b))
		if sql == "" {
			return fmt.Errorf("migration %s is empty", name)
		}

		if err := applyMigration(ctx, pool, name, sql); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}

	return nil
}

func listMigrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var names []string

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		names = append(names, e.Name())
	}

	sort.Strings(names)

	return names, nil
}

func isMigrationApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	var n int

	err := pool.QueryRow(ctx, `SELECT 1 FROM schema_migrations WHERE version = $1`, version).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, version, sql string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Simple query protocol: multiple statements in one round trip, one implicit transaction.
	full := sql + "\nINSERT INTO schema_migrations (version) VALUES (" + quoteSQLString(version) + ");"

	results, err := conn.Conn().PgConn().Exec(ctx, full).ReadAll()
	if err != nil {
		return err
	}

	return firstExecError(results)
}

func firstExecError(results []*pgconn.Result) error {
	for _, r := range results {
		if r.Err != nil {
			return r.Err
		}
	}

	return nil
}

func quoteSQLString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
