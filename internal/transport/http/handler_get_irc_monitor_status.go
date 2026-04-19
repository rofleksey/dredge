package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetIrcMonitorStatus(ctx context.Context) (*gen.IrcMonitorStatus, error) {
	connected, rows, err := h.twitch.GetIrcMonitorStatus(ctx)
	if err != nil {
		return nil, err
	}

	ch := make([]gen.IrcMonitorStatusChannelsItem, 0, len(rows))

	for _, r := range rows {
		ch = append(ch, gen.IrcMonitorStatusChannelsItem{
			Login: r.Login,
			IrcOk: r.IrcOK,
		})
	}

	return &gen.IrcMonitorStatus{
		Connected: connected,
		Channels:  ch,
	}, nil
}
