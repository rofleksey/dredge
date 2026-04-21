package settings

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) GetChannelDiscoverySettings(ctx context.Context) (entity.ChannelDiscoverySettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.get_channel_discovery_settings")
	defer span.End()

	out, err := s.repo.GetChannelDiscoverySettings(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "get channel discovery settings failed", err)
	}

	return out, err
}

func (s *Usecase) UpdateChannelDiscoverySettings(ctx context.Context, in entity.ChannelDiscoverySettings) (entity.ChannelDiscoverySettings, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.update_channel_discovery_settings")
	defer span.End()

	norm := normalizeChannelDiscoverySettings(in)

	if norm.Enabled && strings.TrimSpace(norm.GameID) == "" {
		return entity.ChannelDiscoverySettings{}, entity.ErrInvalidChannelDiscoverySettings
	}

	if err := s.repo.UpdateChannelDiscoverySettings(ctx, norm); err != nil {
		s.obs.LogError(ctx, span, "update channel discovery settings failed", err)
		return entity.ChannelDiscoverySettings{}, err
	}

	return s.repo.GetChannelDiscoverySettings(ctx)
}

func (s *Usecase) ListDiscoveryCandidates(ctx context.Context) ([]entity.TwitchDiscoveryCandidate, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.list_discovery_candidates")
	defer span.End()

	out, err := s.repo.ListTwitchDiscoveryCandidates(ctx)
	if err != nil {
		s.obs.LogError(ctx, span, "list discovery candidates failed", err)
	}

	return out, err
}

func (s *Usecase) ApproveDiscoveryCandidate(ctx context.Context, twitchUserID int64) (entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.approve_discovery_candidate")
	defer span.End()

	cur, err := s.repo.GetTwitchUserByID(ctx, twitchUserID)
	if err != nil {
		s.obs.LogError(ctx, span, "approve discovery load user failed", err, zap.Int64("id", twitchUserID))
		return entity.TwitchUser{}, err
	}

	mon := true

	patch := entity.TwitchUserPatch{Monitored: &mon}

	if patch.IrcOnlyWhenLive != nil && *patch.IrcOnlyWhenLive {
		f := false
		patch.NotifyOffStreamMessages = &f
	}

	effIrcOnly := cur.IrcOnlyWhenLive
	if patch.IrcOnlyWhenLive != nil {
		effIrcOnly = *patch.IrcOnlyWhenLive
	}

	effNotifyOff := cur.NotifyOffStreamMessages
	if patch.NotifyOffStreamMessages != nil {
		effNotifyOff = *patch.NotifyOffStreamMessages
	}

	if effNotifyOff && effIrcOnly {
		return entity.TwitchUser{}, entity.ErrInvalidTwitchUserMonitorSettings
	}

	out, err := s.repo.ApproveDiscoveryCandidate(ctx, twitchUserID)
	if err != nil {
		s.obs.LogError(ctx, span, "approve discovery candidate failed", err, zap.Int64("id", twitchUserID))
	}

	return out, err
}

func (s *Usecase) DenyDiscoveryCandidate(ctx context.Context, twitchUserID int64) error {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.deny_discovery_candidate")
	defer span.End()

	err := s.repo.DenyDiscoveryCandidate(ctx, twitchUserID)
	if err != nil {
		s.obs.LogError(ctx, span, "deny discovery candidate failed", err, zap.Int64("id", twitchUserID))
	}

	return err
}
