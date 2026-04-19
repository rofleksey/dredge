package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) CountTwitchDirectoryUsers(ctx context.Context, params gen.CountTwitchDirectoryUsersParams) (*gen.CountResponse, error) {
	f := entity.TwitchUserBrowseFilter{}
	if params.Username.IsSet() {
		f.Username = params.Username.Value
	}

	if params.MonitoredOnly.IsSet() {
		f.MonitoredOnly = params.MonitoredOnly.Value
	}

	n, err := h.twitch.CountTwitchUsersBrowse(ctx, f)
	if err != nil {
		return nil, err
	}

	return &gen.CountResponse{Total: n}, nil
}
