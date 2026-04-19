package twitch

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// StreamLeaderboard builds merged stats for a recorded stream.
func (s *Usecase) StreamLeaderboard(ctx context.Context, stream entity.Stream, sort entity.StreamLeaderboardSort, q string) ([]entity.StreamLeaderboardRow, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.stream_leaderboard")
	defer span.End()

	streamEnd := time.Now().UTC()
	if stream.EndedAt != nil {
		streamEnd = stream.EndedAt.UTC()
	}

	evs, err := s.repo.ListUserActivityEventsForChannelPresence(ctx, stream.ChannelTwitchUserID, stream.StartedAt.UTC(), streamEnd)
	if err != nil {
		s.obs.LogError(ctx, span, "channel presence events for leaderboard failed", err)
		return nil, err
	}

	byChatter := make(map[int64][]entity.UserActivityEvent)

	for _, e := range evs {
		byChatter[e.ChatterTwitchUserID] = append(byChatter[e.ChatterTwitchUserID], e)
	}

	presence := make(map[int64]int64, len(byChatter))

	for chatterID, list := range byChatter {
		segs := BuildActivityTimelineSegments(list, streamEnd)
		presence[chatterID] = presenceSecondsClipped(segs, stream.StartedAt.UTC(), streamEnd)
	}

	msgCounts, err := s.repo.CountChatMessagesPerChatterForStream(ctx, stream.ID)
	if err != nil {
		s.obs.LogError(ctx, span, "message counts for stream leaderboard failed", err)
		return nil, err
	}

	ids := make(map[int64]struct{})
	for id := range presence {
		ids[id] = struct{}{}
	}

	for id := range msgCounts {
		ids[id] = struct{}{}
	}

	out := make([]entity.StreamLeaderboardRow, 0, len(ids))

	for id := range ids {
		u, err := s.repo.GetTwitchUserByID(ctx, id)
		if err != nil {
			s.obs.LogError(ctx, span, "get twitch user for leaderboard row failed", err, zap.Int64("id", id))
			continue
		}

		mc := msgCounts[id]
		ps := presence[id]

		acct, _, _, err := s.repo.GetHelixMeta(ctx, id)
		if err != nil {
			s.obs.LogError(ctx, span, "helix meta for leaderboard failed", err, zap.Int64("id", id))
		}

		out = append(out, entity.StreamLeaderboardRow{
			Login:            u.Username,
			UserTwitchID:     id,
			PresenceSeconds:  ps,
			MessageCount:     mc,
			AccountCreatedAt: acct,
		})
	}

	out = filterLeaderboardByQuery(out, q)

	sortStreamLeaderboard(out, sort)

	return out, nil
}
