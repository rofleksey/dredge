package twitch

import (
	"context"
	"time"
)

// StartEnrichmentWorker drains enrichQueue with one worker until ctx is cancelled.
func (s *Service) StartEnrichmentWorker(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case uid := <-s.enrichQueue:
				runCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				s.enrichSingleUser(runCtx, uid)
				cancel()
			}
		}
	}()
}
