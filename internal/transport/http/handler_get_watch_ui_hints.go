package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetWatchUiHints(ctx context.Context) (*gen.WatchUiHints, error) {
	v, c, m := h.twitch.WatchUiHints()

	return &gen.WatchUiHints{
		ViewerPollIntervalSeconds:          int64(v),
		ChannelChattersSyncIntervalSeconds: int64(c),
		MonitoredLivePollIntervalSeconds:   int64(m),
	}, nil
}
