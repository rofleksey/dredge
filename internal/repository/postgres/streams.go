package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
)

// ActiveStreamIDForChannel returns the open stream id for a channel, if any.
func (r *Repository) ActiveStreamIDForChannel(ctx context.Context, channelTwitchUserID int64) (*int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.active_stream_id_for_channel")
	defer span.End()

	var id sql.NullInt64

	err := r.pool.QueryRow(ctx, `
		SELECT id FROM streams
		WHERE channel_twitch_user_id = $1 AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`, channelTwitchUserID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		r.obs.LogError(ctx, span, "active stream id query failed", err)
		return nil, err
	}

	if !id.Valid {
		return nil, nil
	}

	v := id.Int64

	return &v, nil
}

// UpsertStreamFromHelix closes any other open session on the channel, then inserts or updates by helix_stream_id.
func (r *Repository) UpsertStreamFromHelix(ctx context.Context, channelTwitchUserID int64, helixStreamID string, startedAt time.Time, title, gameName string, viewerCount *int64) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.upsert_stream_from_helix")
	defer span.End()

	if helixStreamID == "" {
		return 0, errors.New("empty helix stream id")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	var existingID int64

	err = tx.QueryRow(ctx, `SELECT id FROM streams WHERE helix_stream_id = $1`, helixStreamID).Scan(&existingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		r.obs.LogError(ctx, span, "select stream by helix id failed", err)
		return 0, err
	}

	if err == nil {
		_, err = tx.Exec(ctx, `
			UPDATE streams SET title = $2, game_name = $3, viewer_count = $4, helix_synced_at = NOW()
			WHERE id = $1
		`, existingID, nullIfEmpty(title), nullIfEmpty(gameName), viewerCount)
		if err != nil {
			r.obs.LogError(ctx, span, "update stream meta failed", err)
			return 0, err
		}

		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}

		return existingID, nil
	}

	_, err = tx.Exec(ctx, `
		UPDATE streams SET ended_at = NOW()
		WHERE channel_twitch_user_id = $1 AND ended_at IS NULL AND helix_stream_id <> $2
	`, channelTwitchUserID, helixStreamID)
	if err != nil {
		r.obs.LogError(ctx, span, "close stale open streams failed", err)
		return 0, err
	}

	var newID int64

	err = tx.QueryRow(ctx, `
		INSERT INTO streams (channel_twitch_user_id, helix_stream_id, started_at, ended_at, title, game_name, viewer_count, helix_synced_at)
		VALUES ($1, $2, $3, NULL, $4, $5, $6, NOW())
		RETURNING id
	`, channelTwitchUserID, helixStreamID, startedAt, nullIfEmpty(title), nullIfEmpty(gameName), viewerCount).Scan(&newID)
	if err != nil {
		r.obs.LogError(ctx, span, "insert stream failed", err)
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return newID, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	return s
}

// CloseOpenStreamsForChannel sets ended_at on any open stream for the channel.
func (r *Repository) CloseOpenStreamsForChannel(ctx context.Context, channelTwitchUserID int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.close_open_streams_for_channel")
	defer span.End()

	_, err := r.pool.Exec(ctx, `
		UPDATE streams SET ended_at = NOW()
		WHERE channel_twitch_user_id = $1 AND ended_at IS NULL
	`, channelTwitchUserID)
	if err != nil {
		r.obs.LogError(ctx, span, "close open streams for channel failed", err)
	}

	return err
}

func scanStreamRow(rows interface {
	Scan(dest ...any) error
}) (entity.Stream, error) {
	var s entity.Stream

	var ended sql.NullTime

	err := rows.Scan(
		&s.ID,
		&s.ChannelTwitchUserID,
		&s.ChannelLogin,
		&s.HelixStreamID,
		&s.StartedAt,
		&ended,
		&s.Title,
		&s.GameName,
		&s.CreatedAt,
	)
	if err != nil {
		return s, err
	}

	if ended.Valid {
		t := ended.Time
		s.EndedAt = &t
	}

	return s, nil
}

// GetStreamByID loads a stream joined with channel login.
func (r *Repository) GetStreamByID(ctx context.Context, id int64) (entity.Stream, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_stream_by_id")
	defer span.End()

	row := r.pool.QueryRow(ctx, `
		SELECT s.id, s.channel_twitch_user_id, u.username, s.helix_stream_id, s.started_at, s.ended_at, COALESCE(s.title, ''), COALESCE(s.game_name, ''), s.created_at
		FROM streams s
		INNER JOIN twitch_users u ON u.id = s.channel_twitch_user_id
		WHERE s.id = $1
	`, id)

	s, err := scanStreamRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Stream{}, err
	}

	if err != nil {
		r.obs.LogError(ctx, span, "get stream by id failed", err)
	}

	return s, err
}

// GetMonitoredStreamByID returns the stream only if the channel is monitored.
func (r *Repository) GetMonitoredStreamByID(ctx context.Context, id int64) (entity.Stream, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_monitored_stream_by_id")
	defer span.End()

	row := r.pool.QueryRow(ctx, `
		SELECT s.id, s.channel_twitch_user_id, u.username, s.helix_stream_id, s.started_at, s.ended_at, COALESCE(s.title, ''), COALESCE(s.game_name, ''), s.created_at
		FROM streams s
		INNER JOIN twitch_users u ON u.id = s.channel_twitch_user_id AND u.monitored = true
		WHERE s.id = $1
	`, id)

	s, err := scanStreamRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Stream{}, err
	}

	if err != nil {
		r.obs.LogError(ctx, span, "get monitored stream by id failed", err)
	}

	return s, err
}

// ListMonitoredStreams returns recorded streams for monitored channels only, newest first.
func (r *Repository) ListMonitoredStreams(ctx context.Context, f entity.StreamListFilter) ([]entity.Stream, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_monitored_streams")
	defer span.End()

	limit := f.Limit
	if limit < 1 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	q := `
		SELECT s.id, s.channel_twitch_user_id, u.username, s.helix_stream_id, s.started_at, s.ended_at, COALESCE(s.title, ''), COALESCE(s.game_name, ''), s.created_at
		FROM streams s
		INNER JOIN twitch_users u ON u.id = s.channel_twitch_user_id AND u.monitored = true
		WHERE 1=1
	`
	args := make([]any, 0, 8)
	n := 1

	if ch := normalizeChannelName(f.ChannelLogin); ch != "" {
		q += ` AND lower(u.username) = lower($` + strconv.Itoa(n) + `)`

		args = append(args, ch)
		n++
	}

	if f.CursorStartedAt != nil && f.CursorID != nil {
		q += ` AND (s.started_at, s.id) < ($` + strconv.Itoa(n) + `, $` + strconv.Itoa(n+1) + `)`

		args = append(args, *f.CursorStartedAt, *f.CursorID)
		n += 2
	}

	q += ` ORDER BY s.started_at DESC, s.id DESC LIMIT $` + strconv.Itoa(n)

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		r.obs.LogError(ctx, span, "list monitored streams failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.Stream, 0, limit)

	for rows.Next() {
		s, err := scanStreamRow(rows)
		if err != nil {
			r.obs.LogError(ctx, span, "scan stream row failed", err)
			return nil, err
		}

		out = append(out, s)
	}

	return out, rows.Err()
}

// CountChatMessagesPerChatterForStream aggregates message counts by chatter for a stream.
func (r *Repository) CountChatMessagesPerChatterForStream(ctx context.Context, streamID int64) (map[int64]int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_chat_messages_per_chatter_for_stream")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT chatter_twitch_user_id, count(*)::bigint
		FROM chat_messages
		WHERE stream_id = $1 AND chatter_twitch_user_id IS NOT NULL
		GROUP BY chatter_twitch_user_id
	`, streamID)
	if err != nil {
		r.obs.LogError(ctx, span, "count messages per chatter for stream failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]int64)

	for rows.Next() {
		var uid, cnt int64
		if err := rows.Scan(&uid, &cnt); err != nil {
			return nil, err
		}

		out[uid] = cnt
	}

	return out, rows.Err()
}

// CountChatMessagesPerChatterForChannelSince returns message counts per chatter since an instant (inclusive), for one channel.
func (r *Repository) CountChatMessagesPerChatterForChannelSince(ctx context.Context, channelTwitchUserID int64, since time.Time) (map[int64]int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_chat_messages_per_chatter_channel_since")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT m.chatter_twitch_user_id, count(*)::bigint
		FROM chat_messages m
		WHERE m.twitch_user_id = $1
		  AND m.chatter_twitch_user_id IS NOT NULL
		  AND m.created_at >= $2
		GROUP BY m.chatter_twitch_user_id
	`, channelTwitchUserID, since)
	if err != nil {
		r.obs.LogError(ctx, span, "count messages per chatter channel since failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]int64)

	for rows.Next() {
		var uid, cnt int64
		if err := rows.Scan(&uid, &cnt); err != nil {
			return nil, err
		}

		out[uid] = cnt
	}

	return out, rows.Err()
}
