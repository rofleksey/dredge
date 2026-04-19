package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) UpdateRule(ctx context.Context, req *gen.UpdateRulePostRequest) (gen.UpdateRuleRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.update_rule")
	defer span.End()

	r, err := h.sett.UpdateRule(ctx, req.GetID(), updateRulePostReqToEntity(req))
	if err != nil {
		if errors.Is(err, entity.ErrRuleNotFound) {
			return &gen.ErrorMessage{Message: "rule not found"}, nil
		}

		h.obs.LogError(ctx, span, "update rule failed", err, zap.Int64("id", req.GetID()))
		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		h.obs.LogError(ctx, span, "restart monitor failed", err, zap.Int64("id", req.GetID()))
		return nil, err
	}

	out := ruleEntityToGen(r)

	return &out, nil
}
