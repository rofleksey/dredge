package twitch

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

const (
	ircJoinedHistoryDefaultDays = 7
	ircJoinedHistoryMaxDays     = 90
)

// CountIrcJoinedChannels returns how many monitored channels are currently joined on IRC (reconciler state).
func (s *Usecase) CountIrcJoinedChannels(ctx context.Context) (int, error) {
	_, rows, err := s.live.GetIrcMonitorStatus(ctx)
	if err != nil {
		return 0, err
	}

	n := 0

	for _, r := range rows {
		if r.IrcOK {
			n++
		}
	}

	return n, nil
}

// RecordIrcJoinedSnapshot persists the current joined count (0 when IRC not connected / not joined).
func (s *Usecase) RecordIrcJoinedSnapshot(ctx context.Context) error {
	n, err := s.CountIrcJoinedChannels(ctx)
	if err != nil {
		return err
	}

	return s.repo.InsertIrcJoinedSample(ctx, n)
}

// ListIrcJoinedSamplesLastDays returns samples from now back `days` calendar days (UTC window).
func (s *Usecase) ListIrcJoinedSamplesLastDays(ctx context.Context, days int) ([]entity.IrcJoinedSample, error) {
	if days <= 0 {
		days = ircJoinedHistoryDefaultDays
	}

	if days > ircJoinedHistoryMaxDays {
		days = ircJoinedHistoryMaxDays
	}

	to := time.Now().UTC()
	from := to.Add(-time.Duration(days) * 24 * time.Hour)

	return s.repo.ListIrcJoinedSamples(ctx, from, to)
}
