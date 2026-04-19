package httptransport

import (
	"context"
	"errors"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetTwitchUserActivityTimeline(ctx context.Context, req *gen.GetTwitchUserActivityTimelineRequest) (gen.GetTwitchUserActivityTimelineRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_twitch_user_activity_timeline")
	defer span.End()

	if _, err := h.twitch.GetTwitchUser(ctx, req.GetID()); err != nil {
		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return &gen.ErrorMessage{Message: "twitch user not found"}, nil
		}

		return nil, err
	}

	to := time.Now().UTC()
	if req.To.IsSet() {
		to = req.To.Value
	}

	from := to.Add(-7 * 24 * time.Hour)
	if req.From.IsSet() {
		from = req.From.Value
	}

	segs, err := h.twitch.GetUserActivityTimeline(ctx, req.GetID(), from, to)
	if err != nil {
		h.obs.LogError(ctx, span, "timeline failed", err)
		return nil, err
	}

	tl := make(gen.GetTwitchUserActivityTimelineOKApplicationJSON, 0, len(segs))

	for _, seg := range segs {
		tl = append(tl, gen.ActivityTimelineSegment{
			ChannelID:    seg.ChannelTwitchUserID,
			ChannelLogin: seg.ChannelLogin,
			Start:        seg.Start,
			End:          seg.End,
		})
	}

	return &tl, nil
}
