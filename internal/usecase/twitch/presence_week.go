package twitch

import (
	"context"
	"time"
)

func startOfWeekMondayUTC(t time.Time) time.Time {
	t = t.UTC()
	wd := int(t.Weekday())
	daysSinceMon := (wd + 6) % 7
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return start.AddDate(0, 0, -daysSinceMon)
}

func (s *Usecase) presenceSecondsThisWeek(ctx context.Context, chatterID int64) (int64, error) {
	now := time.Now().UTC()
	from := startOfWeekMondayUTC(now)

	evs, err := s.repo.ListUserActivityEventsForTimeline(ctx, chatterID, from, now)
	if err != nil {
		return 0, err
	}

	segs := BuildActivityTimelineSegments(evs, now)

	var sum int64

	for _, seg := range segs {
		if seg.End.After(seg.Start) {
			sum += int64(seg.End.Sub(seg.Start).Seconds())
		}
	}

	return sum, nil
}
