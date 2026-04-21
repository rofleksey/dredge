package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// GetChannelDiscoverySettings returns the singleton row (id=1).
func (r *Repository) GetChannelDiscoverySettings(ctx context.Context) (entity.ChannelDiscoverySettings, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_channel_discovery_settings")
	defer span.End()

	var (
		s          entity.ChannelDiscoverySettings
		tags       []string
		pollSec    int32
		minViewers int32
		maxPages   int32
	)

	err := r.pool.QueryRow(ctx, `
		SELECT enabled, poll_interval_seconds, game_id, min_live_viewers, required_stream_tags, max_stream_pages_per_run
		FROM channel_discovery_settings WHERE id = 1
	`).Scan(
		&s.Enabled, &pollSec, &s.GameID, &minViewers, &tags, &maxPages,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return defaultChannelDiscoverySettings(), nil
		}

		r.obs.LogError(ctx, span, "get channel discovery settings failed", err)
	}

	s.PollIntervalSeconds = int(pollSec)
	s.MinLiveViewers = int(minViewers)
	s.MaxStreamPagesPerRun = int(maxPages)
	s.RequiredStreamTags = tags

	return s, err
}

func defaultChannelDiscoverySettings() entity.ChannelDiscoverySettings {
	return entity.ChannelDiscoverySettings{
		PollIntervalSeconds:  3600,
		GameID:               "",
		MinLiveViewers:       0,
		RequiredStreamTags:   nil,
		MaxStreamPagesPerRun: 20,
	}
}

// UpdateChannelDiscoverySettings upserts the singleton row.
func (r *Repository) UpdateChannelDiscoverySettings(ctx context.Context, s entity.ChannelDiscoverySettings) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_channel_discovery_settings")
	defer span.End()

	tags := s.RequiredStreamTags
	if tags == nil {
		tags = []string{}
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO channel_discovery_settings (
			id, enabled, poll_interval_seconds, game_id, min_live_viewers, required_stream_tags, max_stream_pages_per_run
		) VALUES (1, $1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			poll_interval_seconds = EXCLUDED.poll_interval_seconds,
			game_id = EXCLUDED.game_id,
			min_live_viewers = EXCLUDED.min_live_viewers,
			required_stream_tags = EXCLUDED.required_stream_tags,
			max_stream_pages_per_run = EXCLUDED.max_stream_pages_per_run
	`,
		s.Enabled, s.PollIntervalSeconds, s.GameID, s.MinLiveViewers, tags, s.MaxStreamPagesPerRun,
	)
	if err != nil {
		r.obs.LogError(ctx, span, "update channel discovery settings failed", err)
	}

	return err
}

// ListTwitchDiscoveryDeniedUserIDs returns all denied broadcaster twitch user ids.
func (r *Repository) ListTwitchDiscoveryDeniedUserIDs(ctx context.Context) ([]int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_twitch_discovery_denied_user_ids")
	defer span.End()

	rows, err := r.pool.Query(ctx, `SELECT twitch_user_id FROM twitch_discovery_denied ORDER BY twitch_user_id`)
	if err != nil {
		r.obs.LogError(ctx, span, "list twitch discovery denied failed", err)
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

// AddTwitchDiscoveryDenied records that a channel must never be suggested again by discovery.
func (r *Repository) AddTwitchDiscoveryDenied(ctx context.Context, twitchUserID int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.add_twitch_discovery_denied")
	defer span.End()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO twitch_discovery_denied (twitch_user_id) VALUES ($1)
		ON CONFLICT (twitch_user_id) DO NOTHING
	`, twitchUserID)
	if err != nil {
		r.obs.LogError(ctx, span, "add twitch discovery denied failed", err, zap.Int64("twitch_user_id", twitchUserID))
	}

	return err
}

// ListTwitchDiscoveryCandidates returns pending discovery rows with twitch_users joined.
func (r *Repository) ListTwitchDiscoveryCandidates(ctx context.Context) ([]entity.TwitchDiscoveryCandidate, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_twitch_discovery_candidates")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT c.discovered_at, c.last_seen_at, c.viewer_count, c.title, c.game_name, c.stream_tags,
			u.id, u.username, u.monitored, u.marked, u.is_sus, u.sus_type, u.sus_description, u.sus_auto_suppressed,
			u.irc_only_when_live, u.notify_off_stream_messages, u.notify_stream_start
		FROM twitch_discovery_candidates c
		JOIN twitch_users u ON u.id = c.twitch_user_id
		ORDER BY c.last_seen_at DESC
	`)
	if err != nil {
		r.obs.LogError(ctx, span, "list twitch discovery candidates failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.TwitchDiscoveryCandidate, 0)

	for rows.Next() {
		var (
			cand          entity.TwitchDiscoveryCandidate
			susType       pgtype.Text
			susDesc       pgtype.Text
			viewerCount   pgtype.Int8
			title         pgtype.Text
			gameName      pgtype.Text
			streamTags    []string
		)

		err := rows.Scan(
			&cand.DiscoveredAt, &cand.LastSeenAt, &viewerCount, &title, &gameName, &streamTags,
			&cand.User.ID, &cand.User.Username, &cand.User.Monitored, &cand.User.Marked, &cand.User.IsSus,
			&susType, &susDesc, &cand.User.SusAutoSuppressed,
			&cand.User.IrcOnlyWhenLive, &cand.User.NotifyOffStreamMessages, &cand.User.NotifyStreamStart,
		)
		if err != nil {
			r.obs.LogError(ctx, span, "scan twitch discovery candidate failed", err)
			return nil, err
		}

		if susType.Valid {
			s := susType.String
			cand.User.SusType = &s
		}

		if susDesc.Valid {
			s := susDesc.String
			cand.User.SusDescription = &s
		}

		if viewerCount.Valid {
			v := viewerCount.Int64
			cand.ViewerCount = &v
		}

		if title.Valid {
			t := title.String
			cand.Title = &t
		}

		if gameName.Valid {
			g := gameName.String
			cand.GameName = &g
		}

		cand.StreamTags = streamTags
		if cand.StreamTags == nil {
			cand.StreamTags = []string{}
		}

		out = append(out, cand)
	}

	return out, rows.Err()
}

// UpsertTwitchDiscoveryCandidate inserts or refreshes a pending discovery row.
func (r *Repository) UpsertTwitchDiscoveryCandidate(ctx context.Context, twitchUserID int64, viewerCount *int64, title, gameName *string, streamTags []string) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.upsert_twitch_discovery_candidate")
	defer span.End()

	if streamTags == nil {
		streamTags = []string{}
	}

	var vc pgtype.Int8

	if viewerCount != nil {
		vc = pgtype.Int8{Int64: *viewerCount, Valid: true}
	}

	var titlePG, gamePG pgtype.Text

	if title != nil {
		titlePG = pgtype.Text{String: *title, Valid: true}
	}

	if gameName != nil {
		gamePG = pgtype.Text{String: *gameName, Valid: true}
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO twitch_discovery_candidates (twitch_user_id, viewer_count, title, game_name, stream_tags)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (twitch_user_id) DO UPDATE SET
			last_seen_at = NOW(),
			viewer_count = EXCLUDED.viewer_count,
			title = EXCLUDED.title,
			game_name = EXCLUDED.game_name,
			stream_tags = EXCLUDED.stream_tags
	`, twitchUserID, vc, titlePG, gamePG, streamTags)
	if err != nil {
		r.obs.LogError(ctx, span, "upsert twitch discovery candidate failed", err, zap.Int64("twitch_user_id", twitchUserID))
	}

	return err
}

// DeleteTwitchDiscoveryCandidate removes a pending discovery row.
func (r *Repository) DeleteTwitchDiscoveryCandidate(ctx context.Context, twitchUserID int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.delete_twitch_discovery_candidate")
	defer span.End()

	cmd, err := r.pool.Exec(ctx, `DELETE FROM twitch_discovery_candidates WHERE twitch_user_id = $1`, twitchUserID)
	if err != nil {
		r.obs.LogError(ctx, span, "delete twitch discovery candidate failed", err, zap.Int64("twitch_user_id", twitchUserID))
		return err
	}

	if cmd.RowsAffected() == 0 {
		return entity.ErrDiscoveryCandidateNotFound
	}

	return nil
}
