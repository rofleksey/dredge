package twitch

import (
	"sort"
	"strings"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
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

func filterLeaderboardByQuery(rows []entity.StreamLeaderboardRow, q string) []entity.StreamLeaderboardRow {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return rows
	}

	filtered := rows[:0]
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.Login), q) {
			filtered = append(filtered, row)
		}
	}

	return filtered
}
