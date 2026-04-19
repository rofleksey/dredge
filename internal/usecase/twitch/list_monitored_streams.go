package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListMonitoredStreams lists recorded streams for monitored channels.
func (s *Usecase) ListMonitoredStreams(ctx context.Context, f entity.StreamListFilter) ([]entity.Stream, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_monitored_streams")
	defer span.End()

	return s.repo.ListMonitoredStreams(ctx, f)
}
