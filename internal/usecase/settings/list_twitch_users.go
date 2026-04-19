package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) ListTwitchUsers(ctx context.Context, monitoredOnly bool) ([]entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.list_twitch_users")
	defer span.End()

	var (
		out []entity.TwitchUser
		err error
	)

	if monitoredOnly {
		out, err = s.repo.ListMonitoredTwitchUsers(ctx)
	} else {
		out, err = s.repo.ListTwitchUsers(ctx)
	}

	if err != nil {
		s.obs.LogError(ctx, span, "list twitch users failed", err)
	}

	return out, err
}
