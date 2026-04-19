package handler

import (
	"context"
	"errors"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func (h *Handler) ListChannelChatters(ctx context.Context, req *gen.ListChannelChattersRequest) (gen.ListChannelChattersRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_channel_chatters")
	defer span.End()

	var sessionAt *time.Time

	opt := req.GetSessionStartedAt()
	if opt.IsSet() && !opt.IsNull() {
		t := opt.Value
		sessionAt = &t
	}

	list, err := h.twitch.ListChannelChatters(ctx, req.GetAccountID(), req.GetLogin(), sessionAt)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return &gen.ErrorMessage{Message: "twitch account not found"}, nil
		}

		if errors.Is(err, entity.ErrNoTwitchUserForChannel) {
			return &gen.ErrorMessage{Message: "twitch user not linked for this account"}, nil
		}

		if errors.Is(err, twitchuc.ErrInvalidChannelName) || errors.Is(err, twitchuc.ErrUnknownTwitchChannel) {
			return &gen.ErrorMessage{Message: "unknown channel"}, nil
		}

		h.obs.LogError(ctx, span, "list channel chatters failed", err)

		return nil, err
	}

	out := make(gen.ListChannelChattersOKApplicationJSON, 0, len(list))

	for _, e := range list {
		row := gen.ChannelChatterEntry{
			Login:        e.Login,
			UserTwitchID: e.UserTwitchID,
			PresentSince: e.PresentSince,
		}

		if e.AccountCreatedAt != nil {
			row.AccountCreatedAt = gen.NewOptNilDateTime(*e.AccountCreatedAt)
		} else {
			var z gen.OptNilDateTime
			z.SetToNull()
			row.AccountCreatedAt = z
		}

		if e.MessageCount != nil {
			row.MessageCount = gen.NewOptNilInt64(*e.MessageCount)
		} else {
			var mc gen.OptNilInt64
			mc.SetToNull()
			row.MessageCount = mc
		}

		out = append(out, row)
	}

	return &out, nil
}
