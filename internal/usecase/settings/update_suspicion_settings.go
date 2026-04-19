package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) UpdateSuspicionSettings(ctx context.Context, in entity.SuspicionSettings) (entity.SuspicionSettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.update_suspicion_settings")
	defer span.End()

	if err := s.repo.UpdateSuspicionSettings(ctx, in); err != nil {
		s.obs.LogError(ctx, span, "update suspicion settings failed", err)
		return entity.SuspicionSettings{}, err
	}

	return s.repo.GetSuspicionSettings(ctx)
}
