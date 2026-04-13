package httptransport

import (
	"context"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) CountTwitchMessages(ctx context.Context, params gen.CountTwitchMessagesParams) (*gen.CountResponse, error) {
	f := entity.ChatMessageListFilter{}
	if params.Username.IsSet() {
		f.Username = params.Username.Value
	}

	if params.Text.IsSet() {
		f.Text = params.Text.Value
	}

	if params.Channel.IsSet() {
		f.Channel = params.Channel.Value
	}

	if t, ok := params.CreatedFrom.Get(); ok {
		f.CreatedFrom = &t
	}

	if t, ok := params.CreatedTo.Get(); ok {
		f.CreatedTo = &t
	}

	if params.ChatterUserID.IsSet() {
		v := params.ChatterUserID.Value
		f.ChatterUserID = &v
	}

	n, err := h.twitch.CountChatMessages(ctx, f)
	if err != nil {
		return nil, err
	}

	return &gen.CountResponse{Total: n}, nil
}

func (h *Handler) CountTwitchDirectoryUsers(ctx context.Context, params gen.CountTwitchDirectoryUsersParams) (*gen.CountResponse, error) {
	f := entity.TwitchUserBrowseFilter{}
	if params.Username.IsSet() {
		f.Username = params.Username.Value
	}

	n, err := h.twitch.CountTwitchUsersBrowse(ctx, f)
	if err != nil {
		return nil, err
	}
	return &gen.CountResponse{Total: n}, nil
}

func (h *Handler) CountTwitchAccounts(ctx context.Context) (*gen.CountResponse, error) {
	n, err := h.sett.CountTwitchAccounts(ctx)
	if err != nil {
		return nil, err
	}
	return &gen.CountResponse{Total: n}, nil
}
