package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

func (r *Repository) ListRules(ctx context.Context) ([]entity.Rule, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_rules")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, enabled, event_type, event_settings, middlewares, action_type, action_settings,
			use_shared_pool, created_at, updated_at
		FROM rules ORDER BY id
	`)
	if err != nil {
		r.obs.LogError(ctx, span, "list rules query failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Rule, 0)

	for rows.Next() {
		rr, err := scanRule(rows)
		if err != nil {
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

func scanRule(row interface {
	Scan(dest ...any) error
}) (entity.Rule, error) {
	var (
		rr            entity.Rule
		eventJSON     []byte
		middlewaresJSON []byte
		actionJSON    []byte
	)

	err := row.Scan(
		&rr.ID,
		&rr.Name,
		&rr.Enabled,
		&rr.EventType,
		&eventJSON,
		&middlewaresJSON,
		&rr.ActionType,
		&actionJSON,
		&rr.UseSharedPool,
		&rr.CreatedAt,
		&rr.UpdatedAt,
	)
	if err != nil {
		return entity.Rule{}, err
	}

	if len(eventJSON) > 0 {
		if err := json.Unmarshal(eventJSON, &rr.EventSettings); err != nil {
			return entity.Rule{}, err
		}
	}

	if rr.EventSettings == nil {
		rr.EventSettings = map[string]any{}
	}

	if len(middlewaresJSON) > 0 {
		if err := json.Unmarshal(middlewaresJSON, &rr.Middlewares); err != nil {
			return entity.Rule{}, err
		}
	}

	if len(actionJSON) > 0 {
		if err := json.Unmarshal(actionJSON, &rr.ActionSettings); err != nil {
			return entity.Rule{}, err
		}
	}

	if rr.ActionSettings == nil {
		rr.ActionSettings = map[string]any{}
	}

	return rr, nil
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

	eventJSON, err := json.Marshal(rr.EventSettings)
	if err != nil {
		return entity.Rule{}, err
	}

	mwJSON, err := json.Marshal(rr.Middlewares)
	if err != nil {
		return entity.Rule{}, err
	}

	actionJSON, err := json.Marshal(rr.ActionSettings)
	if err != nil {
		return entity.Rule{}, err
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO rules (name, enabled, event_type, event_settings, middlewares, action_type, action_settings, use_shared_pool)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, enabled, event_type, event_settings, middlewares, action_type, action_settings,
			use_shared_pool, created_at, updated_at
	`, rr.Name, rr.Enabled, rr.EventType, eventJSON, mwJSON, rr.ActionType, actionJSON, rr.UseSharedPool).Scan(
		&rr.ID,
		&rr.Name,
		&rr.Enabled,
		&rr.EventType,
		&eventJSON,
		&mwJSON,
		&rr.ActionType,
		&actionJSON,
		&rr.UseSharedPool,
		&rr.CreatedAt,
		&rr.UpdatedAt,
	)
	if err != nil {
		r.obs.LogError(ctx, span, "create rule failed", err, zap.String("event_type", rr.EventType))
		return entity.Rule{}, err
	}

	return scanRuleFromInsert(rr, eventJSON, mwJSON, actionJSON)
}

func scanRuleFromInsert(rr entity.Rule, eventJSON, mwJSON, actionJSON []byte) (entity.Rule, error) {
	rr.EventSettings = nil
	if len(eventJSON) > 0 {
		if err := json.Unmarshal(eventJSON, &rr.EventSettings); err != nil {
			return entity.Rule{}, err
		}
	}

	if rr.EventSettings == nil {
		rr.EventSettings = map[string]any{}
	}

	rr.Middlewares = nil
	if len(mwJSON) > 0 {
		if err := json.Unmarshal(mwJSON, &rr.Middlewares); err != nil {
			return entity.Rule{}, err
		}
	}

	if rr.ActionSettings == nil {
		rr.ActionSettings = map[string]any{}
	}

	if len(actionJSON) > 0 {
		if err := json.Unmarshal(actionJSON, &rr.ActionSettings); err != nil {
			return entity.Rule{}, err
		}
	}

	return rr, nil
}

func (r *Repository) UpdateRule(ctx context.Context, id int64, rr entity.Rule) (entity.Rule, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_rule")
	defer span.End()

	eventJSON, err := json.Marshal(rr.EventSettings)
	if err != nil {
		return entity.Rule{}, err
	}

	mwJSON, err := json.Marshal(rr.Middlewares)
	if err != nil {
		return entity.Rule{}, err
	}

	actionJSON, err := json.Marshal(rr.ActionSettings)
	if err != nil {
		return entity.Rule{}, err
	}

	err = r.pool.QueryRow(ctx, `
		UPDATE rules SET
			name = $2,
			enabled = $3,
			event_type = $4,
			event_settings = $5,
			middlewares = $6,
			action_type = $7,
			action_settings = $8,
			use_shared_pool = $9,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, enabled, event_type, event_settings, middlewares, action_type, action_settings,
			use_shared_pool, created_at, updated_at
	`, id, rr.Name, rr.Enabled, rr.EventType, eventJSON, mwJSON, rr.ActionType, actionJSON, rr.UseSharedPool).Scan(
		&rr.ID,
		&rr.Name,
		&rr.Enabled,
		&rr.EventType,
		&eventJSON,
		&mwJSON,
		&rr.ActionType,
		&actionJSON,
		&rr.UseSharedPool,
		&rr.CreatedAt,
		&rr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Rule{}, entity.ErrRuleNotFound
		}

		r.obs.LogError(ctx, span, "update rule failed", err, zap.Int64("id", id))
		return entity.Rule{}, err
	}

	return scanRuleFromInsert(rr, eventJSON, mwJSON, actionJSON)
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
