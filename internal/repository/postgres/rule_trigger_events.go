package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

func (r *Repository) InsertRuleTriggerEvent(ctx context.Context, ruleID int64, ruleName, triggerEvent, actionType, displayText string) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.insert_rule_trigger_event")
	defer span.End()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO rule_trigger_events (rule_id, rule_name, trigger_event, action_type, display_text)
		VALUES ($1, $2, $3, $4, $5)
	`, ruleID, ruleName, triggerEvent, actionType, displayText)
	if err != nil {
		r.obs.LogError(ctx, span, "insert rule trigger event failed", err,
			zap.Int64("rule_id", ruleID),
			zap.String("trigger_event", triggerEvent),
			zap.String("action_type", actionType))

		return err
	}

	return nil
}

func (r *Repository) ListRuleTriggerEvents(ctx context.Context, f entity.RuleTriggerListFilter) ([]entity.RuleTriggerEvent, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_rule_trigger_events")
	defer span.End()

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, created_at, rule_id, rule_name, trigger_event, action_type, display_text
		FROM rule_trigger_events
		WHERE ($1::timestamptz IS NULL OR $2::bigint IS NULL OR (created_at, id) < ($1, $2))
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`, f.CursorCreatedAt, f.CursorID, limit)
	if err != nil {
		r.obs.LogError(ctx, span, "list rule trigger events failed", err)

		return nil, err
	}

	defer rows.Close()

	out := make([]entity.RuleTriggerEvent, 0)

	for rows.Next() {
		var (
			e      entity.RuleTriggerEvent
			ruleID pgtype.Int8
		)

		if err := rows.Scan(&e.ID, &e.CreatedAt, &ruleID, &e.RuleName, &e.TriggerEvent, &e.ActionType, &e.DisplayText); err != nil {
			r.obs.LogError(ctx, span, "scan rule trigger event failed", err)

			return nil, err
		}

		if ruleID.Valid {
			v := ruleID.Int64
			e.RuleID = &v
		}

		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		r.obs.LogError(ctx, span, "rule trigger event rows iteration failed", err)

		return nil, err
	}

	return out, nil
}
