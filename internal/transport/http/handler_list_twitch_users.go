package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListTwitchUsers(ctx context.Context, params gen.ListTwitchUsersParams) ([]gen.TwitchUser, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_twitch_users")
	defer span.End()

	monitoredOnly := params.MonitoredOnly.Value
	list, err := h.sett.ListTwitchUsers(ctx, monitoredOnly)
	if err != nil {
		h.obs.LogError(ctx, span, "list twitch users failed", err)
		return nil, err
	}

	out := make([]gen.TwitchUser, 0, len(list))

	for _, u := range list {
		out = append(out, entityTwitchUserToGen(u))
	}

	return out, nil
}
