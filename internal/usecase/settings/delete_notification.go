package settings

import (
	"context"
)

func (s *Usecase) DeleteNotification(ctx context.Context, id int64) error {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.delete_notification")
	defer span.End()

	return s.repo.DeleteNotificationEntry(ctx, id)
}
