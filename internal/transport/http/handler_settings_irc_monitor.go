package httptransport

import (
	"context"
	"errors"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetIrcMonitorSettings(ctx context.Context) (*gen.IrcMonitorSettings, error) {
	s, err := h.sett.GetIrcMonitorSettings(ctx)
	if err != nil {
		return nil, err
	}

	return ircMonitorEntityToGen(s), nil
}

func (h *Handler) UpdateIrcMonitorSettings(ctx context.Context, req *gen.IrcMonitorSettings) (*gen.IrcMonitorSettings, error) {
	in := ircMonitorGenToEntity(req)

	out, err := h.sett.UpdateIrcMonitorSettings(ctx, in)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return nil, err
		}
		return nil, err
	}

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		return nil, err
	}

	return ircMonitorEntityToGen(out), nil
}

func ircMonitorEntityToGen(s entity.IrcMonitorSettings) *gen.IrcMonitorSettings {
	g := &gen.IrcMonitorSettings{}
	if s.OauthTwitchAccountID == nil {
		g.OAuthTwitchAccountID.SetToNull()
	} else {
		g.OAuthTwitchAccountID.SetTo(*s.OauthTwitchAccountID)
	}
	g.EnrichmentCooldownHours = int(s.EnrichmentCooldown / time.Hour)
	return g
}

func ircMonitorGenToEntity(req *gen.IrcMonitorSettings) entity.IrcMonitorSettings {
	if req == nil {
		return entity.IrcMonitorSettings{}
	}

	n := req.GetOAuthTwitchAccountID()
	cooldown := time.Duration(req.EnrichmentCooldownHours) * time.Hour
	if cooldown <= 0 {
		cooldown = 24 * time.Hour
	}
	if n.IsNull() {
		return entity.IrcMonitorSettings{
			OauthTwitchAccountID: nil,
			EnrichmentCooldown:   cooldown,
		}
	}

	v, ok := n.Get()
	if !ok {
		return entity.IrcMonitorSettings{
			OauthTwitchAccountID: nil,
			EnrichmentCooldown:   cooldown,
		}
	}

	return entity.IrcMonitorSettings{
		OauthTwitchAccountID: &v,
		EnrichmentCooldown:   cooldown,
	}
}
