package ai

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListAIConversations returns conversations newest first.
func (u *Usecase) ListAIConversations(ctx context.Context) ([]entity.AIConversation, error) {
	return u.repo.ListAIConversations(ctx)
}

// CreateAIConversation creates a conversation with optional title.
func (u *Usecase) CreateAIConversation(ctx context.Context, title *string) (entity.AIConversation, error) {
	return u.repo.CreateAIConversation(ctx, title)
}

// DeleteAIConversation removes a conversation and messages.
func (u *Usecase) DeleteAIConversation(ctx context.Context, id int64) error {
	u.StopRun(id)
	return u.repo.DeleteAIConversation(ctx, id)
}

// GetAIConversation returns one conversation or ErrNotFound.
func (u *Usecase) GetAIConversation(ctx context.Context, id int64) (entity.AIConversation, error) {
	return u.repo.GetAIConversation(ctx, id)
}

// ListAIMessages returns messages for a conversation.
func (u *Usecase) ListAIMessages(ctx context.Context, conversationID int64) ([]entity.AIMessage, error) {
	return u.repo.ListAIMessages(ctx, conversationID)
}
