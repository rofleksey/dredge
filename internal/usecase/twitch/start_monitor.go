package twitch

import "context"

func (s *Usecase) StartMonitor(ctx context.Context) error {
	return s.live.StartMonitor(ctx)
}
