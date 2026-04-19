package twitch

import (
	"go.uber.org/zap"
)

// EnqueueUserEnrichment queues Helix meta fetch for a user (non-blocking; drops if full).
func (s *Usecase) EnqueueUserEnrichment(userID int64) {
	if s.enrichQueue == nil {
		return
	}

	select {
	case s.enrichQueue <- userID:
	default:
		s.obs.Logger.Warn("enrichment queue full", zap.Int64("user_id", userID))
	}
}
