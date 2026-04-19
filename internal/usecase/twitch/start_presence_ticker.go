package twitch

import "context"

func (s *Usecase) StartPresenceTicker(ctx context.Context) {
	s.live.StartPresenceTicker(ctx)
}
