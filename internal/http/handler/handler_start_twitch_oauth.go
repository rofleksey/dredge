package handler

import (
	"context"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) StartTwitchOAuth(ctx context.Context, req gen.OptStartTwitchOAuthRequest) (*gen.StartTwitchOAuthResponse, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.start_twitch_oauth")
	defer span.End()

	spaReturn := ""

	if req.IsSet() {
		body, _ := req.Get()
		if v, ok := body.GetReturnURL().Get(); ok {
			spaReturn = v
		}
	}

	state, err := h.twitchOAuth.NewState(spaReturn)
	if err != nil {
		h.obs.LogError(ctx, span, "twitch oauth state failed", err)
		return nil, err
	}

	return &gen.StartTwitchOAuthResponse{AuthorizeURL: h.twitchOAuth.AuthorizeURL(state)}, nil
}
