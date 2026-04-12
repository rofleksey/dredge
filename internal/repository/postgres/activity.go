package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// TruncateChannelChatters clears all in-chat snapshots (e.g. when IRC monitor restarts).
func (r *Repository) TruncateChannelChatters(ctx context.Context) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.truncate_channel_chatters")
	defer span.End()

	_, err := r.pool.Exec(ctx, `TRUNCATE channel_chatters`)
	if err != nil {
		r.obs.LogError(ctx, span, "truncate channel_chatters failed", err)
	}
	return err
}

// ReplaceChannelChattersSnapshot merges the IRC NAMES list with the DB: upserts new rows (present_since set on first insert only), deletes missing.
func (r *Repository) ReplaceChannelChattersSnapshot(ctx context.Context, channelTwitchUserID int64, chatterIDs []int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.replace_channel_chatters_snapshot")
	defer span.End()

	want := make(map[int64]struct{})

	for _, cid := range chatterIDs {
		if cid != 0 {
			want[cid] = struct{}{}
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.obs.LogError(ctx, span, "begin tx failed", err)
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	if len(want) == 0 {
		_, err = tx.Exec(ctx, `DELETE FROM channel_chatters WHERE channel_twitch_user_id = $1`, channelTwitchUserID)
		if err != nil {
			r.obs.LogError(ctx, span, "delete channel chatters failed", err)
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			r.obs.LogError(ctx, span, "commit channel chatters failed", err)
			return err
		}

		return nil
	}

	uniq := make([]int64, 0, len(want))
	for id := range want {
		uniq = append(uniq, id)
	}

	for _, cid := range uniq {
		_, err = tx.Exec(ctx, `
			INSERT INTO channel_chatters (channel_twitch_user_id, chatter_twitch_user_id, present_since, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (channel_twitch_user_id, chatter_twitch_user_id) DO UPDATE SET
				updated_at = EXCLUDED.updated_at
		`, channelTwitchUserID, cid)
		if err != nil {
			r.obs.LogError(ctx, span, "upsert channel chatter failed", err)
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM channel_chatters
		WHERE channel_twitch_user_id = $1
		  AND chatter_twitch_user_id <> ALL($2::bigint[])
	`, channelTwitchUserID, uniq)
	if err != nil {
		r.obs.LogError(ctx, span, "prune channel chatters failed", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		r.obs.LogError(ctx, span, "commit channel chatters failed", err)
		return err
	}

	return nil
}

// UpsertChannelChatterPresence records a chatter in channel_chatters (IRC JOIN). present_since is set only on first insert.
func (r *Repository) UpsertChannelChatterPresence(ctx context.Context, channelTwitchUserID, chatterTwitchUserID int64) (time.Time, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.upsert_channel_chatter_presence")
	defer span.End()

	var since time.Time

	err := r.pool.QueryRow(ctx, `
		INSERT INTO channel_chatters (channel_twitch_user_id, chatter_twitch_user_id, present_since, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (channel_twitch_user_id, chatter_twitch_user_id) DO UPDATE SET
			updated_at = EXCLUDED.updated_at
		RETURNING present_since
	`, channelTwitchUserID, chatterTwitchUserID).Scan(&since)
	if err != nil {
		r.obs.LogError(ctx, span, "upsert channel chatter presence failed", err)
		return time.Time{}, err
	}

	return since, nil
}

// DeleteChannelChatterPresence removes a chatter row (IRC PART) and returns present_since when a row existed.
func (r *Repository) DeleteChannelChatterPresence(ctx context.Context, channelTwitchUserID, chatterTwitchUserID int64) (time.Time, bool, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.delete_channel_chatter_presence")
	defer span.End()

	var since time.Time

	err := r.pool.QueryRow(ctx, `
		DELETE FROM channel_chatters
		WHERE channel_twitch_user_id = $1 AND chatter_twitch_user_id = $2
		RETURNING present_since
	`, channelTwitchUserID, chatterTwitchUserID).Scan(&since)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}

	if err != nil {
		r.obs.LogError(ctx, span, "delete channel chatter presence failed", err)
		return time.Time{}, false, err
	}

	return since, true, nil
}

// ListChannelChatterIDs returns chatter ids currently recorded for a channel.
func (r *Repository) ListChannelChatterIDs(ctx context.Context, channelTwitchUserID int64) ([]int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_channel_chatter_ids")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT chatter_twitch_user_id FROM channel_chatters WHERE channel_twitch_user_id = $1
	`, channelTwitchUserID)
	if err != nil {
		r.obs.LogError(ctx, span, "list channel chatters failed", err)
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

// CountChannelChatters returns how many chat participants are recorded for the channel (IRC snapshot).
func (r *Repository) CountChannelChatters(ctx context.Context, channelTwitchUserID int64) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_channel_chatters")
	defer span.End()

	var n int64

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::bigint FROM channel_chatters WHERE channel_twitch_user_id = $1
	`, channelTwitchUserID).Scan(&n)
	if err != nil {
		r.obs.LogError(ctx, span, "count channel chatters failed", err)

		return 0, err
	}

	return n, nil
}

// ListChannelChatterEntries returns chatters with presence and Helix account creation when known.
func (r *Repository) ListChannelChatterEntries(ctx context.Context, channelTwitchUserID int64) ([]entity.ChannelChatterEntry, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_channel_chatter_entries")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.username, c.present_since, h.account_created_at
		FROM channel_chatters c
		INNER JOIN twitch_users u ON u.id = c.chatter_twitch_user_id
		LEFT JOIN twitch_user_helix_meta h ON h.twitch_user_id = u.id
		WHERE c.channel_twitch_user_id = $1
		ORDER BY u.username ASC
	`, channelTwitchUserID)
	if err != nil {
		r.obs.LogError(ctx, span, "list channel chatter entries failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []entity.ChannelChatterEntry

	for rows.Next() {
		var (
			e          entity.ChannelChatterEntry
			accountCre sql.NullTime
		)

		if err := rows.Scan(&e.UserTwitchID, &e.Login, &e.PresentSince, &accountCre); err != nil {
			return nil, err
		}

		if accountCre.Valid {
			t := accountCre.Time
			e.AccountCreatedAt = &t
		}

		out = append(out, e)
	}

	return out, rows.Err()
}

// InsertUserActivityEvent appends one activity row.
func (r *Repository) InsertUserActivityEvent(ctx context.Context, chatterID int64, eventType string, channelTwitchUserID *int64, details map[string]any) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.insert_user_activity_event")
	defer span.End()

	if details == nil {
		details = map[string]any{}
	}

	dj, err := json.Marshal(details)
	if err != nil {
		r.obs.LogError(ctx, span, "marshal activity details failed", err)
		return err
	}

	var ch sql.NullInt64
	if channelTwitchUserID != nil && *channelTwitchUserID != 0 {
		ch = sql.NullInt64{Int64: *channelTwitchUserID, Valid: true}
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO user_activity_events (chatter_twitch_user_id, event_type, channel_twitch_user_id, details)
		VALUES ($1, $2, $3, $4::jsonb)
	`, chatterID, eventType, ch, dj)
	if err != nil {
		r.obs.LogError(ctx, span, "insert activity event failed", err)
	}

	return err
}

// ListUserActivityEvents lists activity for a chatter (newest first).
func (r *Repository) ListUserActivityEvents(ctx context.Context, f entity.UserActivityListFilter) ([]entity.UserActivityEvent, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_user_activity_events")
	defer span.End()

	limit := f.Limit
	if limit < 1 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	q := strings.Builder{}
	q.WriteString(`
		SELECT e.id, e.chatter_twitch_user_id, e.event_type, e.channel_twitch_user_id, uc.username, e.details, e.created_at
		FROM user_activity_events e
		LEFT JOIN twitch_users uc ON uc.id = e.channel_twitch_user_id
		WHERE e.chatter_twitch_user_id = $1
		  AND e.event_type <> 'message'
	`)

	args := []any{f.ChatterUserID}
	n := 2

	if f.CursorCreatedAt != nil && f.CursorID != nil {
		q.WriteString(` AND (e.created_at, e.id) < ($` + strconv.Itoa(n) + `, $` + strconv.Itoa(n+1) + `)`)

		args = append(args, *f.CursorCreatedAt, *f.CursorID)
		n += 2
	}

	q.WriteString(` ORDER BY e.created_at DESC, e.id DESC LIMIT $` + strconv.Itoa(n))

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, q.String(), args...)
	if err != nil {
		r.obs.LogError(ctx, span, "list activity events failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.UserActivityEvent, 0, limit)

	for rows.Next() {
		var (
			e          entity.UserActivityEvent
			ch         sql.NullInt64
			chLogin    sql.NullString
			detailsRaw []byte
		)

		if err := rows.Scan(&e.ID, &e.ChatterTwitchUserID, &e.EventType, &ch, &chLogin, &detailsRaw, &e.CreatedAt); err != nil {
			return nil, err
		}

		if ch.Valid {
			v := ch.Int64
			e.ChannelTwitchUserID = &v
		}

		if chLogin.Valid {
			e.ChannelLogin = chLogin.String
		}

		if len(detailsRaw) > 0 {
			if err := json.Unmarshal(detailsRaw, &e.Details); err != nil {
				r.obs.Zap().Warn("unmarshal user_activity_events.details failed", zap.Error(err),
					zap.Int64("event_id", e.ID), zap.Int64("chatter_twitch_user_id", e.ChatterTwitchUserID))
			}
		}

		out = append(out, e)
	}

	return out, rows.Err()
}

func (r *Repository) ListUserActivityEventsForTimeline(ctx context.Context, chatterID int64, from, to time.Time) ([]entity.UserActivityEvent, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_user_activity_timeline")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.chatter_twitch_user_id, e.event_type, e.channel_twitch_user_id, uc.username, e.details, e.created_at
		FROM user_activity_events e
		LEFT JOIN twitch_users uc ON uc.id = e.channel_twitch_user_id
		WHERE e.chatter_twitch_user_id = $1
		  AND e.created_at >= $2 AND e.created_at <= $3
		  AND e.event_type IN ($4, $5)
		ORDER BY e.created_at ASC, e.id ASC
	`, chatterID, from, to, entity.UserActivityChatOnline, entity.UserActivityChatOffline)
	if err != nil {
		r.obs.LogError(ctx, span, "list timeline events failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []entity.UserActivityEvent

	for rows.Next() {
		var (
			e          entity.UserActivityEvent
			ch         sql.NullInt64
			chLogin    sql.NullString
			detailsRaw []byte
		)

		if err := rows.Scan(&e.ID, &e.ChatterTwitchUserID, &e.EventType, &ch, &chLogin, &detailsRaw, &e.CreatedAt); err != nil {
			return nil, err
		}

		if ch.Valid {
			v := ch.Int64
			e.ChannelTwitchUserID = &v
		}

		if chLogin.Valid {
			e.ChannelLogin = chLogin.String
		}

		if len(detailsRaw) > 0 {
			if err := json.Unmarshal(detailsRaw, &e.Details); err != nil {
				r.obs.Zap().Warn("unmarshal user_activity_events.details failed", zap.Error(err),
					zap.Int64("event_id", e.ID), zap.Int64("chatter_twitch_user_id", e.ChatterTwitchUserID))
			}
		}

		out = append(out, e)
	}

	return out, rows.Err()
}

// UpsertHelixMeta sets account creation time from Helix.
func (r *Repository) UpsertHelixMeta(ctx context.Context, twitchUserID int64, accountCreatedAt *time.Time, fetchedAt time.Time) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.upsert_helix_meta")
	defer span.End()

	var ca sql.NullTime
	if accountCreatedAt != nil {
		ca = sql.NullTime{Time: *accountCreatedAt, Valid: true}
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO twitch_user_helix_meta (twitch_user_id, account_created_at, helix_fetched_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (twitch_user_id) DO UPDATE SET
			account_created_at = COALESCE(EXCLUDED.account_created_at, twitch_user_helix_meta.account_created_at),
			helix_fetched_at = EXCLUDED.helix_fetched_at
	`, twitchUserID, ca, fetchedAt)
	if err != nil {
		r.obs.LogError(ctx, span, "upsert helix meta failed", err)
	}

	return err
}

// GetHelixMeta returns helix enrichment for a user if present.
func (r *Repository) GetHelixMeta(ctx context.Context, twitchUserID int64) (accountCreatedAt *time.Time, helixFetchedAt *time.Time, err error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_helix_meta")
	defer span.End()

	var ca, hf sql.NullTime

	err = r.pool.QueryRow(ctx, `
		SELECT account_created_at, helix_fetched_at FROM twitch_user_helix_meta WHERE twitch_user_id = $1
	`, twitchUserID).Scan(&ca, &hf)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil
	}

	if err != nil {
		r.obs.LogError(ctx, span, "get helix meta failed", err)
		return nil, nil, err
	}

	if ca.Valid {
		t := ca.Time
		accountCreatedAt = &t
	}

	if hf.Valid {
		t := hf.Time
		helixFetchedAt = &t
	}

	return accountCreatedAt, helixFetchedAt, nil
}

// UpsertChannelFollow stores follow time for a chatter/channel pair.
func (r *Repository) UpsertChannelFollow(ctx context.Context, chatterID, channelID int64, followedAt *time.Time, checkedAt time.Time) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.upsert_channel_follow")
	defer span.End()

	var fa sql.NullTime
	if followedAt != nil {
		fa = sql.NullTime{Time: *followedAt, Valid: true}
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO twitch_user_channel_follows (chatter_twitch_user_id, channel_twitch_user_id, followed_at, last_checked_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chatter_twitch_user_id, channel_twitch_user_id) DO UPDATE SET
			followed_at = COALESCE(EXCLUDED.followed_at, twitch_user_channel_follows.followed_at),
			last_checked_at = EXCLUDED.last_checked_at
	`, chatterID, channelID, fa, checkedAt)
	if err != nil {
		r.obs.LogError(ctx, span, "upsert channel follow failed", err)
	}

	return err
}

// ListFollowedMonitoredChannels returns follow rows joined with channel login for profile.
func (r *Repository) ListFollowedMonitoredChannels(ctx context.Context, chatterID int64) ([]entity.FollowedMonitoredChannel, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_followed_monitored")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT f.channel_twitch_user_id, u.username, f.followed_at
		FROM twitch_user_channel_follows f
		INNER JOIN twitch_users u ON u.id = f.channel_twitch_user_id
		WHERE f.chatter_twitch_user_id = $1 AND u.monitored = true
		ORDER BY u.username ASC
	`, chatterID)
	if err != nil {
		r.obs.LogError(ctx, span, "list followed monitored failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []entity.FollowedMonitoredChannel

	for rows.Next() {
		var (
			f  entity.FollowedMonitoredChannel
			fa sql.NullTime
		)

		if err := rows.Scan(&f.ChannelTwitchUserID, &f.ChannelLogin, &fa); err != nil {
			return nil, err
		}

		if fa.Valid {
			t := fa.Time
			f.FollowedAt = &t
		}

		out = append(out, f)
	}

	return out, rows.Err()
}

// ListDistinctChattersWithMessages returns twitch user ids that have at least one chat message as chatter.
func (r *Repository) ListDistinctChattersWithMessages(ctx context.Context, limit int) ([]int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_distinct_chatters")
	defer span.End()

	if limit < 1 {
		limit = 500
	}

	if limit > 5000 {
		limit = 5000
	}

	rows, err := r.pool.Query(ctx, `
		SELECT m.chatter_twitch_user_id
		FROM chat_messages m
		WHERE m.chatter_twitch_user_id IS NOT NULL
		GROUP BY m.chatter_twitch_user_id
		ORDER BY m.chatter_twitch_user_id
		LIMIT $1
	`, limit)
	if err != nil {
		r.obs.LogError(ctx, span, "list distinct chatters failed", err)
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

// ListChatterChannelPairsForFollowEnrichment returns (chatter, channel) pairs that appear in chat_messages for monitored channels.
func (r *Repository) ListChatterChannelPairsForFollowEnrichment(ctx context.Context, limit int) ([]entity.ChatterChannelPair, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_chatter_channel_pairs")
	defer span.End()

	if limit < 1 {
		limit = 200
	}

	if limit > 2000 {
		limit = 2000
	}

	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT m.chatter_twitch_user_id, m.twitch_user_id
		FROM chat_messages m
		INNER JOIN twitch_users u ON u.id = m.twitch_user_id AND u.monitored = true
		WHERE m.chatter_twitch_user_id IS NOT NULL
		LIMIT $1
	`, limit)
	if err != nil {
		r.obs.LogError(ctx, span, "list chatter channel pairs failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []entity.ChatterChannelPair

	for rows.Next() {
		var p entity.ChatterChannelPair
		if err := rows.Scan(&p.ChatterID, &p.ChannelID); err != nil {
			return nil, err
		}

		out = append(out, p)
	}

	return out, rows.Err()
}

// ListUserActivityEventsForChannelPresence returns join/part events for all chatters in a channel window (chronological).
func (r *Repository) ListUserActivityEventsForChannelPresence(ctx context.Context, channelTwitchUserID int64, from, to time.Time) ([]entity.UserActivityEvent, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_user_activity_channel_presence")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.chatter_twitch_user_id, e.event_type, e.channel_twitch_user_id, uc.username, e.details, e.created_at
		FROM user_activity_events e
		LEFT JOIN twitch_users uc ON uc.id = e.channel_twitch_user_id
		WHERE e.channel_twitch_user_id = $1
		  AND e.created_at >= $2 AND e.created_at <= $3
		  AND e.event_type IN ($4, $5)
		ORDER BY e.created_at ASC, e.id ASC
	`, channelTwitchUserID, from, to, entity.UserActivityChatOnline, entity.UserActivityChatOffline)
	if err != nil {
		r.obs.LogError(ctx, span, "list channel presence events failed", err)
		return nil, err
	}
	defer rows.Close()

	return r.scanUserActivityEventRows(rows)
}

// ListUserActivityForStream lists non-message activity for a channel in a time window (newest first).
func (r *Repository) ListUserActivityForStream(ctx context.Context, f entity.UserActivityListFilterForStream) ([]entity.UserActivityEvent, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_user_activity_for_stream")
	defer span.End()

	limit := f.Limit
	if limit < 1 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	q := `
		SELECT e.id, e.chatter_twitch_user_id, COALESCE(ch.username, ''), e.event_type, e.channel_twitch_user_id, uc.username, e.details, e.created_at
		FROM user_activity_events e
		LEFT JOIN twitch_users ch ON ch.id = e.chatter_twitch_user_id
		LEFT JOIN twitch_users uc ON uc.id = e.channel_twitch_user_id
		WHERE e.channel_twitch_user_id = $1
		  AND e.created_at >= $2 AND e.created_at <= $3
		  AND e.event_type <> 'message'
	`
	args := []any{f.ChannelTwitchUserID, f.From, f.To}
	n := 4

	if f.CursorCreatedAt != nil && f.CursorID != nil {
		q += ` AND (e.created_at, e.id) < ($` + strconv.Itoa(n) + `, $` + strconv.Itoa(n+1) + `)`

		args = append(args, *f.CursorCreatedAt, *f.CursorID)
		n += 2
	}

	q += ` ORDER BY e.created_at DESC, e.id DESC LIMIT $` + strconv.Itoa(n)

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		r.obs.LogError(ctx, span, "list user activity for stream failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.UserActivityEvent, 0, 64)

	for rows.Next() {
		var (
			e          entity.UserActivityEvent
			ch         sql.NullInt64
			chLogin    sql.NullString
			detailsRaw []byte
		)

		if err := rows.Scan(&e.ID, &e.ChatterTwitchUserID, &e.ChatterLogin, &e.EventType, &ch, &chLogin, &detailsRaw, &e.CreatedAt); err != nil {
			return nil, err
		}

		if ch.Valid {
			v := ch.Int64
			e.ChannelTwitchUserID = &v
		}

		if chLogin.Valid {
			e.ChannelLogin = chLogin.String
		}

		if len(detailsRaw) > 0 {
			if err := json.Unmarshal(detailsRaw, &e.Details); err != nil {
				r.obs.Zap().Warn("unmarshal user_activity_events.details failed", zap.Error(err),
					zap.Int64("event_id", e.ID), zap.Int64("chatter_twitch_user_id", e.ChatterTwitchUserID))
			}
		}

		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *Repository) scanUserActivityEventRows(rows pgx.Rows) ([]entity.UserActivityEvent, error) {
	out := make([]entity.UserActivityEvent, 0, 64)

	for rows.Next() {
		var (
			e          entity.UserActivityEvent
			ch         sql.NullInt64
			chLogin    sql.NullString
			detailsRaw []byte
		)

		if err := rows.Scan(&e.ID, &e.ChatterTwitchUserID, &e.EventType, &ch, &chLogin, &detailsRaw, &e.CreatedAt); err != nil {
			return nil, err
		}

		if ch.Valid {
			v := ch.Int64
			e.ChannelTwitchUserID = &v
		}

		if chLogin.Valid {
			e.ChannelLogin = chLogin.String
		}

		if len(detailsRaw) > 0 {
			if err := json.Unmarshal(detailsRaw, &e.Details); err != nil {
				r.obs.Zap().Warn("unmarshal user_activity_events.details failed", zap.Error(err),
					zap.Int64("event_id", e.ID), zap.Int64("chatter_twitch_user_id", e.ChatterTwitchUserID))
			}
		}

		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
