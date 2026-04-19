package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) GetIrcMonitorSettings(ctx context.Context) (*gen.IrcMonitorSettings, error) {
	s, err := h.sett.GetIrcMonitorSettings(ctx)
	if err != nil {
		return nil, err
	}

	return ircMonitorEntityToGen(s), nil
}
