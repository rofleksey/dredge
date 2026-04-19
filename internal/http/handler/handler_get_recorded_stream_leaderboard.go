package handler

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) GetRecordedStreamLeaderboard(ctx context.Context, params gen.GetRecordedStreamLeaderboardParams) (gen.GetRecordedStreamLeaderboardRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_recorded_stream_leaderboard")
	defer span.End()

	st, err := h.twitch.GetMonitoredStream(ctx, params.StreamId)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return &gen.ErrorMessage{Message: "stream not found"}, nil
		}

		h.obs.LogError(ctx, span, "get stream for leaderboard failed", err)
		return nil, err
	}

	sort := entity.StreamLeaderboardSortPresenceDesc
	if v, ok := params.Sort.Get(); ok {
		sort = entity.StreamLeaderboardSort(v)
	}

	q := ""
	if v, ok := params.Q.Get(); ok {
		q = v
	}

	rows, err := h.twitch.StreamLeaderboard(ctx, st, sort, q)
	if err != nil {
		h.obs.LogError(ctx, span, "stream leaderboard failed", err)
		return nil, err
	}

	out := make(gen.GetRecordedStreamLeaderboardOKApplicationJSON, 0, len(rows))

	for _, r := range rows {
		row := gen.StreamLeaderboardEntry{
			Login:           r.Login,
			UserTwitchID:    r.UserTwitchID,
			PresenceSeconds: r.PresenceSeconds,
			MessageCount:    r.MessageCount,
		}

		if r.AccountCreatedAt != nil {
			row.AccountCreatedAt = gen.NewOptNilDateTime(*r.AccountCreatedAt)
		} else {
			var z gen.OptNilDateTime
			z.SetToNull()
			row.AccountCreatedAt = z
		}

		out = append(out, row)
	}

	return &out, nil
}
