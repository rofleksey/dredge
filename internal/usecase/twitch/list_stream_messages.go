package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListStreamMessages lists chat messages tagged with this stream.
func (s *Service) ListStreamMessages(ctx context.Context, streamID int64, f entity.ChatMessageListFilter) ([]entity.ChatHistoryMessage, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_stream_messages")
	defer span.End()

	f.StreamID = &streamID

	return s.repo.ListChatMessages(ctx, f)
}
