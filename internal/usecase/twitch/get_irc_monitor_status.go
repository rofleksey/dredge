package twitch

import "context"

func (s *Service) GetIrcMonitorStatus(ctx context.Context) (connected bool, channels []IRCMonitorChannelStatus, err error) {
	return s.live.GetIrcMonitorStatus(ctx)
}
