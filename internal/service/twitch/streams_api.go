package twitch

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rofleksey/dredge/internal/entity"
)

// ListMonitoredStreams lists recorded streams for monitored channels.
func (s *Service) ListMonitoredStreams(ctx context.Context, f entity.StreamListFilter) ([]entity.Stream, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_monitored_streams")
	defer span.End()

	return s.repo.ListMonitoredStreams(ctx, f)
}

// GetMonitoredStream returns a stream row if the channel is monitored.
func (s *Service) GetMonitoredStream(ctx context.Context, id int64) (entity.Stream, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.get_monitored_stream")
	defer span.End()

	st, err := s.repo.GetMonitoredStreamByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Stream{}, entity.ErrStreamNotFound
	}

	return st, err
}

// ListStreamMessages lists chat messages tagged with this stream.
func (s *Service) ListStreamMessages(ctx context.Context, streamID int64, f entity.ChatMessageListFilter) ([]entity.ChatHistoryMessage, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_stream_messages")
	defer span.End()

	f.StreamID = &streamID

	return s.repo.ListChatMessages(ctx, f)
}

// ListStreamActivity lists non-message activity in the stream time window.
func (s *Service) ListStreamActivity(ctx context.Context, streamID int64, limit int, cursorCreatedAt *time.Time, cursorID *int64) ([]entity.UserActivityEvent, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_stream_activity")
	defer span.End()

	st, err := s.repo.GetMonitoredStreamByID(ctx, streamID)
	if err != nil {
		return nil, err
	}

	streamEnd := time.Now().UTC()
	if st.EndedAt != nil {
		streamEnd = st.EndedAt.UTC()
	}

	f := entity.UserActivityListFilterForStream{
		ChannelTwitchUserID: st.ChannelTwitchUserID,
		From:                st.StartedAt.UTC(),
		To:                  streamEnd,
		Limit:               limit,
		CursorCreatedAt:     cursorCreatedAt,
		CursorID:            cursorID,
	}

	return s.repo.ListUserActivityForStream(ctx, f)
}
