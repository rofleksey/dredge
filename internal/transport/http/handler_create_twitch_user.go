package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/transport/http/gen"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func (h *Handler) CreateTwitchUser(ctx context.Context, req *gen.CreateTwitchUserRequest) (gen.CreateTwitchUserRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.create_twitch_user")
	defer span.End()

	resolved, err := h.twitch.ResolveChannel(ctx, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, twitchuc.ErrUnknownTwitchChannel):
			return &gen.ErrorMessage{Message: "unknown Twitch channel"}, nil
		case errors.Is(err, twitchuc.ErrInvalidChannelName):
			return &gen.ErrorMessage{Message: "invalid channel name"}, nil
		default:
			h.obs.LogError(ctx, span, "resolve channel failed", err, zap.String("name", req.Name))
			return nil, err
		}
	}

	u, err := h.sett.CreateTwitchUser(ctx, resolved.ID, resolved.Username)
	if err != nil {
		h.obs.LogError(ctx, span, "create twitch user failed", err, zap.String("username", resolved.Username))
		return nil, err
	}

	h.twitch.ReconcileIRCJoins(ctx)

	tu := entityTwitchUserToGen(u)

	return &tu, nil
}
