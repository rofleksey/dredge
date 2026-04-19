package settings

import (
	"context"
)

func (s *Service) CountRules(ctx context.Context) (int64, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.count_rules")
	defer span.End()

	return s.repo.CountRules(ctx)
}
