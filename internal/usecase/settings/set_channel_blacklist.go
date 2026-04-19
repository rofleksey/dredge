package settings

import (
	"context"

	"go.uber.org/zap"
)

func (s *Usecase) SetChannelBlacklist(ctx context.Context, login string, add bool) error {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.set_channel_blacklist")
	defer span.End()

	var err error

	if add {
		err = s.repo.AddChannelBlacklist(ctx, login)
	} else {
		err = s.repo.RemoveChannelBlacklist(ctx, login)
	}

	if err != nil {
		s.obs.LogError(ctx, span, "set channel blacklist failed", err, zap.String("login", login), zap.Bool("add", add))
	}

	return err
}
