package settings

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) GetIrcMonitorSettings(ctx context.Context) (entity.IrcMonitorSettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.get_irc_monitor_settings")
	defer span.End()

	out, err := s.repo.GetIrcMonitorSettings(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "get irc monitor settings failed", err)
	}

	return out, err
}
