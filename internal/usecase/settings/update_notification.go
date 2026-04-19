package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) UpdateNotification(ctx context.Context, id int64, provider *string, settings map[string]any, enabled *bool) (entity.NotificationEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.update_notification")
	defer span.End()

	return s.repo.UpdateNotificationEntry(ctx, id, provider, settings, enabled)
}
