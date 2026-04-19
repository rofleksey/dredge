package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListChatMessages returns persisted messages matching filters (newest first).
func (s *Usecase) ListChatMessages(ctx context.Context, f entity.ChatMessageListFilter) ([]entity.ChatHistoryMessage, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_chat_messages")
	defer span.End()

	list, err := s.repo.ListChatMessages(ctx, f)
	if err != nil {
		s.obs.LogError(ctx, span, "list chat messages failed", err)
		return nil, err
	}

	return list, nil
}
