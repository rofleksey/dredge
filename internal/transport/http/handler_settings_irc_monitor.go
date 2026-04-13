package httptransport

import (
	"context"
	"errors"

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
	return g
}

func ircMonitorGenToEntity(req *gen.IrcMonitorSettings) entity.IrcMonitorSettings {
	if req == nil {
		return entity.IrcMonitorSettings{}
	}

	n := req.GetOAuthTwitchAccountID()
	if n.IsNull() {
		return entity.IrcMonitorSettings{OauthTwitchAccountID: nil}
	}

	v, ok := n.Get()
	if !ok {
		return entity.IrcMonitorSettings{OauthTwitchAccountID: nil}
	}

	return entity.IrcMonitorSettings{OauthTwitchAccountID: &v}
}
