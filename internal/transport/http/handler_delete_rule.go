package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) DeleteRule(ctx context.Context, req *gen.DeleteByIDRequest) (gen.DeleteRuleRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.delete_rule")
	defer span.End()

	if err := h.sett.DeleteRule(ctx, req.ID); err != nil {
		if errors.Is(err, entity.ErrRuleNotFound) {
			return &gen.ErrorMessage{Message: "rule not found"}, nil
		}

		h.obs.LogError(ctx, span, "delete rule failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		h.obs.LogError(ctx, span, "restart monitor failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	return &gen.DeleteRuleNoContent{}, nil
}
