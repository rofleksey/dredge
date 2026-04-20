package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) ListRuleTriggers(ctx context.Context, params gen.ListRuleTriggersParams) ([]gen.RuleTrigger, error) {
	f := entity.RuleTriggerListFilter{}

	if v, ok := params.Limit.Get(); ok {
		f.Limit = v
	}

	if v, ok := params.CursorCreatedAt.Get(); ok {
		f.CursorCreatedAt = &v
	}

	if v, ok := params.CursorID.Get(); ok {
		f.CursorID = &v
	}

	list, err := h.rules.ListRuleTriggers(ctx, f)
	if err != nil {
		return nil, err
	}

	out := make([]gen.RuleTrigger, 0, len(list))

	for _, e := range list {
		out = append(out, ruleTriggerEntityToGen(e))
	}

	return out, nil
}
