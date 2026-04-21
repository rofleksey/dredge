package twitch

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// StartChannelDiscoveryLoop wakes periodically and runs RunChannelDiscovery when enabled and the poll interval has elapsed.
func (s *Usecase) StartChannelDiscoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	var lastRun time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			st, err := s.repo.GetChannelDiscoverySettings(ctx)
			if err != nil {
				s.obs.Logger.Warn("channel discovery settings tick load failed", zap.Error(err))
				continue
			}

			if !st.Enabled {
				continue
			}

			interval := time.Duration(st.PollIntervalSeconds) * time.Second
			if interval < time.Minute {
				interval = time.Minute
			}

			if !lastRun.IsZero() && time.Since(lastRun) < interval {
				continue
			}

			runCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

			runErr := s.RunChannelDiscovery(runCtx)

			cancel()

			if runErr != nil {
				s.obs.Logger.Warn("channel discovery run failed", zap.Error(runErr))
				continue
			}

			lastRun = time.Now()
		}
	}
}
