package settings

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) PatchTwitchAccount(ctx context.Context, id int64, accountType *string) (entity.TwitchAccount, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.patch_twitch_account")
	defer span.End()

	if accountType != nil {
		t := normalizeTwitchAccountLinkType(*accountType)
		accountType = &t
	}

	out, err := s.repo.PatchTwitchAccount(ctx, id, accountType)
	if err != nil {
		s.obs.LogError(ctx, span, "patch twitch account failed", err, zap.Int64("id", id))
	}

	return out, err
}
