package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func (h *Handler) UpdateTwitchUser(ctx context.Context, req *gen.UpdateTwitchUserPostRequest) (gen.UpdateTwitchUserRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.update_twitch_user")
	defer span.End()

	patch := entity.TwitchUserPatch{}
	enqueueEnrichmentOnMonitoredChange := false

	if req.Monitored.IsSet() {
		v := req.Monitored.Value
		patch.Monitored = &v

		before, err := h.twitch.GetTwitchUser(ctx, req.ID)
		if err != nil {
			if errors.Is(err, entity.ErrTwitchUserNotFound) {
				return &gen.UpdateTwitchUserNotFound{Message: "twitch user not found"}, nil
			}

			h.obs.LogError(ctx, span, "load twitch user before monitored update failed", err, zap.Int64("id", req.ID))
			return nil, err
		}

		enqueueEnrichmentOnMonitoredChange = before.Monitored != v
	}

	if req.Marked.IsSet() {
		v := req.Marked.Value
		patch.Marked = &v
	}

	if req.IsSus.IsSet() {
		v := req.IsSus.Value
		patch.IsSus = &v
	}

	if req.SusType.IsSet() {
		if req.SusType.IsNull() {
			empty := ""
			patch.SusType = &empty
		} else {
			v := req.SusType.Value
			patch.SusType = &v
		}
	}

	if req.SusDescription.IsSet() {
		if req.SusDescription.IsNull() {
			empty := ""
			patch.SusDescription = &empty
		} else {
			v := req.SusDescription.Value
			patch.SusDescription = &v
		}
	}

	if req.SusAutoSuppressed.IsSet() {
		v := req.SusAutoSuppressed.Value
		patch.SusAutoSuppressed = &v
	}

	if req.IrcOnlyWhenLive.IsSet() {
		v := req.IrcOnlyWhenLive.Value
		patch.IrcOnlyWhenLive = &v
	}

	if req.NotifyOffStreamMessages.IsSet() {
		v := req.NotifyOffStreamMessages.Value
		patch.NotifyOffStreamMessages = &v
	}

	if req.NotifyStreamStart.IsSet() {
		v := req.NotifyStreamStart.Value
		patch.NotifyStreamStart = &v
	}

	u, err := h.sett.PatchTwitchUser(ctx, req.ID, patch)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return &gen.UpdateTwitchUserNotFound{Message: "twitch user not found"}, nil
		}

		if errors.Is(err, entity.ErrInvalidTwitchUserMonitorSettings) {
			return &gen.UpdateTwitchUserBadRequest{Message: "notify_off_stream_messages is only allowed when irc_only_when_live is false"}, nil
		}

		h.obs.LogError(ctx, span, "update twitch user failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	if patch.Monitored != nil || patch.IrcOnlyWhenLive != nil {
		h.twitch.ReconcileIRCJoins(ctx)
	}

	if enqueueEnrichmentOnMonitoredChange {
		h.twitch.EnqueueUserEnrichment(req.ID)
	}

	if twitchuc.PatchTouchesSuspicionFields(patch) {
		h.twitch.BroadcastTwitchUserSuspicion(u)
	}

	tu := entityTwitchUserToGen(u)

	return &tu, nil
}
