package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListTwitchDirectoryUsers(ctx context.Context, params gen.ListTwitchDirectoryUsersParams) ([]gen.TwitchUser, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_twitch_directory_users")
	defer span.End()

	limit := 50
	if params.Limit.IsSet() {
		limit = params.Limit.Value
	}

	if limit < 1 {
		limit = 1
	}

	if limit > 200 {
		limit = 200
	}

	f := entity.TwitchUserBrowseFilter{Limit: limit}

	if params.Username.IsSet() {
		f.Username = params.Username.Value
	}

	if params.CursorID.IsSet() {
		v := params.CursorID.Value
		f.CursorID = &v
	}

	if params.MonitoredOnly.IsSet() {
		f.MonitoredOnly = params.MonitoredOnly.Value
	}

	list, err := h.twitch.ListTwitchUsersBrowse(ctx, f)
	if err != nil {
		h.obs.LogError(ctx, span, "list twitch directory users failed", err)
		return nil, err
	}

	out := make([]gen.TwitchUser, 0, len(list))
	for _, ent := range list {
		out = append(out, directoryEntryToGen(ent))
	}

	return out, nil
}
