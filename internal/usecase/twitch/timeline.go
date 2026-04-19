package twitch

import (
	"sort"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

// BuildActivityTimelineSegments merges chat_online/offline activity into intervals for charts.
func BuildActivityTimelineSegments(events []entity.UserActivityEvent, windowEnd time.Time) []entity.ActivityTimelineSegment {
	if len(events) == 0 {
		return nil
	}

	sorted := append([]entity.UserActivityEvent(nil), events...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CreatedAt.Equal(sorted[j].CreatedAt) {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	ircSegs := buildIRCPresenceSegments(sorted, windowEnd)

	return mergeActivityTimelineSegments(ircSegs)
}

func buildIRCPresenceSegments(sorted []entity.UserActivityEvent, windowEnd time.Time) []entity.ActivityTimelineSegment {
	type openSeg struct {
		start time.Time
		login string
	}

	open := make(map[int64]openSeg)

	var out []entity.ActivityTimelineSegment

	for _, e := range sorted {
		var chID int64
		if e.ChannelTwitchUserID != nil {
			chID = *e.ChannelTwitchUserID
		}

		if chID == 0 {
			continue
		}

		login := e.ChannelLogin

		switch e.EventType {
		case entity.UserActivityChatOnline:
			open[chID] = openSeg{start: e.CreatedAt, login: login}

		case entity.UserActivityChatOffline:
			if st, ok := open[chID]; ok {
				out = append(out, entity.ActivityTimelineSegment{
					ChannelTwitchUserID: chID,
					ChannelLogin:        st.login,
					Start:               st.start,
					End:                 e.CreatedAt,
				})
				delete(open, chID)
			}
		}
	}

	for chID, st := range open {
		out = append(out, entity.ActivityTimelineSegment{
			ChannelTwitchUserID: chID,
			ChannelLogin:        st.login,
			Start:               st.start,
			End:                 windowEnd,
		})
		_ = chID
	}

	return out
}

func mergeActivityTimelineSegments(segs []entity.ActivityTimelineSegment) []entity.ActivityTimelineSegment {
	if len(segs) == 0 {
		return nil
	}

	sort.Slice(segs, func(i, j int) bool {
		if segs[i].ChannelTwitchUserID != segs[j].ChannelTwitchUserID {
			return segs[i].ChannelTwitchUserID < segs[j].ChannelTwitchUserID
		}

		return segs[i].Start.Before(segs[j].Start)
	})

	out := make([]entity.ActivityTimelineSegment, 0, len(segs))
	cur := segs[0]

	for i := 1; i < len(segs); i++ {
		next := segs[i]
		if next.ChannelTwitchUserID != cur.ChannelTwitchUserID {
			out = append(out, cur)
			cur = next
			continue
		}

		if next.Start.After(cur.End) {
			out = append(out, cur)
			cur = next
			continue
		}

		if next.End.After(cur.End) {
			cur.End = next.End
		}
	}

	out = append(out, cur)

	return out
}
