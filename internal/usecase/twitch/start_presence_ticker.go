package twitch

import "context"

func (s *Service) StartPresenceTicker(ctx context.Context) {
	s.live.StartPresenceTicker(ctx)
}
