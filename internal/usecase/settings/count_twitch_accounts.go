package settings

import (
	"context"
)

func (s *Service) CountTwitchAccounts(ctx context.Context) (int64, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.count_twitch_accounts")
	defer span.End()

	return s.repo.CountTwitchAccounts(ctx)
}
