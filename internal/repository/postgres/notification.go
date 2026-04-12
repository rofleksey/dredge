package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

func (r *Repository) ListNotificationEntries(ctx context.Context) ([]entity.NotificationEntry, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_notification_entries")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, provider, settings, enabled, created_at
		FROM notification_entries ORDER BY id
	`)
	if err != nil {
		r.obs.LogError(ctx, span, "list notification entries failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.NotificationEntry, 0)

	for rows.Next() {
		var (
			e   entity.NotificationEntry
			raw []byte
		)

		if err := rows.Scan(&e.ID, &e.Provider, &raw, &e.Enabled, &e.CreatedAt); err != nil {
			r.obs.LogError(ctx, span, "scan notification entry failed", err)
			return nil, err
		}

		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &e.Settings)
		}

		if e.Settings == nil {
			e.Settings = map[string]any{}
		}

		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		r.obs.LogError(ctx, span, "notification rows iteration failed", err)
		return nil, err
	}

	return out, nil
}

func (r *Repository) ListEnabledNotificationEntries(ctx context.Context) ([]entity.NotificationEntry, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_enabled_notification_entries")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, provider, settings, enabled, created_at
		FROM notification_entries WHERE enabled = true ORDER BY id
	`)
	if err != nil {
		r.obs.LogError(ctx, span, "list enabled notification entries failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.NotificationEntry, 0)

	for rows.Next() {
		var (
			e   entity.NotificationEntry
			raw []byte
		)

		if err := rows.Scan(&e.ID, &e.Provider, &raw, &e.Enabled, &e.CreatedAt); err != nil {
			r.obs.LogError(ctx, span, "scan notification entry failed", err)
			return nil, err
		}

		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &e.Settings)
		}

		if e.Settings == nil {
			e.Settings = map[string]any{}
		}

		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		r.obs.LogError(ctx, span, "notification rows iteration failed", err)
		return nil, err
	}

	return out, nil
}

func (r *Repository) CreateNotificationEntry(ctx context.Context, provider string, settings map[string]any, enabled bool) (entity.NotificationEntry, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.create_notification_entry")
	defer span.End()

	if settings == nil {
		settings = map[string]any{}
	}

	raw, err := json.Marshal(settings)
	if err != nil {
		return entity.NotificationEntry{}, err
	}

	var e entity.NotificationEntry

	err = r.pool.QueryRow(ctx, `
		INSERT INTO notification_entries (provider, settings, enabled)
		VALUES ($1, $2::jsonb, $3)
		RETURNING id, provider, settings, enabled, created_at
	`, provider, raw, enabled).Scan(&e.ID, &e.Provider, &raw, &e.Enabled, &e.CreatedAt)
	if err != nil {
		r.obs.LogError(ctx, span, "create notification entry failed", err)
		return entity.NotificationEntry{}, err
	}

	_ = json.Unmarshal(raw, &e.Settings)
	if e.Settings == nil {
		e.Settings = map[string]any{}
	}

	return e, nil
}

func (r *Repository) UpdateNotificationEntry(ctx context.Context, id int64, provider *string, settings map[string]any, enabled *bool) (entity.NotificationEntry, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_notification_entry")
	defer span.End()

	// Load current
	var (
		cur entity.NotificationEntry
		raw []byte
	)

	err := r.pool.QueryRow(ctx, `
		SELECT id, provider, settings, enabled, created_at FROM notification_entries WHERE id = $1
	`, id).Scan(&cur.ID, &cur.Provider, &raw, &cur.Enabled, &cur.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.NotificationEntry{}, entity.ErrNotificationNotFound
		}
		return entity.NotificationEntry{}, err
	}

	_ = json.Unmarshal(raw, &cur.Settings)
	if cur.Settings == nil {
		cur.Settings = map[string]any{}
	}

	if provider != nil {
		cur.Provider = *provider
	}

	if settings != nil {
		cur.Settings = settings
	}

	if enabled != nil {
		cur.Enabled = *enabled
	}

	raw, err = json.Marshal(cur.Settings)
	if err != nil {
		return entity.NotificationEntry{}, err
	}

	err = r.pool.QueryRow(ctx, `
		UPDATE notification_entries SET provider = $2, settings = $3::jsonb, enabled = $4
		WHERE id = $1
		RETURNING id, provider, settings, enabled, created_at
	`, id, cur.Provider, raw, cur.Enabled).Scan(&cur.ID, &cur.Provider, &raw, &cur.Enabled, &cur.CreatedAt)
	if err != nil {
		r.obs.LogError(ctx, span, "update notification entry failed", err, zap.Int64("id", id))
		return entity.NotificationEntry{}, err
	}

	_ = json.Unmarshal(raw, &cur.Settings)

	return cur, nil
}

func (r *Repository) DeleteNotificationEntry(ctx context.Context, id int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.delete_notification_entry")
	defer span.End()

	tag, err := r.pool.Exec(ctx, `DELETE FROM notification_entries WHERE id = $1`, id)
	if err != nil {
		r.obs.LogError(ctx, span, "delete notification entry failed", err, zap.Int64("id", id))
		return err
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrNotificationNotFound
	}

	return nil
}
