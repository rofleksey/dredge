package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
)

// GetIrcMonitorSettings returns the singleton row (id=1).
func (r *Repository) GetIrcMonitorSettings(ctx context.Context) (entity.IrcMonitorSettings, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_irc_monitor_settings")
	defer span.End()

	var accID *int64

	err := r.pool.QueryRow(ctx, `
		SELECT oauth_twitch_account_id FROM irc_monitor_settings WHERE id = 1
	`).Scan(&accID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.IrcMonitorSettings{}, nil
		}

		r.obs.LogError(ctx, span, "get irc monitor settings failed", err)
	}

	return entity.IrcMonitorSettings{OauthTwitchAccountID: accID}, err
}

// UpdateIrcMonitorSettings replaces the singleton row.
func (r *Repository) UpdateIrcMonitorSettings(ctx context.Context, s entity.IrcMonitorSettings) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_irc_monitor_settings")
	defer span.End()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO irc_monitor_settings (id, oauth_twitch_account_id) VALUES (1, $1)
		ON CONFLICT (id) DO UPDATE SET oauth_twitch_account_id = EXCLUDED.oauth_twitch_account_id
	`, s.OauthTwitchAccountID)
	if err != nil {
		r.obs.LogError(ctx, span, "update irc monitor settings failed", err)
	}

	return err
}
