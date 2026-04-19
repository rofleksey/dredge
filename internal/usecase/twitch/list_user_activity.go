package twitch

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

// ListUserActivity returns paginated activity events for a chatter.
func (s *Usecase) ListUserActivity(ctx context.Context, f entity.UserActivityListFilter) ([]entity.UserActivityEvent, error) {
	ctx, span := s.obs.StartSpan(ctx, "service.twitch.list_user_activity")
	defer span.End()

	return s.repo.ListUserActivityEvents(ctx, f)
}
