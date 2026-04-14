package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
)

// GetIrcMonitorSettings returns the singleton row (id=1).
func (r *Repository) GetIrcMonitorSettings(ctx context.Context) (entity.IrcMonitorSettings, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_irc_monitor_settings")
	defer span.End()

	var (
		accID     *int64
		cooldown  time.Duration
		cooldownS int64
	)

	err := r.pool.QueryRow(ctx, `
		SELECT oauth_twitch_account_id, enrichment_cooldown_seconds
		FROM irc_monitor_settings WHERE id = 1
	`).Scan(&accID, &cooldownS)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.IrcMonitorSettings{EnrichmentCooldown: 24 * time.Hour}, nil
		}

		r.obs.LogError(ctx, span, "get irc monitor settings failed", err)
	}

	if cooldownS <= 0 {
		cooldown = 24 * time.Hour
	} else {
		cooldown = time.Duration(cooldownS) * time.Second
	}

	return entity.IrcMonitorSettings{
		OauthTwitchAccountID: accID,
		EnrichmentCooldown:   cooldown,
	}, err
}

// UpdateIrcMonitorSettings replaces the singleton row.
func (r *Repository) UpdateIrcMonitorSettings(ctx context.Context, s entity.IrcMonitorSettings) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_irc_monitor_settings")
	defer span.End()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO irc_monitor_settings (id, oauth_twitch_account_id, enrichment_cooldown_seconds) VALUES (1, $1, $2)
		ON CONFLICT (id) DO UPDATE SET
			oauth_twitch_account_id = EXCLUDED.oauth_twitch_account_id,
			enrichment_cooldown_seconds = EXCLUDED.enrichment_cooldown_seconds
	`, s.OauthTwitchAccountID, int64(s.EnrichmentCooldown/time.Second))
	if err != nil {
		r.obs.LogError(ctx, span, "update irc monitor settings failed", err)
	}

	return err
}
