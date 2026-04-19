package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListChatHistory returns persisted messages for a monitored channel (oldest first).
func (s *Usecase) ListChatHistory(ctx context.Context, channel string, limit int) ([]entity.ChatHistoryMessage, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_chat_history")
	defer span.End()

	ok, err := s.repo.IsMonitoredChannel(ctx, channel)
	if err != nil {
		s.obs.LogError(ctx, span, "check monitored channel failed", err)
		return nil, err
	}

	if !ok {
		return nil, ErrChannelNotMonitored
	}

	list, err := s.repo.ListChatHistory(ctx, channel, limit)
	if err != nil {
		s.obs.LogError(ctx, span, "list chat history failed", err)
		return nil, err
	}

	return list, nil
}
