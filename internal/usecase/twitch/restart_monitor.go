package twitch

import "context"

func (s *Service) RestartMonitor(ctx context.Context) error {
	return s.live.RestartMonitor(ctx)
}
