package handler

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) CreateRule(ctx context.Context, req *gen.CreateRuleRequest) (*gen.Rule, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.create_rule")
	defer span.End()

	r, err := h.rules.CreateRule(ctx, createRuleReqToEntity(req))
	if err != nil {
		h.obs.LogError(ctx, span, "create rule failed", err, zap.String("event_type", string(req.EventType)))
		return nil, err
	}

	out := ruleEntityToGen(r)

	return &out, nil
}
