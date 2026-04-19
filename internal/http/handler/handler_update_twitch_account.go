package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) UpdateTwitchAccount(ctx context.Context, req *gen.UpdateTwitchAccountPostRequest) (gen.UpdateTwitchAccountRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.update_twitch_account")
	defer span.End()

	at := string(req.GetAccountType())

	a, err := h.sett.PatchTwitchAccount(ctx, req.GetID(), &at)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return &gen.ErrorMessage{Message: "twitch account not found"}, nil
		}

		h.obs.LogError(ctx, span, "update twitch account failed", err, zap.Int64("id", req.GetID()))
		return nil, err
	}

	out := twitchAccountToAPI(a)

	return &out, nil
}
