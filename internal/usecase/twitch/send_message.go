package twitch

import "context"

func (s *Service) SendMessage(ctx context.Context, accountID int64, channel, message string) error {
	return s.live.SendMessage(ctx, accountID, channel, message)
}
