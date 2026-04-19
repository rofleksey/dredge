package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/usecase/rules"
)

func (h *Handler) ListRuleTemplateVariables(ctx context.Context) (*gen.RuleTemplateVariablesResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_rule_template_variables")
	defer span.End()

	list := rules.RuleTemplateVariables()
	out := make([]gen.RuleTemplateVariable, len(list))

	for i := range list {
		out[i] = gen.RuleTemplateVariable{
			Name:        list[i].Name,
			Description: list[i].Description,
		}
	}

	return &gen.RuleTemplateVariablesResponse{Variables: out}, nil
}
