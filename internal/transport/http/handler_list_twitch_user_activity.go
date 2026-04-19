package httptransport

import (
	"context"
	"errors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListTwitchUserActivity(ctx context.Context, req *gen.ListTwitchUserActivityRequest) (gen.ListTwitchUserActivityRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_twitch_user_activity")
	defer span.End()

	u, err := h.twitch.GetTwitchUser(ctx, req.GetID())
	if err != nil {
		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return &gen.ErrorMessage{Message: "twitch user not found"}, nil
		}

		h.obs.LogError(ctx, span, "get twitch user for activity failed", err)
		return nil, err
	}

	limit := 50
	if req.Limit.IsSet() {
		limit = req.Limit.Value
	}

	f := entity.UserActivityListFilter{ChatterUserID: req.GetID(), Limit: limit}

	if req.CursorCreatedAt.IsSet() && req.CursorID.IsSet() {
		t := req.CursorCreatedAt.Value
		f.CursorCreatedAt = &t
		id := req.CursorID.Value
		f.CursorID = &id
	}

	evs, err := h.twitch.ListUserActivity(ctx, f)
	if err != nil {
		h.obs.LogError(ctx, span, "list activity failed", err)
		return nil, err
	}

	out := make(gen.ListTwitchUserActivityOKApplicationJSON, 0, len(evs))

	for _, e := range evs {
		out = append(out, entityActivityToGen(e, u.Username))
	}

	return &out, nil
}
