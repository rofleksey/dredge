package settings

import (
	"context"

	"go.uber.org/zap"
)

func (s *Usecase) DeleteTwitchAccount(ctx context.Context, id int64) error {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.delete_twitch_account")
	defer span.End()

	err := s.repo.DeleteTwitchAccount(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "delete twitch account failed", err, zap.Int64("id", id))
	}

	return err
}
