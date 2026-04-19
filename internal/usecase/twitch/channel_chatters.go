package twitch

import (
	"context"
	"sort"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

const enrichMissingAccountBatch = 64

// userIDsToEnrichMissingAccounts picks chatters without Helix account_created_at, newest presence first, capped.
func userIDsToEnrichMissingAccounts(list []entity.ChannelChatterEntry, limit int) []int64 {
	var missing []entity.ChannelChatterEntry

	for _, e := range list {
		if e.AccountCreatedAt == nil {
			missing = append(missing, e)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	sort.Slice(missing, func(i, j int) bool {
		return missing[i].PresentSince.After(missing[j].PresentSince)
	})

	if len(missing) > limit {
		missing = missing[:limit]
	}

	out := make([]int64, 0, len(missing))

	for _, e := range missing {
		out = append(out, e.UserTwitchID)
	}

	return out
}

// ListChannelChatters returns chatters from the IRC-maintained snapshot with presence and Helix account dates when known.
// If sessionStartedAt is set, MessageCount is filled from persisted chat for this channel since that instant.
func (s *Service) ListChannelChatters(ctx context.Context, accountID int64, channelLogin string, sessionStartedAt *time.Time) ([]entity.ChannelChatterEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_channel_chatters")
	defer span.End()

	if _, err := s.repo.GetTwitchAccountByID(ctx, accountID); err != nil {
		return nil, err
	}

	ch, err := s.ResolveChannel(ctx, channelLogin)
	if err != nil {
		return nil, err
	}

	list, err := s.repo.ListChannelChatterEntries(ctx, ch.ID)
	if err != nil {
		return nil, err
	}

	var counts map[int64]int64

	if sessionStartedAt != nil {
		counts, err = s.repo.CountChatMessagesPerChatterForChannelSince(ctx, ch.ID, sessionStartedAt.UTC())
		if err != nil {
			s.obs.LogError(ctx, span, "count chatter messages since session failed", err)

			counts = nil
		}
	}

	for i := range list {
		if counts != nil {
			if n, ok := counts[list[i].UserTwitchID]; ok {
				v := n
				list[i].MessageCount = &v
			}
		}
	}

	for _, id := range userIDsToEnrichMissingAccounts(list, enrichMissingAccountBatch) {
		s.EnqueueUserEnrichment(id)
	}

	return list, nil
}
