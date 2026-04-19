package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// CountChatMessages delegates to repository (same filters as list, no cursor).
func (s *Service) CountChatMessages(ctx context.Context, f entity.ChatMessageListFilter) (int64, error) {
	return s.repo.CountChatMessages(ctx, f)
}
