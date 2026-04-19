package handler

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) UpdateIrcMonitorSettings(ctx context.Context, req *gen.IrcMonitorSettings) (*gen.IrcMonitorSettings, error) {
	in := ircMonitorGenToEntity(req)

	out, err := h.sett.UpdateIrcMonitorSettings(ctx, in)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return nil, err
		}

		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		return nil, err
	}

	return ircMonitorEntityToGen(out), nil
}
