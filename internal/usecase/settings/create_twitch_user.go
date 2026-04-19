package settings

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) CreateTwitchUser(ctx context.Context, id int64, username string) (entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.create_twitch_user")
	defer span.End()

	s.obs.Logger.Debug("create twitch user", zap.Int64("id", id), zap.String("username", username))

	out, err := s.repo.CreateTwitchUser(ctx, id, username)
	if err != nil {
		s.obs.LogError(ctx, span, "create twitch user failed", err,
			zap.Int64("id", id), zap.String("username", username))
	}

	return out, err
}
