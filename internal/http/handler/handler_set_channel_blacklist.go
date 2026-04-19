package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) SetChannelBlacklist(ctx context.Context, req *gen.ChannelBlacklistChange) (gen.SetChannelBlacklistRes, error) {
	if err := h.sett.SetChannelBlacklist(ctx, req.Login, req.Add); err != nil {
		return &gen.ErrorMessage{Message: err.Error()}, nil
	}

	return &gen.SetChannelBlacklistNoContent{}, nil
}
