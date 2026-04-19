package settings

import (
	"context"

	"go.uber.org/zap"
)

func (s *Service) DeleteRule(ctx context.Context, id int64) error {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.delete_rule")
	defer span.End()

	err := s.repo.DeleteRule(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "delete rule failed", err, zap.Int64("id", id))
	}

	return err
}
