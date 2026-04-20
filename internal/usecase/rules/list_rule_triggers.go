package rules

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) ListRuleTriggers(ctx context.Context, f entity.RuleTriggerListFilter) ([]entity.RuleTriggerEvent, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.rules.list_rule_triggers")
	defer span.End()

	list, err := s.repo.ListRuleTriggerEvents(ctx, f)
	if err != nil {
		s.obs.LogError(ctx, span, "list rule trigger events failed", err)

		return nil, err
	}

	return list, nil
}
