package handler

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/http/gen"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func (h *Handler) GetChannelLive(ctx context.Context, req *gen.GetChannelLiveRequest) (gen.GetChannelLiveRes, error) {
	info, chatterCount, err := h.twitch.GetChannelLive(ctx, req.Login)
	if err != nil {
		if errors.Is(err, twitchuc.ErrInvalidChannelName) || errors.Is(err, twitchuc.ErrUnknownTwitchChannel) {
			return &gen.ErrorMessage{Message: "unknown channel"}, nil
		}

		return nil, err
	}

	cl := gen.ChannelLive{
		BroadcasterID:    info.BroadcasterID,
		BroadcasterLogin: info.BroadcasterLogin,
		DisplayName:      info.DisplayName,
		ProfileImageURL:  info.ProfileImageURL,
		IsLive:           info.IsLive,
	}

	if info.Title != "" {
		cl.SetTitle(gen.NewOptNilString(info.Title))
	} else {
		var t gen.OptNilString
		t.SetToNull()
		cl.SetTitle(t)
	}

	if info.GameName != "" {
		cl.SetGameName(gen.NewOptNilString(info.GameName))
	} else {
		var g gen.OptNilString
		g.SetToNull()
		cl.SetGameName(g)
	}

	if info.IsLive && info.ViewerCount >= 0 {
		cl.SetViewerCount(gen.NewOptNilInt64(info.ViewerCount))
	} else {
		var v gen.OptNilInt64
		v.SetToNull()
		cl.SetViewerCount(v)
	}

	if chatterCount != nil {
		cl.SetChannelChatterCount(gen.NewOptNilInt64(*chatterCount))
	} else {
		var cc gen.OptNilInt64
		cc.SetToNull()
		cl.SetChannelChatterCount(cc)
	}

	if info.StreamStartedAt != nil {
		cl.SetStartedAt(gen.NewOptNilDateTime(*info.StreamStartedAt))
	} else {
		var s gen.OptNilDateTime
		s.SetToNull()
		cl.SetStartedAt(s)
	}

	return &cl, nil
}
