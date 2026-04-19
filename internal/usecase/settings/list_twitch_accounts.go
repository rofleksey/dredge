package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) ListTwitchAccounts(ctx context.Context) ([]entity.TwitchAccount, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.list_twitch_accounts")
	defer span.End()

	out, err := s.repo.ListTwitchAccounts(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list twitch accounts failed", err)
	}

	return out, err
}
