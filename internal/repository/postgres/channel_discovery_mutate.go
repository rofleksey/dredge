package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// DenyDiscoveryCandidate removes a pending candidate and records the channel as permanently denied for discovery.
func (r *Repository) DenyDiscoveryCandidate(ctx context.Context, twitchUserID int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.deny_discovery_candidate")
	defer span.End()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.obs.LogError(ctx, span, "deny discovery begin tx failed", err)
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	cmd, err := tx.Exec(ctx, `DELETE FROM twitch_discovery_candidates WHERE twitch_user_id = $1`, twitchUserID)
	if err != nil {
		r.obs.LogError(ctx, span, "deny discovery delete candidate failed", err, zap.Int64("twitch_user_id", twitchUserID))
		return err
	}

	if cmd.RowsAffected() == 0 {
		return entity.ErrDiscoveryCandidateNotFound
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO twitch_discovery_denied (twitch_user_id) VALUES ($1)
		ON CONFLICT (twitch_user_id) DO NOTHING
	`, twitchUserID); err != nil {
		r.obs.LogError(ctx, span, "deny discovery insert denied failed", err, zap.Int64("twitch_user_id", twitchUserID))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		r.obs.LogError(ctx, span, "deny discovery commit failed", err)
		return err
	}

	return nil
}

// ApproveDiscoveryCandidate sets monitored=true and removes the pending discovery row in one transaction.
func (r *Repository) ApproveDiscoveryCandidate(ctx context.Context, twitchUserID int64) (entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.approve_discovery_candidate")
	defer span.End()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.obs.LogError(ctx, span, "approve discovery begin tx failed", err)
		return entity.TwitchUser{}, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	var lockOne int

	err = tx.QueryRow(ctx, `
		SELECT 1 FROM twitch_discovery_candidates WHERE twitch_user_id = $1 FOR UPDATE
	`, twitchUserID).Scan(&lockOne)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TwitchUser{}, entity.ErrDiscoveryCandidateNotFound
		}

		r.obs.LogError(ctx, span, "approve discovery lock candidate failed", err, zap.Int64("twitch_user_id", twitchUserID))
		return entity.TwitchUser{}, err
	}

	u, err := scanTwitchUser(tx.QueryRow(ctx, `
		UPDATE twitch_users SET monitored = true WHERE id = $1
		RETURNING id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
			irc_only_when_live, notify_off_stream_messages, notify_stream_start
	`, twitchUserID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TwitchUser{}, entity.ErrTwitchUserNotFound
		}

		r.obs.LogError(ctx, span, "approve discovery update monitored failed", err, zap.Int64("twitch_user_id", twitchUserID))
		return entity.TwitchUser{}, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM twitch_discovery_candidates WHERE twitch_user_id = $1`, twitchUserID); err != nil {
		r.obs.LogError(ctx, span, "approve discovery delete candidate failed", err, zap.Int64("twitch_user_id", twitchUserID))
		return entity.TwitchUser{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		r.obs.LogError(ctx, span, "approve discovery commit failed", err)
		return entity.TwitchUser{}, err
	}

	return u, nil
}
