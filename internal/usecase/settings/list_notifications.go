package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) ListNotifications(ctx context.Context) ([]entity.NotificationEntry, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.list_notifications")
	defer span.End()

	return s.repo.ListNotificationEntries(ctx)
}
