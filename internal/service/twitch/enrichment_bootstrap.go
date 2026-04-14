package twitch

import (
	"context"

	"go.uber.org/zap"
)

// EnqueueMonitoredAndMarkedUsersForEnrichment pushes monitored/marked users to the enrichment queue.
func (s *Service) EnqueueMonitoredAndMarkedUsersForEnrichment(ctx context.Context) {
	ids, err := s.repo.ListMonitoredOrMarkedTwitchUserIDs(ctx)
	if err != nil {
		s.obs.Logger.Warn("enqueue startup enrichment users failed", zap.Error(err))
		return
	}

	for _, id := range ids {
		s.EnqueueUserEnrichment(id)
	}
}
