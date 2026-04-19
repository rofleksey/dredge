package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) CreateNotification(ctx context.Context, provider string, settings map[string]any, enabled bool) (entity.NotificationEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.create_notification")
	defer span.End()

	return s.repo.CreateNotificationEntry(ctx, provider, settings, enabled)
}
