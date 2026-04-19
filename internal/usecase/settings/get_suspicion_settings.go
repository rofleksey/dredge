package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) GetSuspicionSettings(ctx context.Context) (entity.SuspicionSettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.get_suspicion_settings")
	defer span.End()

	out, err := s.repo.GetSuspicionSettings(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "get suspicion settings failed", err)
	}

	return out, err
}
