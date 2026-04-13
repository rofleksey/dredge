package httptransport

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	twitchsvc "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) ListTwitchUsers(ctx context.Context) ([]gen.TwitchUser, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_twitch_users")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	list, err := h.sett.ListTwitchUsers(ctx)
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

func (h *Handler) CreateTwitchUser(ctx context.Context, req *gen.CreateTwitchUserRequest) (gen.CreateTwitchUserRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.create_twitch_user")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	resolved, err := h.twitch.ResolveChannel(ctx, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, twitchsvc.ErrUnknownTwitchChannel):
			return &gen.ErrorMessage{Message: "unknown Twitch channel"}, nil
		case errors.Is(err, twitchsvc.ErrInvalidChannelName):
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

	if err := h.twitch.RestartMonitor(ctx); err != nil {
		h.obs.LogError(ctx, span, "restart monitor failed", err, zap.String("username", resolved.Username))
		return nil, err
	}

	tu := entityTwitchUserToGen(u)

	return &tu, nil
}

func (h *Handler) UpdateTwitchUser(ctx context.Context, req *gen.UpdateTwitchUserPostRequest) (gen.UpdateTwitchUserRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.update_twitch_user")
	defer span.End()

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	patch := entity.TwitchUserPatch{}

	if req.Monitored.IsSet() {
		v := req.Monitored.Value
		patch.Monitored = &v
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
			return &gen.UpdateTwitchUserBadRequest{Message: "notify_off_stream_messages requires irc_only_when_live"}, nil
		}

		h.obs.LogError(ctx, span, "update twitch user failed", err, zap.Int64("id", req.ID))
		return nil, err
	}

	restartIRC := patch.Monitored != nil || patch.IrcOnlyWhenLive != nil ||
		patch.NotifyOffStreamMessages != nil || patch.NotifyStreamStart != nil
	if restartIRC {
		if err := h.twitch.RestartMonitor(ctx); err != nil {
			h.obs.LogError(ctx, span, "restart monitor failed", err, zap.Int64("id", req.ID))
			return nil, err
		}
	}

	tu := entityTwitchUserToGen(u)

	return &tu, nil
}
