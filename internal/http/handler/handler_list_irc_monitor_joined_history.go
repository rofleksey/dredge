package handler

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) ListIrcMonitorJoinedHistory(ctx context.Context, params gen.ListIrcMonitorJoinedHistoryParams) ([]gen.IrcJoinedSample, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_irc_monitor_joined_history")
	defer span.End()

	days := params.Days.Or(7)

	rows, err := h.twitch.ListIrcJoinedSamplesLastDays(ctx, days)
	if err != nil {
		h.obs.LogError(ctx, span, "list irc monitor joined history failed", err, zap.Int("days", days))

		return nil, err
	}

	out := make([]gen.IrcJoinedSample, 0, len(rows))

	for _, r := range rows {
		out = append(out, gen.IrcJoinedSample{
			CapturedAt:  r.CapturedAt,
			JoinedCount: r.JoinedCount,
		})
	}

	return out, nil
}
