package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) ListRules(ctx context.Context) ([]entity.Rule, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.list_rules")
	defer span.End()

	out, err := s.repo.ListRules(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list rules failed", err)
	}

	return out, err
}
