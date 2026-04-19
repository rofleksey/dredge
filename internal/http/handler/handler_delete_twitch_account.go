package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) DeleteTwitchAccount(ctx context.Context, req *gen.DeleteByIDRequest) (gen.DeleteTwitchAccountRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.delete_twitch_account")
	defer span.End()

	if err := h.sett.DeleteTwitchAccount(ctx, req.ID); err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return &gen.ErrorMessage{Message: "twitch account not found"}, nil
		}

		h.obs.LogError(ctx, span, "delete twitch account failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		h.obs.LogError(ctx, span, "restart monitor after twitch account delete failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	return &gen.DeleteTwitchAccountNoContent{}, nil
}
