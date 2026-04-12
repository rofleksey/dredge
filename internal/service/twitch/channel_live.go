package twitch

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

// GetChannelLive returns Helix stream metadata plus the IRC-maintained chatter count when the DB query succeeds.
func (s *Service) GetChannelLive(ctx context.Context, login string) (info helix.ChannelLiveInfo, channelChatterCount *int64, err error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.get_channel_live")
	defer span.End()

	info, err = s.Client.GetChannelLive(ctx, login)
	if err != nil {
		return helix.ChannelLiveInfo{}, nil, err
	}

	n, err := s.repo.CountChannelChatters(ctx, info.BroadcasterID)
	if err != nil {
		s.obs.LogError(ctx, span, "count channel chatters for channel live failed", err,
			zap.Int64("broadcaster_id", info.BroadcasterID))

		return info, nil, nil
	}

	return info, &n, nil
}
