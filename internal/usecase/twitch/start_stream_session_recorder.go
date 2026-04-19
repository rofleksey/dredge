package twitch

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// StartStreamSessionRecorder polls Helix on an interval and persists open/closed stream rows for monitored channels.
// Live metadata is refreshed via batched GET /helix/streams (see helix.Client.HelixStreamsMetadataByBroadcasterIDs, up to 100 user_ids per request).
func (s *Usecase) StartStreamSessionRecorder(ctx context.Context) {
	d := s.streamSessionPollInterval
	if d <= 0 {
		d = 10 * time.Second
	}

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			syncCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

			if err := s.syncStreamSessions(syncCtx); err != nil {
				s.obs.Logger.Warn("sync stream sessions failed", zap.Error(err))
			}

			cancel()
		}
	}
}
