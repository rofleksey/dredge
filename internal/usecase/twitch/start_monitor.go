package twitch

import "context"

func (s *Service) StartMonitor(ctx context.Context) error {
	return s.live.StartMonitor(ctx)
}
