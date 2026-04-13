package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

func scanTwitchUser(scanner interface {
	Scan(dest ...any) error
}) (entity.TwitchUser, error) {
	var (
		u                entity.TwitchUser
		susType, susDesc sql.NullString
	)

	err := scanner.Scan(
		&u.ID, &u.Username, &u.Monitored, &u.Marked,
		&u.IsSus, &susType, &susDesc, &u.SusAutoSuppressed,
		&u.IrcOnlyWhenLive, &u.NotifyOffStreamMessages, &u.NotifyStreamStart,
	)
	if err != nil {
		return u, err
	}

	if susType.Valid {
		s := susType.String
		u.SusType = &s
	}

	if susDesc.Valid {
		s := susDesc.String
		u.SusDescription = &s
	}

	return u, nil
}

func (r *Repository) ListTwitchUsers(ctx context.Context) ([]entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_twitch_users")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
			irc_only_when_live, notify_off_stream_messages, notify_stream_start
		FROM twitch_users ORDER BY id`)
	if err != nil {
		r.obs.LogError(ctx, span, "list twitch users query failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.TwitchUser, 0)

	for rows.Next() {
		u, err := scanTwitchUser(rows)
		if err != nil {
			r.obs.LogError(ctx, span, "scan twitch user failed", err)
			return nil, err
		}

		out = append(out, u)
	}

	err = rows.Err()
	if err != nil {
		r.obs.LogError(ctx, span, "rows iteration failed", err)
	}

	return out, err
}

func (r *Repository) ListMonitoredTwitchUsers(ctx context.Context) ([]entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_monitored_twitch_users")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
			irc_only_when_live, notify_off_stream_messages, notify_stream_start
		FROM twitch_users WHERE monitored = true ORDER BY id
	`)
	if err != nil {
		r.obs.LogError(ctx, span, "list monitored twitch users query failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.TwitchUser, 0)

	for rows.Next() {
		u, err := scanTwitchUser(rows)
		if err != nil {
			r.obs.LogError(ctx, span, "scan twitch user failed", err)
			return nil, err
		}

		out = append(out, u)
	}

	err = rows.Err()
	if err != nil {
		r.obs.LogError(ctx, span, "rows iteration failed", err)
	}

	return out, err
}

func (r *Repository) CreateTwitchUser(ctx context.Context, id int64, username string) (entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.create_twitch_user")
	defer span.End()

	username = normalizeStoredUsername(username)

	u, err := scanTwitchUser(r.pool.QueryRow(ctx, `
		INSERT INTO twitch_users (id, username, monitored) VALUES ($1, $2, true)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			monitored = true
		RETURNING id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
			irc_only_when_live, notify_off_stream_messages, notify_stream_start
	`, id, username))
	if err != nil {
		r.obs.LogError(ctx, span, "create twitch user failed", err,
			zap.Int64("id", id), zap.String("username", username))
	}

	return u, err
}

func (r *Repository) SetTwitchUserMonitored(ctx context.Context, id int64, monitored bool) (entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.set_twitch_user_monitored")
	defer span.End()

	u, err := scanTwitchUser(r.pool.QueryRow(ctx, `
		UPDATE twitch_users SET monitored = $2 WHERE id = $1
		RETURNING id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
			irc_only_when_live, notify_off_stream_messages, notify_stream_start
	`, id, monitored))
	if err != nil {
		r.obs.LogError(ctx, span, "set twitch user monitored failed", err, zap.Int64("id", id))

		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TwitchUser{}, entity.ErrTwitchUserNotFound
		}
		return entity.TwitchUser{}, err
	}

	return u, nil
}

// PatchTwitchUser updates any non-nil fields in patch.
func (r *Repository) PatchTwitchUser(ctx context.Context, id int64, patch entity.TwitchUserPatch) (entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.patch_twitch_user")
	defer span.End()

	if patch.Monitored == nil && patch.Marked == nil && patch.IsSus == nil &&
		patch.SusType == nil && patch.SusDescription == nil && patch.SusAutoSuppressed == nil &&
		patch.IrcOnlyWhenLive == nil && patch.NotifyOffStreamMessages == nil && patch.NotifyStreamStart == nil {
		return r.GetTwitchUserByID(ctx, id)
	}

	setParts := make([]string, 0, 12)
	args := make([]any, 0, 16)

	argN := 1
	if patch.Monitored != nil {
		setParts = append(setParts, fmt.Sprintf("monitored = $%d", argN))
		args = append(args, *patch.Monitored)
		argN++
	}

	if patch.Marked != nil {
		setParts = append(setParts, fmt.Sprintf("marked = $%d", argN))
		args = append(args, *patch.Marked)
		argN++
	}

	if patch.IsSus != nil {
		setParts = append(setParts, fmt.Sprintf("is_sus = $%d", argN))
		args = append(args, *patch.IsSus)
		argN++
	}

	if patch.SusType != nil {
		if *patch.SusType == "" {
			setParts = append(setParts, "sus_type = NULL")
		} else {
			setParts = append(setParts, fmt.Sprintf("sus_type = $%d", argN))
			args = append(args, *patch.SusType)
			argN++
		}
	}

	if patch.SusDescription != nil {
		if *patch.SusDescription == "" {
			setParts = append(setParts, "sus_description = NULL")
		} else {
			setParts = append(setParts, fmt.Sprintf("sus_description = $%d", argN))
			args = append(args, *patch.SusDescription)
			argN++
		}
	}

	if patch.SusAutoSuppressed != nil {
		setParts = append(setParts, fmt.Sprintf("sus_auto_suppressed = $%d", argN))
		args = append(args, *patch.SusAutoSuppressed)
		argN++
	}

	if patch.IrcOnlyWhenLive != nil {
		setParts = append(setParts, fmt.Sprintf("irc_only_when_live = $%d", argN))
		args = append(args, *patch.IrcOnlyWhenLive)
		argN++
	}

	if patch.NotifyOffStreamMessages != nil {
		setParts = append(setParts, fmt.Sprintf("notify_off_stream_messages = $%d", argN))
		args = append(args, *patch.NotifyOffStreamMessages)
		argN++
	}

	if patch.NotifyStreamStart != nil {
		setParts = append(setParts, fmt.Sprintf("notify_stream_start = $%d", argN))
		args = append(args, *patch.NotifyStreamStart)
		argN++
	}

	args = append(args, id)
	q := fmt.Sprintf(
		`UPDATE twitch_users SET %s WHERE id = $%d RETURNING id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
			irc_only_when_live, notify_off_stream_messages, notify_stream_start`,
		strings.Join(setParts, ", "), argN,
	)

	u, err := scanTwitchUser(r.pool.QueryRow(ctx, q, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TwitchUser{}, entity.ErrTwitchUserNotFound
		}

		r.obs.LogError(ctx, span, "patch twitch user failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, err
	}

	return u, nil
}

// UpsertTwitchUserFromChat ensures a row exists for IRC persistence (e.g. sent messages); does not change monitored.
// The bool is true when this call created a new row (best-effort; see EXISTS check).
func (r *Repository) UpsertTwitchUserFromChat(ctx context.Context, id int64, username string) (inserted bool, err error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.upsert_twitch_user_from_chat")
	defer span.End()

	username = normalizeStoredUsername(username)

	var existed bool
	if qerr := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM twitch_users WHERE id = $1)`, id).Scan(&existed); qerr != nil {
		r.obs.LogError(ctx, span, "upsert twitch user exists check failed", qerr, zap.Int64("id", id))
		return false, qerr
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO twitch_users (id, username, monitored) VALUES ($1, $2, false)
		ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username
	`, id, username)
	if err != nil {
		r.obs.LogError(ctx, span, "upsert twitch user from chat failed", err,
			zap.Int64("id", id), zap.String("username", username))
		return false, err
	}

	return !existed, nil
}

func (r *Repository) TwitchUserIDByUsername(ctx context.Context, username string) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.twitch_user_id_by_username")
	defer span.End()

	username = normalizeStoredUsername(username)
	if username == "" {
		return 0, entity.ErrNoTwitchUserForChannel
	}

	var id int64

	err := r.pool.QueryRow(ctx, `
		SELECT id FROM twitch_users WHERE username = $1 LIMIT 1
	`, username).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, entity.ErrNoTwitchUserForChannel
		}

		r.obs.LogError(ctx, span, "twitch user id by username failed", err, zap.String("username", username))
		return 0, err
	}

	return id, nil
}

// ListTwitchUsersBrowse returns twitch_users for the public directory UI.
func (r *Repository) ListTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) ([]entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_twitch_users_browse")
	defer span.End()

	limit := f.Limit
	if limit < 1 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	var b strings.Builder
	b.WriteString(`SELECT id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
		irc_only_when_live, notify_off_stream_messages, notify_stream_start FROM twitch_users WHERE 1=1`)

	args := make([]any, 0, 8)
	argN := 1

	if q := normalizeStoredUsername(f.Username); q != "" {
		b.WriteString(` AND username ILIKE $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, "%"+q+"%")
		argN++
	}

	if f.CursorID != nil && *f.CursorID > 0 {
		lastID := *f.CursorID
		b.WriteString(` AND id < $` + strconv.Itoa(argN))
		args = append(args, lastID)
		argN++
	}

	b.WriteString(` ORDER BY id DESC LIMIT $`)
	b.WriteString(strconv.Itoa(argN))

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, b.String(), args...)
	if err != nil {
		r.obs.LogError(ctx, span, "list twitch users browse failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.TwitchUser, 0, limit)

	for rows.Next() {
		u, err := scanTwitchUser(rows)
		if err != nil {
			r.obs.LogError(ctx, span, "scan twitch user browse failed", err)
			return nil, err
		}

		out = append(out, u)
	}

	if err := rows.Err(); err != nil {
		r.obs.LogError(ctx, span, "browse rows iteration failed", err)
		return nil, err
	}

	return out, nil
}

// CountTwitchUsersBrowse counts rows matching the same filters as ListTwitchUsersBrowse (without limit/cursor).
func (r *Repository) CountTwitchUsersBrowse(ctx context.Context, f entity.TwitchUserBrowseFilter) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_twitch_users_browse")
	defer span.End()

	var b strings.Builder
	b.WriteString(`SELECT count(*) FROM twitch_users WHERE 1=1`)

	args := make([]any, 0, 4)
	argN := 1

	if q := normalizeStoredUsername(f.Username); q != "" {
		b.WriteString(` AND username ILIKE $`)
		b.WriteString(strconv.Itoa(argN))

		args = append(args, "%"+q+"%")
	}

	var n int64

	err := r.pool.QueryRow(ctx, b.String(), args...).Scan(&n)
	if err != nil {
		r.obs.LogError(ctx, span, "count twitch users browse failed", err)
	}

	return n, err
}

// GetTwitchUserByID returns a single twitch_users row.
func (r *Repository) GetTwitchUserByID(ctx context.Context, id int64) (entity.TwitchUser, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_twitch_user_by_id")
	defer span.End()

	u, err := scanTwitchUser(r.pool.QueryRow(ctx, `
		SELECT id, username, monitored, marked, is_sus, sus_type, sus_description, sus_auto_suppressed,
			irc_only_when_live, notify_off_stream_messages, notify_stream_start
		FROM twitch_users WHERE id = $1
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TwitchUser{}, entity.ErrTwitchUserNotFound
		}

		r.obs.LogError(ctx, span, "get twitch user by id failed", err, zap.Int64("id", id))
	}

	return u, err
}

// IsTwitchUserMarked returns marked flag for id (false if not found).
func (r *Repository) IsTwitchUserMarked(ctx context.Context, id int64) (bool, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.is_twitch_user_marked")
	defer span.End()

	var m bool

	err := r.pool.QueryRow(ctx, `SELECT marked FROM twitch_users WHERE id = $1`, id).Scan(&m)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return m, nil
}

// IsTwitchUserSuspicious returns is_sus for id (false if not found).
func (r *Repository) IsTwitchUserSuspicious(ctx context.Context, id int64) (bool, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.is_twitch_user_suspicious")
	defer span.End()

	var s bool

	err := r.pool.QueryRow(ctx, `SELECT is_sus FROM twitch_users WHERE id = $1`, id).Scan(&s)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return s, nil
}

// CountChatMessagesByChatter counts persisted rows for a chatter.
func (r *Repository) CountChatMessagesByChatter(ctx context.Context, chatterTwitchUserID int64) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_chat_messages_by_chatter")
	defer span.End()

	var n int64

	err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM chat_messages WHERE chatter_twitch_user_id = $1
	`, chatterTwitchUserID).Scan(&n)
	if err != nil {
		r.obs.LogError(ctx, span, "count chat messages by chatter failed", err, zap.Int64("chatter_id", chatterTwitchUserID))
	}

	return n, err
}
