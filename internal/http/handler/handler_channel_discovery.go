package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) GetChannelDiscoverySettings(ctx context.Context) (*gen.ChannelDiscoverySettings, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_channel_discovery_settings")
	defer span.End()

	s, err := h.sett.GetChannelDiscoverySettings(ctx)
	if err != nil {
		h.obs.LogError(ctx, span, "get channel discovery settings failed", err)
		return nil, err
	}

	out := channelDiscoveryEntityToGen(s)
	return out, nil
}

func (h *Handler) UpdateChannelDiscoverySettings(ctx context.Context, req *gen.ChannelDiscoverySettings) (gen.UpdateChannelDiscoverySettingsRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.update_channel_discovery_settings")
	defer span.End()

	in := channelDiscoveryGenToEntity(req)

	out, err := h.sett.UpdateChannelDiscoverySettings(ctx, in)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidChannelDiscoverySettings) {
			return &gen.ErrorMessage{Message: "when enabled is true, game_id must be non-empty"}, nil
		}

		h.obs.LogError(ctx, span, "update channel discovery settings failed", err)
		return nil, err
	}

	return channelDiscoveryEntityToGen(out), nil
}

func (h *Handler) ListChannelDiscoveryCandidates(ctx context.Context) ([]gen.DiscoveryCandidate, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.list_channel_discovery_candidates")
	defer span.End()

	list, err := h.sett.ListDiscoveryCandidates(ctx)
	if err != nil {
		h.obs.LogError(ctx, span, "list channel discovery candidates failed", err)
		return nil, err
	}

	out := make([]gen.DiscoveryCandidate, 0, len(list))
	for _, c := range list {
		out = append(out, discoveryCandidateEntityToGen(c))
	}

	return out, nil
}

func (h *Handler) ApproveChannelDiscoveryCandidate(ctx context.Context, params gen.ApproveChannelDiscoveryCandidateParams) (gen.ApproveChannelDiscoveryCandidateRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.approve_channel_discovery_candidate")
	defer span.End()

	u, err := h.sett.ApproveDiscoveryCandidate(ctx, params.TwitchUserID)
	if err != nil {
		if errors.Is(err, entity.ErrDiscoveryCandidateNotFound) {
			return &gen.ApproveChannelDiscoveryCandidateNotFound{Message: "discovery candidate not found"}, nil
		}

		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return &gen.ApproveChannelDiscoveryCandidateNotFound{Message: "twitch user not found"}, nil
		}

		if errors.Is(err, entity.ErrInvalidTwitchUserMonitorSettings) {
			return &gen.ApproveChannelDiscoveryCandidateBadRequest{
				Message: "notify_off_stream_messages is only allowed when irc_only_when_live is false",
			}, nil
		}

		h.obs.LogError(ctx, span, "approve discovery candidate failed", err, zap.Int64("twitch_user_id", params.TwitchUserID))
		return nil, err
	}

	h.twitch.ReconcileIRCJoins(ctx)
	h.twitch.EnqueueUserEnrichment(params.TwitchUserID)

	tu := entityTwitchUserToGen(u)
	return &tu, nil
}

func (h *Handler) DenyChannelDiscoveryCandidate(ctx context.Context, params gen.DenyChannelDiscoveryCandidateParams) (gen.DenyChannelDiscoveryCandidateRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.deny_channel_discovery_candidate")
	defer span.End()

	err := h.sett.DenyDiscoveryCandidate(ctx, params.TwitchUserID)
	if err != nil {
		if errors.Is(err, entity.ErrDiscoveryCandidateNotFound) {
			return &gen.ErrorMessage{Message: "discovery candidate not found"}, nil
		}

		h.obs.LogError(ctx, span, "deny discovery candidate failed", err, zap.Int64("twitch_user_id", params.TwitchUserID))
		return nil, err
	}

	return &gen.DenyChannelDiscoveryCandidateNoContent{}, nil
}
