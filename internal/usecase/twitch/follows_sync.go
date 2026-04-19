package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/service/twitch/gql"
	"go.uber.org/zap"
)

// syncUserFollowsFromGQL fetches outgoing follows via Twitch GQL and replaces stored rows. Returns total follow count from Twitch (for suspicion thresholds).
func (s *Usecase) syncUserFollowsFromGQL(ctx context.Context, userID int64) (totalCount int, err error) {
	settings, err := s.repo.GetSuspicionSettings(ctx)
	if err != nil {
		return 0, err
	}

	maxPages := settings.MaxGQLFollowPages
	if maxPages < 1 {
		maxPages = 1
	}

	u, err := s.repo.GetTwitchUserByID(ctx, userID)
	if err != nil {
		return 0, err
	}

	raw, total, err := s.gql.FetchUserFollows(ctx, u.Username, 100, maxPages)
	if err != nil {
		s.obs.Logger.Debug("gql fetch follows failed", zap.Int64("user_id", userID), zap.Error(err))
		return 0, err
	}

	rows := make([]entity.FollowedChannelRow, 0, len(raw))
	for _, r := range raw {
		rows = append(rows, followedChannelFromGQL(r))
	}

	if err := s.repo.ReplaceUserFollowedChannels(ctx, userID, rows); err != nil {
		return 0, err
	}

	if total <= 0 && len(rows) > 0 {
		total = len(rows)
	}

	return total, nil
}

func followedChannelFromGQL(r gql.FollowedChannel) entity.FollowedChannelRow {
	return entity.FollowedChannelRow{
		FollowedChannelID:    r.ChannelID,
		FollowedChannelLogin: r.ChannelLogin,
		FollowedAt:           r.FollowedAt,
	}
}
