package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

const twitchAccountOrder = `ORDER BY CASE account_type WHEN 'main' THEN 0 ELSE 1 END, created_at ASC, id ASC`

func (r *Repository) ListTwitchAccounts(ctx context.Context) ([]entity.TwitchAccount, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_twitch_accounts")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, username, refresh_token, account_type, created_at
		FROM twitch_accounts
		WHERE deleted_at IS NULL
	`+twitchAccountOrder)
	if err != nil {
		r.obs.LogError(ctx, span, "list twitch accounts query failed", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]entity.TwitchAccount, 0)

	for rows.Next() {
		var a entity.TwitchAccount
		if err := rows.Scan(&a.ID, &a.Username, &a.RefreshToken, &a.AccountType, &a.CreatedAt); err != nil {
			r.obs.LogError(ctx, span, "scan twitch account failed", err)
			return nil, err
		}

		out = append(out, a)
	}

	err = rows.Err()
	if err != nil {
		r.obs.LogError(ctx, span, "rows iteration failed", err)
	}

	return out, err
}

func (r *Repository) CountTwitchAccounts(ctx context.Context) (int64, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.count_twitch_accounts")
	defer span.End()

	var n int64

	err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM twitch_accounts WHERE deleted_at IS NULL
	`).Scan(&n)
	if err != nil {
		r.obs.LogError(ctx, span, "count twitch accounts failed", err)
	}

	return n, err
}

func (r *Repository) CreateTwitchAccount(ctx context.Context, id int64, username, refreshToken, accountType string) (entity.TwitchAccount, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.create_twitch_account")
	defer span.End()

	username = normalizeStoredUsername(username)

	var a entity.TwitchAccount

	err := r.pool.QueryRow(ctx, `
		INSERT INTO twitch_accounts (id, username, refresh_token, account_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, refresh_token, account_type, created_at
	`, id, username, refreshToken, accountType).Scan(&a.ID, &a.Username, &a.RefreshToken, &a.AccountType, &a.CreatedAt)
	if err != nil {
		r.obs.LogError(ctx, span, "create twitch account failed", err, zap.String("username", username), zap.Int64("twitch_user_id", id))
	}

	return a, err
}

func (r *Repository) GetTwitchAccountByID(ctx context.Context, id int64) (entity.TwitchAccount, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_twitch_account_by_id")
	defer span.End()

	var a entity.TwitchAccount

	err := r.pool.QueryRow(ctx, `
		SELECT id, username, refresh_token, account_type, created_at
		FROM twitch_accounts
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&a.ID, &a.Username, &a.RefreshToken, &a.AccountType, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TwitchAccount{}, entity.ErrTwitchAccountNotFound
		}

		r.obs.LogError(ctx, span, "get twitch account by id failed", err, zap.Int64("account_id", id))
	}

	return a, err
}

// GetTwitchAccountByTwitchUserID returns the linked OAuth row for that Twitch user id (same as primary key id).
func (r *Repository) GetTwitchAccountByTwitchUserID(ctx context.Context, twitchUserID int64) (entity.TwitchAccount, error) {
	return r.GetTwitchAccountByID(ctx, twitchUserID)
}

func (r *Repository) UpdateTwitchRefreshToken(ctx context.Context, id int64, refreshToken string) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.update_twitch_refresh_token")
	defer span.End()

	tag, err := r.pool.Exec(ctx, `
		UPDATE twitch_accounts SET refresh_token = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, id, refreshToken)
	if err != nil {
		r.obs.LogError(ctx, span, "update twitch refresh token failed", err, zap.Int64("account_id", id))
		return err
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrTwitchAccountNotFound
	}

	return nil
}

func (r *Repository) PatchTwitchAccount(ctx context.Context, id int64, accountType *string) (entity.TwitchAccount, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.patch_twitch_account")
	defer span.End()

	var accountTypeArg sql.NullString
	if accountType != nil {
		accountTypeArg = sql.NullString{String: *accountType, Valid: true}
	}

	var a entity.TwitchAccount

	err := r.pool.QueryRow(ctx, `
		UPDATE twitch_accounts SET
			account_type = CASE WHEN $2::text IS NULL THEN account_type ELSE $2 END
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, username, refresh_token, account_type, created_at
	`, id, accountTypeArg).Scan(&a.ID, &a.Username, &a.RefreshToken, &a.AccountType, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.TwitchAccount{}, entity.ErrTwitchAccountNotFound
		}

		r.obs.LogError(ctx, span, "patch twitch account failed", err, zap.Int64("id", id))
		return entity.TwitchAccount{}, err
	}

	return a, nil
}

func (r *Repository) DeleteTwitchAccount(ctx context.Context, id int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.delete_twitch_account")
	defer span.End()

	tag, err := r.pool.Exec(ctx, `
		UPDATE twitch_accounts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		r.obs.LogError(ctx, span, "delete twitch account failed", err, zap.Int64("id", id))
		return err
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrTwitchAccountNotFound
	}

	return nil
}
