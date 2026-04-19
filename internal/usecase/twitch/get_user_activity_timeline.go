package twitch

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

// GetUserActivityTimeline returns merged presence segments in a window.
func (s *Usecase) GetUserActivityTimeline(ctx context.Context, chatterID int64, from, to time.Time) ([]entity.ActivityTimelineSegment, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.user_activity_timeline")
	defer span.End()

	if !from.Before(to) {
		from = to.Add(-7 * 24 * time.Hour)
	}

	ev, err := s.repo.ListUserActivityEventsForTimeline(ctx, chatterID, from, to)
	if err != nil {
		return nil, err
	}

	return BuildActivityTimelineSegments(ev, to), nil
}
