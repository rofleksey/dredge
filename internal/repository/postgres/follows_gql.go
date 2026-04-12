package postgres

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// ReplaceUserFollowedChannels replaces all GQL-synced follows for a follower (transactional).
func (r *Repository) ReplaceUserFollowedChannels(ctx context.Context, followerID int64, rows []entity.FollowedChannelRow) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.replace_user_followed_channels")
	defer span.End()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.obs.LogError(ctx, span, "begin replace follows failed", err)
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM user_followed_channels WHERE follower_twitch_user_id = $1`, followerID); err != nil {
		r.obs.LogError(ctx, span, "delete old follows failed", err, zap.Int64("follower_id", followerID))
		return err
	}

	synced := time.Now().UTC()

	for _, row := range rows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_followed_channels (follower_twitch_user_id, followed_channel_id, followed_channel_login, followed_at, synced_at)
			VALUES ($1, $2, $3, $4, $5)
		`, followerID, row.FollowedChannelID, row.FollowedChannelLogin, row.FollowedAt, synced); err != nil {
			r.obs.LogError(ctx, span, "insert follow failed", err, zap.Int64("follower_id", followerID))
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.obs.LogError(ctx, span, "commit replace follows failed", err)
		return err
	}

	return nil
}

// ListUserFollowedChannels returns all GQL-synced follows for a user (for profile API).
func (r *Repository) ListUserFollowedChannels(ctx context.Context, followerID int64) ([]entity.FollowedChannelRow, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_user_followed_channels")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT followed_channel_id, followed_channel_login, followed_at
		FROM user_followed_channels
		WHERE follower_twitch_user_id = $1
		ORDER BY followed_channel_login ASC
	`, followerID)
	if err != nil {
		r.obs.LogError(ctx, span, "list user followed channels failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []entity.FollowedChannelRow

	for rows.Next() {
		var (
			row entity.FollowedChannelRow
			fa  *time.Time
		)

		if err := rows.Scan(&row.FollowedChannelID, &row.FollowedChannelLogin, &fa); err != nil {
			return nil, err
		}

		row.FollowedAt = fa
		out = append(out, row)
	}

	return out, rows.Err()
}
