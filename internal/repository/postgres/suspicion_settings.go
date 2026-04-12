package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
)

// GetSuspicionSettings returns the singleton row (id=1).
func (r *Repository) GetSuspicionSettings(ctx context.Context) (entity.SuspicionSettings, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_suspicion_settings")
	defer span.End()

	var s entity.SuspicionSettings

	err := r.pool.QueryRow(ctx, `
		SELECT auto_check_account_age, account_age_sus_days, auto_check_blacklist, auto_check_low_follows,
			low_follows_threshold, max_gql_follow_pages
		FROM suspicion_settings WHERE id = 1
	`).Scan(
		&s.AutoCheckAccountAge, &s.AccountAgeSusDays, &s.AutoCheckBlacklist, &s.AutoCheckLowFollows,
		&s.LowFollowsThreshold, &s.MaxGQLFollowPages,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return defaultSuspicionSettings(), nil
		}

		r.obs.LogError(ctx, span, "get suspicion settings failed", err)
	}

	return s, err
}

func defaultSuspicionSettings() entity.SuspicionSettings {
	return entity.SuspicionSettings{
		AutoCheckAccountAge: true,
		AccountAgeSusDays:   14,
		AutoCheckBlacklist:  true,
		AutoCheckLowFollows: true,
		LowFollowsThreshold: 10,
		MaxGQLFollowPages:   1,
	}
}

// UpdateSuspicionSettings updates the singleton row (partial nil = leave unchanged — caller passes full struct from GET-merge).
func (r *Repository) UpdateSuspicionSettings(ctx context.Context, s entity.SuspicionSettings) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_suspicion_settings")
	defer span.End()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO suspicion_settings (
			id, auto_check_account_age, account_age_sus_days, auto_check_blacklist, auto_check_low_follows,
			low_follows_threshold, max_gql_follow_pages
		) VALUES (1, $1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			auto_check_account_age = EXCLUDED.auto_check_account_age,
			account_age_sus_days = EXCLUDED.account_age_sus_days,
			auto_check_blacklist = EXCLUDED.auto_check_blacklist,
			auto_check_low_follows = EXCLUDED.auto_check_low_follows,
			low_follows_threshold = EXCLUDED.low_follows_threshold,
			max_gql_follow_pages = EXCLUDED.max_gql_follow_pages
	`,
		s.AutoCheckAccountAge, s.AccountAgeSusDays, s.AutoCheckBlacklist, s.AutoCheckLowFollows,
		s.LowFollowsThreshold, s.MaxGQLFollowPages,
	)
	if err != nil {
		r.obs.LogError(ctx, span, "update suspicion settings failed", err)
	}

	return err
}

// ListLinkedTwitchAccountUserIDs returns Twitch user ids for OAuth-linked accounts (never mark as suspicious).
func (r *Repository) ListLinkedTwitchAccountUserIDs(ctx context.Context) ([]int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_linked_twitch_account_ids")
	defer span.End()

	rows, err := r.pool.Query(ctx, `SELECT id FROM twitch_accounts WHERE deleted_at IS NULL`)
	if err != nil {
		r.obs.LogError(ctx, span, "list linked account ids failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		out = append(out, id)
	}

	return out, rows.Err()
}
