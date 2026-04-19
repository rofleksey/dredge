package twitch

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListStreamActivity lists non-message activity in the stream time window.
func (s *Usecase) ListStreamActivity(ctx context.Context, streamID int64, limit int, cursorCreatedAt *time.Time, cursorID *int64) ([]entity.UserActivityEvent, error) {
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
