package settings

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) CreateTwitchAccount(ctx context.Context, id int64, username, refreshToken, accountType string) (entity.TwitchAccount, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.create_twitch_account")
	defer span.End()

	accountType = normalizeTwitchAccountLinkType(accountType)

	s.obs.Logger.Debug("create twitch account", zap.Int64("id", id), zap.String("username", username), zap.String("account_type", accountType))

	out, err := s.repo.CreateTwitchAccount(ctx, id, username, refreshToken, accountType)
	if err != nil {
		s.obs.LogError(ctx, span, "create twitch account failed", err, zap.String("username", username))
	}

	return out, err
}
