package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

func normalizeChannelName(ch string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ch)), "#")
}

// InsertChatMessage stores a chat line for history and replay in the UI.
func (r *Repository) InsertChatMessage(ctx context.Context, channelTwitchUserID int64, chatterTwitchUserID *int64, chatterUsername, body string, keywordMatch bool, msgType string, badgeTags []string) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.insert_chat_message")
	defer span.End()

	if channelTwitchUserID == 0 || msgType == "" {
		return 0, errors.New("invalid chat message insert")
	}

	chatterUsername = normalizeStoredUsername(chatterUsername)
	if chatterUsername == "" {
		return 0, errors.New("invalid chat message insert")
	}

	if badgeTags == nil {
		badgeTags = []string{}
	}

	badgeJSON, err := json.Marshal(badgeTags)
	if err != nil {
		return 0, err
	}

	var chatter sql.NullInt64
	if chatterTwitchUserID != nil && *chatterTwitchUserID != 0 {
		chatter = sql.NullInt64{Int64: *chatterTwitchUserID, Valid: true}
	}

	var stream sql.NullInt64
	if sid, err := r.ActiveStreamIDForChannel(ctx, channelTwitchUserID); err == nil && sid != nil {
		stream = sql.NullInt64{Int64: *sid, Valid: true}
	}

	var msgID int64

	err = r.pool.QueryRow(ctx, `
		INSERT INTO chat_messages (twitch_user_id, chatter_twitch_user_id, username, body, keyword_match, msg_type, badge_tags, stream_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8)
		RETURNING id
	`, channelTwitchUserID, chatter, chatterUsername, body, keywordMatch, msgType, badgeJSON, stream).Scan(&msgID)
	if err != nil {
		r.obs.LogError(ctx, span, "insert chat message failed", err,
			zap.Int64("twitch_user_id", channelTwitchUserID), zap.String("username", chatterUsername))
	}

	return msgID, err
}

// InsertChatMessageForChannelLogin resolves the channel by username (must exist).
func (r *Repository) InsertChatMessageForChannelLogin(ctx context.Context, channelLogin string, chatterTwitchUserID *int64, chatterUsername, body string, keywordMatch bool, msgType string, badgeTags []string) (int64, error) {
	ch := normalizeChannelName(channelLogin)
	if ch == "" {
		return 0, errors.New("invalid chat message insert")
	}

	id, err := r.TwitchUserIDByUsername(ctx, ch)
	if err != nil {
		return 0, err
	}

	return r.InsertChatMessage(ctx, id, chatterTwitchUserID, chatterUsername, body, keywordMatch, msgType, badgeTags)
}

// IsMonitoredChannel reports whether the normalized channel username is monitored.
func (r *Repository) IsMonitoredChannel(ctx context.Context, channel string) (bool, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.is_monitored_channel")
	defer span.End()

	ch := normalizeChannelName(channel)
	if ch == "" {
		return false, nil
	}

	var one int

	err := r.pool.QueryRow(ctx, `
		SELECT 1 FROM twitch_users
		WHERE lower(username) = lower($1) AND monitored = true
		LIMIT 1
	`, ch).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		r.obs.LogError(ctx, span, "is monitored channel query failed", err, zap.String("channel", ch))
		return false, err
	}

	return true, nil
}

// MonitoredChannelTwitchUserID returns twitch_users.id when the channel login is monitored.
func (r *Repository) MonitoredChannelTwitchUserID(ctx context.Context, channel string) (id int64, ok bool, err error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.monitored_channel_twitch_user_id")
	defer span.End()

	ch := normalizeChannelName(channel)
	if ch == "" {
		return 0, false, nil
	}

	err = r.pool.QueryRow(ctx, `
		SELECT id FROM twitch_users
		WHERE lower(username) = lower($1) AND monitored = true
		LIMIT 1
	`, ch).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}

	if err != nil {
		r.obs.LogError(ctx, span, "monitored channel twitch user id query failed", err, zap.String("channel", ch))
		return 0, false, err
	}

	return id, true, nil
}

func scanChatHistoryRow(rows pgx.Rows) (entity.ChatHistoryMessage, error) {
	var m entity.ChatHistoryMessage

	var badgeRaw []byte

	var (
		chatter       sql.NullInt64
		chatterMarked sql.NullBool
		chatterIsSus  sql.NullBool
	)

	err := rows.Scan(&m.ID, &m.Channel, &m.Username, &chatter, &chatterMarked, &chatterIsSus, &m.Message, &m.KeywordMatch, &m.MsgType, &badgeRaw, &m.CreatedAt)
	if err != nil {
		return m, err
	}

	if chatter.Valid {
		v := chatter.Int64
		m.ChatterTwitchUserID = &v
	}

	if chatterMarked.Valid {
		m.ChatterMarked = chatterMarked.Bool
	}

	if chatterIsSus.Valid {
		m.ChatterIsSus = chatterIsSus.Bool
	}

	if len(badgeRaw) > 0 {
		_ = json.Unmarshal(badgeRaw, &m.BadgeTags)
	}

	return m, nil
}

// ListChatHistory returns the most recent messages for a channel, oldest first.
func (r *Repository) ListChatHistory(ctx context.Context, channel string, limit int) ([]entity.ChatHistoryMessage, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_chat_history")
	defer span.End()

	ch := normalizeChannelName(channel)
	if ch == "" {
		return nil, nil
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 200 {
		limit = 200
	}

	rows, err := r.pool.Query(ctx, `
		SELECT m.id, u.username, m.username, m.chatter_twitch_user_id, COALESCE(cu.marked, false), COALESCE(cu.is_sus, false), m.body, m.keyword_match, m.msg_type, m.badge_tags, m.created_at
		FROM (
			SELECT m.id
			FROM chat_messages m
			INNER JOIN twitch_users u ON u.id = m.twitch_user_id AND lower(u.username) = lower($1)
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT $2
		) t
		INNER JOIN chat_messages m ON m.id = t.id
		INNER JOIN twitch_users u ON u.id = m.twitch_user_id
		LEFT JOIN twitch_users cu ON cu.id = m.chatter_twitch_user_id
		ORDER BY m.created_at ASC, m.id ASC
	`, ch, limit)
	if err != nil {
		r.obs.LogError(ctx, span, "list chat history query failed", err, zap.String("channel", ch))
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.ChatHistoryMessage, 0, limit)

	for rows.Next() {
		m, err := scanChatHistoryRow(rows)
		if err != nil {
			r.obs.LogError(ctx, span, "scan chat history row failed", err)
			return nil, err
		}

		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		r.obs.LogError(ctx, span, "chat history rows iteration failed", err)
		return nil, err
	}

	return out, nil
}

// ListChatMessages returns messages matching filters, newest first.
func (r *Repository) ListChatMessages(ctx context.Context, f entity.ChatMessageListFilter) ([]entity.ChatHistoryMessage, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_chat_messages")
	defer span.End()

	limit := f.Limit
	if limit < 1 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	var b strings.Builder
	b.WriteString(`
		SELECT m.id, uc.username, m.username, m.chatter_twitch_user_id, COALESCE(cu.marked, false), COALESCE(cu.is_sus, false), m.body, m.keyword_match, m.msg_type, m.badge_tags, m.created_at
		FROM chat_messages m
		INNER JOIN twitch_users uc ON uc.id = m.twitch_user_id
		LEFT JOIN twitch_users cu ON cu.id = m.chatter_twitch_user_id
		WHERE 1=1
	`)

	args := make([]any, 0, 16)
	argN := 1

	if f.Username != "" {
		b.WriteString(` AND m.username ILIKE $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, "%"+normalizeStoredUsername(f.Username)+"%")
		argN++
	}

	if f.Text != "" {
		b.WriteString(` AND m.body ILIKE $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, "%"+f.Text+"%")
		argN++
	}

	if ch := normalizeChannelName(f.Channel); ch != "" {
		b.WriteString(` AND lower(uc.username) = lower($`)
		b.WriteString(strconv.Itoa(argN))
		b.WriteString(`)`)

		args = append(args, ch)
		argN++
	}

	if f.CreatedFrom != nil {
		b.WriteString(` AND m.created_at >= $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.CreatedFrom)
		argN++
	}

	if f.CreatedTo != nil {
		b.WriteString(` AND m.created_at <= $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.CreatedTo)
		argN++
	}

	if f.ChatterUserID != nil && *f.ChatterUserID != 0 {
		b.WriteString(` AND m.chatter_twitch_user_id = $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.ChatterUserID)
		argN++
	}

	if f.StreamID != nil && *f.StreamID != 0 {
		b.WriteString(` AND m.stream_id = $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.StreamID)
		argN++
	}

	if f.CursorCreatedAt != nil && f.CursorID != nil {
		b.WriteString(` AND (m.created_at, m.id) < ($`)
		b.WriteString(strconv.Itoa(argN))
		b.WriteString(`, $`)
		b.WriteString(strconv.Itoa(argN + 1))
		b.WriteString(`)`)

		args = append(args, *f.CursorCreatedAt, *f.CursorID)
		argN += 2
	}

	b.WriteString(` ORDER BY m.created_at DESC, m.id DESC LIMIT $`)
	b.WriteString(strconv.Itoa(argN))

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, b.String(), args...)
	if err != nil {
		r.obs.LogError(ctx, span, "list chat messages query failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.ChatHistoryMessage, 0, limit)

	for rows.Next() {
		m, err := scanChatHistoryRow(rows)
		if err != nil {
			r.obs.LogError(ctx, span, "scan chat message row failed", err)
			return nil, err
		}

		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		r.obs.LogError(ctx, span, "list chat messages iteration failed", err)
		return nil, err
	}

	return out, nil
}

// CountChatMessages returns the number of rows matching filters (ignores cursor/limit).
func (r *Repository) CountChatMessages(ctx context.Context, f entity.ChatMessageListFilter) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_chat_messages")
	defer span.End()

	var b strings.Builder
	b.WriteString(`
		SELECT count(*) FROM chat_messages m
		INNER JOIN twitch_users uc ON uc.id = m.twitch_user_id
		WHERE 1=1
	`)

	args := make([]any, 0, 16)
	argN := 1

	if f.Username != "" {
		b.WriteString(` AND m.username ILIKE $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, "%"+normalizeStoredUsername(f.Username)+"%")
		argN++
	}

	if f.Text != "" {
		b.WriteString(` AND m.body ILIKE $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, "%"+f.Text+"%")
		argN++
	}

	if ch := normalizeChannelName(f.Channel); ch != "" {
		b.WriteString(` AND lower(uc.username) = lower($`)
		b.WriteString(strconv.Itoa(argN))
		b.WriteString(`)`)

		args = append(args, ch)
		argN++
	}

	if f.CreatedFrom != nil {
		b.WriteString(` AND m.created_at >= $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.CreatedFrom)
		argN++
	}

	if f.CreatedTo != nil {
		b.WriteString(` AND m.created_at <= $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.CreatedTo)
		argN++
	}

	if f.ChatterUserID != nil && *f.ChatterUserID != 0 {
		b.WriteString(` AND m.chatter_twitch_user_id = $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.ChatterUserID)
		argN++
	}

	if f.StreamID != nil && *f.StreamID != 0 {
		b.WriteString(` AND m.stream_id = $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, *f.StreamID)
	}

	var n int64

	err := r.pool.QueryRow(ctx, b.String(), args...).Scan(&n)
	if err != nil {
		r.obs.LogError(ctx, span, "count chat messages failed", err)
	}

	return n, err
}
