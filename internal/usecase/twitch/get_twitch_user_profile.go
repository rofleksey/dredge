package twitch

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

// GetTwitchUserProfile returns profile fields, message count, IRC presence seconds this week (UTC Mon..now), helix created_at, profile image, monitored follows, GQL full follows list, and global blacklist for UI.
func (s *Service) GetTwitchUserProfile(ctx context.Context, id int64) (
	u entity.TwitchUser,
	messageCount int64,
	presenceSec int64,
	accountCreated *time.Time,
	profileImageURL *string,
	monitoredFollows []entity.FollowedMonitoredChannel,
	gqlFollows []entity.FollowedChannelRow,
	channelBlacklist []string,
	err error,
) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.get_twitch_user_profile")
	defer span.End()

	u, err = s.repo.GetTwitchUserByID(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "get twitch user failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	messageCount, err = s.repo.CountChatMessagesByChatter(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "count chatter messages failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	presenceSec, err = s.presenceSecondsThisWeek(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "presence this week failed", err, zap.Int64("id", id))

		presenceSec = 0
	}

	accountCreated, _, profileImageURL, err = s.repo.GetHelixMeta(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "get helix meta failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	monitoredFollows, err = s.repo.ListFollowedMonitoredChannels(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "list follows failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	gqlFollows, err = s.repo.ListUserFollowedChannels(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "list gql follows failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	channelBlacklist, err = s.repo.ListChannelBlacklist(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list blacklist failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, 0, 0, nil, nil, nil, nil, nil, err
	}

	return u, messageCount, presenceSec, accountCreated, profileImageURL, monitoredFollows, gqlFollows, channelBlacklist, nil
}
