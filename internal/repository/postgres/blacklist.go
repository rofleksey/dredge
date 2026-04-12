package postgres

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

// ListChannelBlacklist returns all blacklisted channel logins (lowercase).
func (r *Repository) ListChannelBlacklist(ctx context.Context) ([]string, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_channel_blacklist")
	defer span.End()

	rows, err := r.pool.Query(ctx, `SELECT login FROM channel_blacklist ORDER BY login ASC`)
	if err != nil {
		r.obs.LogError(ctx, span, "list blacklist failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []string

	for rows.Next() {
		var login string
		if err := rows.Scan(&login); err != nil {
			return nil, err
		}

		out = append(out, login)
	}

	return out, rows.Err()
}

// AddChannelBlacklist inserts a login (normalized lowercase).
func (r *Repository) AddChannelBlacklist(ctx context.Context, login string) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.add_channel_blacklist")
	defer span.End()

	login = normalizeStoredUsername(login)
	if login == "" {
		return errors.New("empty login")
	}

	_, err := r.pool.Exec(ctx, `INSERT INTO channel_blacklist (login) VALUES ($1) ON CONFLICT (login) DO NOTHING`, login)
	if err != nil {
		r.obs.LogError(ctx, span, "add blacklist failed", err, zap.String("login", login))
	}

	return err
}

// RemoveChannelBlacklist deletes a blacklist row.
func (r *Repository) RemoveChannelBlacklist(ctx context.Context, login string) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.remove_channel_blacklist")
	defer span.End()

	login = normalizeStoredUsername(login)
	if login == "" {
		return errors.New("empty login")
	}

	_, err := r.pool.Exec(ctx, `DELETE FROM channel_blacklist WHERE login = $1`, login)
	if err != nil {
		r.obs.LogError(ctx, span, "remove blacklist failed", err, zap.String("login", login))
	}

	return err
}
