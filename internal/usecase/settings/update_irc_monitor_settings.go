package settings

import (
	"context"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Service) UpdateIrcMonitorSettings(ctx context.Context, in entity.IrcMonitorSettings) (entity.IrcMonitorSettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.update_irc_monitor_settings")
	defer span.End()

	if in.EnrichmentCooldown <= 0 {
		in.EnrichmentCooldown = 24 * time.Hour
	}

	if in.OauthTwitchAccountID != nil {
		if _, err := s.repo.GetTwitchAccountByID(ctx, *in.OauthTwitchAccountID); err != nil {
			s.obs.LogError(ctx, span, "irc monitor oauth account validation failed", err)
			return entity.IrcMonitorSettings{}, err
		}
	}

	if err := s.repo.UpdateIrcMonitorSettings(ctx, in); err != nil {
		s.obs.LogError(ctx, span, "update irc monitor settings failed", err)
		return entity.IrcMonitorSettings{}, err
	}

	return s.repo.GetIrcMonitorSettings(ctx)
}
