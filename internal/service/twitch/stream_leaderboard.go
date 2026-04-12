package twitch

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

func presenceSecondsClipped(segs []entity.ActivityTimelineSegment, winStart, winEnd time.Time) int64 {
	var sum int64

	for _, seg := range segs {
		s := seg.Start
		e := seg.End

		if e.Before(winStart) || s.After(winEnd) {
			continue
		}

		if s.Before(winStart) {
			s = winStart
		}

		if e.After(winEnd) {
			e = winEnd
		}

		if e.After(s) {
			sum += int64(e.Sub(s).Seconds())
		}
	}

	return sum
}

// StreamLeaderboard builds merged stats for a recorded stream.
func (s *Service) StreamLeaderboard(ctx context.Context, stream entity.Stream, sort entity.StreamLeaderboardSort, q string) ([]entity.StreamLeaderboardRow, error) {
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

		acct, _, err := s.repo.GetHelixMeta(ctx, id)
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

	q = strings.TrimSpace(strings.ToLower(q))
	if q != "" {
		filtered := out[:0]
		for _, row := range out {
			if strings.Contains(strings.ToLower(row.Login), q) {
				filtered = append(filtered, row)
			}
		}

		out = filtered
	}

	sortStreamLeaderboard(out, sort)

	return out, nil
}

func sortStreamLeaderboard(rows []entity.StreamLeaderboardRow, s entity.StreamLeaderboardSort) {
	if s == "" {
		s = entity.StreamLeaderboardSortPresenceDesc
	}

	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]

		switch s {
		case entity.StreamLeaderboardSortPresenceAsc:
			if a.PresenceSeconds != b.PresenceSeconds {
				return a.PresenceSeconds < b.PresenceSeconds
			}
		case entity.StreamLeaderboardSortPresenceDesc:
			if a.PresenceSeconds != b.PresenceSeconds {
				return a.PresenceSeconds > b.PresenceSeconds
			}
		case entity.StreamLeaderboardSortMessagesAsc:
			if a.MessageCount != b.MessageCount {
				return a.MessageCount < b.MessageCount
			}
		case entity.StreamLeaderboardSortMessagesDesc:
			if a.MessageCount != b.MessageCount {
				return a.MessageCount > b.MessageCount
			}
		case entity.StreamLeaderboardSortLoginAZ:
			if a.Login != b.Login {
				return a.Login < b.Login
			}
		case entity.StreamLeaderboardSortLoginZA:
			if a.Login != b.Login {
				return a.Login > b.Login
			}
		case entity.StreamLeaderboardSortAccountNew:
			ha, hb := a.AccountCreatedAt != nil, b.AccountCreatedAt != nil
			if ha && hb && !a.AccountCreatedAt.Equal(*b.AccountCreatedAt) {
				return a.AccountCreatedAt.After(*b.AccountCreatedAt)
			}

			if ha != hb {
				return ha
			}

			if a.PresenceSeconds != b.PresenceSeconds {
				return a.PresenceSeconds > b.PresenceSeconds
			}
		case entity.StreamLeaderboardSortAccountOld:
			ha, hb := a.AccountCreatedAt != nil, b.AccountCreatedAt != nil
			if ha && hb && !a.AccountCreatedAt.Equal(*b.AccountCreatedAt) {
				return a.AccountCreatedAt.Before(*b.AccountCreatedAt)
			}

			if ha != hb {
				return ha
			}

			if a.PresenceSeconds != b.PresenceSeconds {
				return a.PresenceSeconds > b.PresenceSeconds
			}
		default:
			if a.PresenceSeconds != b.PresenceSeconds {
				return a.PresenceSeconds > b.PresenceSeconds
			}
		}

		if a.MessageCount != b.MessageCount {
			return a.MessageCount > b.MessageCount
		}

		return a.Login < b.Login
	})
}
