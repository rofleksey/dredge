package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

func (r *Repository) ListRules(ctx context.Context) ([]entity.Rule, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_rules")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, regex, included_users, denied_users, included_channels, denied_channels
		FROM rules ORDER BY id
	`)
	if err != nil {
		r.obs.LogError(ctx, span, "list rules query failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Rule, 0)

	for rows.Next() {
		var rr entity.Rule
		if err := rows.Scan(&rr.ID, &rr.Regex, &rr.IncludedUsers, &rr.DeniedUsers, &rr.IncludedChannels, &rr.DeniedChannels); err != nil {
			r.obs.LogError(ctx, span, "scan rule failed", err)
			return nil, err
		}

		out = append(out, rr)
	}

	err = rows.Err()
	if err != nil {
		r.obs.LogError(ctx, span, "rows iteration failed", err)
	}

	return out, err
}

func (r *Repository) CountRules(ctx context.Context) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_rules")
	defer span.End()

	var n int64

	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM rules`).Scan(&n)
	if err != nil {
		r.obs.LogError(ctx, span, "count rules failed", err)
	}
	return n, err
}

func (r *Repository) CreateRule(ctx context.Context, rr entity.Rule) (entity.Rule, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.create_rule")
	defer span.End()

	err := r.pool.QueryRow(ctx, `
		INSERT INTO rules (regex, included_users, denied_users, included_channels, denied_channels)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, regex, included_users, denied_users, included_channels, denied_channels
	`, rr.Regex, rr.IncludedUsers, rr.DeniedUsers, rr.IncludedChannels, rr.DeniedChannels).Scan(
		&rr.ID, &rr.Regex, &rr.IncludedUsers, &rr.DeniedUsers, &rr.IncludedChannels, &rr.DeniedChannels,
	)
	if err != nil {
		r.obs.LogError(ctx, span, "create rule failed", err, zap.String("regex", rr.Regex))
	}

	return rr, err
}

func (r *Repository) UpdateRule(ctx context.Context, id int64, rr entity.Rule) (entity.Rule, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_rule")
	defer span.End()

	err := r.pool.QueryRow(ctx, `
		UPDATE rules SET
			regex = $2,
			included_users = $3,
			denied_users = $4,
			included_channels = $5,
			denied_channels = $6
		WHERE id = $1
		RETURNING id, regex, included_users, denied_users, included_channels, denied_channels
	`, id, rr.Regex, rr.IncludedUsers, rr.DeniedUsers, rr.IncludedChannels, rr.DeniedChannels).Scan(
		&rr.ID, &rr.Regex, &rr.IncludedUsers, &rr.DeniedUsers, &rr.IncludedChannels, &rr.DeniedChannels,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Rule{}, entity.ErrRuleNotFound
		}

		r.obs.LogError(ctx, span, "update rule failed", err, zap.Int64("id", id))
	}

	return rr, err
}

func (r *Repository) DeleteRule(ctx context.Context, id int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.delete_rule")
	defer span.End()

	tag, err := r.pool.Exec(ctx, `DELETE FROM rules WHERE id = $1`, id)
	if err != nil {
		r.obs.LogError(ctx, span, "delete rule failed", err, zap.Int64("id", id))
		return err
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrRuleNotFound
	}

	return nil
}
