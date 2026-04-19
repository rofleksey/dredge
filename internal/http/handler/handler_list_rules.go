package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) ListRules(ctx context.Context) ([]gen.Rule, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_rules")
	defer span.End()

	list, err := h.rules.ListRules(ctx)
	if err != nil {
		h.obs.LogError(ctx, span, "list rules failed", err)
		return nil, err
	}

	out := make([]gen.Rule, 0, len(list))

	for _, r := range list {
		out = append(out, ruleEntityToGen(r))
	}

	return out, nil
}
