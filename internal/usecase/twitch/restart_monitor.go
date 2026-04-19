package twitch

import "context"

func (s *Usecase) RestartMonitor(ctx context.Context) error {
	return s.live.RestartMonitor(ctx)
}
