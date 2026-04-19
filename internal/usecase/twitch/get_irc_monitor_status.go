package twitch

import "context"

func (s *Usecase) GetIrcMonitorStatus(ctx context.Context) (connected bool, channels []IRCMonitorChannelStatus, err error) {
	return s.live.GetIrcMonitorStatus(ctx)
}
