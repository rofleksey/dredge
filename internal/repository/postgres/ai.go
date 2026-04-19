package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
)

func (r *Repository) GetAISettings(ctx context.Context) (entity.AISettings, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_ai_settings")
	defer span.End()

	var s entity.AISettings

	err := r.pool.QueryRow(ctx, `
		SELECT base_url, model, COALESCE(api_token, ''), updated_at
		FROM ai_settings WHERE id = 1
	`).Scan(&s.BaseURL, &s.Model, &s.APIToken, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.AISettings{}, nil
		}
		r.obs.LogError(ctx, span, "get ai settings failed", err)
		return entity.AISettings{}, err
	}

	return s, nil
}

func (r *Repository) UpsertAISettings(ctx context.Context, s entity.AISettings) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.upsert_ai_settings")
	defer span.End()

	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_settings (id, base_url, model, api_token, updated_at)
		VALUES (1, $1, $2, $3, NOW())
		ON CONFLICT (id) DO UPDATE SET
			base_url = EXCLUDED.base_url,
			model = EXCLUDED.model,
			api_token = EXCLUDED.api_token,
			updated_at = NOW()
	`, s.BaseURL, s.Model, s.APIToken)
	if err != nil {
		r.obs.LogError(ctx, span, "upsert ai settings failed", err)
	}

	return err
}

func (r *Repository) ListAIConversations(ctx context.Context) ([]entity.AIConversation, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_ai_conversations")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, title, created_at, updated_at
		FROM ai_conversations
		ORDER BY updated_at DESC
	`)
	if err != nil {
		r.obs.LogError(ctx, span, "list ai conversations failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []entity.AIConversation

	for rows.Next() {
		var c entity.AIConversation
		if err := rows.Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}

	return out, rows.Err()
}

func (r *Repository) CreateAIConversation(ctx context.Context, title *string) (entity.AIConversation, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.create_ai_conversation")
	defer span.End()

	var c entity.AIConversation

	err := r.pool.QueryRow(ctx, `
		INSERT INTO ai_conversations (title) VALUES ($1)
		RETURNING id, title, created_at, updated_at
	`, title).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		r.obs.LogError(ctx, span, "create ai conversation failed", err)
		return entity.AIConversation{}, err
	}

	return c, nil
}

func (r *Repository) GetAIConversation(ctx context.Context, id int64) (entity.AIConversation, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.get_ai_conversation")
	defer span.End()

	var c entity.AIConversation

	err := r.pool.QueryRow(ctx, `
		SELECT id, title, created_at, updated_at FROM ai_conversations WHERE id = $1
	`, id).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.AIConversation{}, pgx.ErrNoRows
		}
		r.obs.LogError(ctx, span, "get ai conversation failed", err)
		return entity.AIConversation{}, err
	}

	return c, nil
}

func (r *Repository) DeleteAIConversation(ctx context.Context, id int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.delete_ai_conversation")
	defer span.End()

	tag, err := r.pool.Exec(ctx, `DELETE FROM ai_conversations WHERE id = $1`, id)
	if err != nil {
		r.obs.LogError(ctx, span, "delete ai conversation failed", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *Repository) TouchAIConversation(ctx context.Context, id int64) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.touch_ai_conversation")
	defer span.End()

	_, err := r.pool.Exec(ctx, `UPDATE ai_conversations SET updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		r.obs.LogError(ctx, span, "touch ai conversation failed", err)
	}

	return err
}

func (r *Repository) ListAIMessages(ctx context.Context, conversationID int64) ([]entity.AIMessage, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.list_ai_messages")
	defer span.End()

	rows, err := r.pool.Query(ctx, `
		SELECT id, conversation_id, role, content, metadata, created_at
		FROM ai_messages
		WHERE conversation_id = $1
		ORDER BY id ASC
	`, conversationID)
	if err != nil {
		r.obs.LogError(ctx, span, "list ai messages failed", err)
		return nil, err
	}
	defer rows.Close()

	var out []entity.AIMessage

	for rows.Next() {
		var (
			m       entity.AIMessage
			metaRaw []byte
			role    string
		)
		if err := rows.Scan(&m.ID, &m.ConversationID, &role, &m.Content, &metaRaw, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Role = entity.AIMessageRole(role)
		meta, err := entity.ParseAIMessageMetadata(metaRaw)
		if err != nil {
			return nil, err
		}
		m.Metadata = meta
		out = append(out, m)
	}

	return out, rows.Err()
}

func (r *Repository) InsertAIMessage(ctx context.Context, m entity.AIMessage) (entity.AIMessage, error) {
	ctx, span := r.obs.StartSpan(ctx, "repo.insert_ai_message")
	defer span.End()

	meta, err := m.MetadataJSON()
	if err != nil {
		return entity.AIMessage{}, err
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO ai_messages (conversation_id, role, content, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, m.ConversationID, string(m.Role), m.Content, meta).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		r.obs.LogError(ctx, span, "insert ai message failed", err)
		return entity.AIMessage{}, err
	}

	return m, nil
}

// SetAIMessageMetadata updates JSON metadata for a message (e.g. after tool results).
func (r *Repository) SetAIMessageMetadata(ctx context.Context, messageID int64, metadata map[string]any) error {
	ctx, span := r.obs.StartSpan(ctx, "repo.set_ai_message_metadata")
	defer span.End()

	b, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `UPDATE ai_messages SET metadata = $1 WHERE id = $2`, b, messageID)
	if err != nil {
		r.obs.LogError(ctx, span, "set ai message metadata failed", err)
	}

	return err
}
