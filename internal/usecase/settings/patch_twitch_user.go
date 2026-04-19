package settings

import (
	"context"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Usecase) PatchTwitchUser(ctx context.Context, id int64, patch entity.TwitchUserPatch) (entity.TwitchUser, error) {
	ctx, span := s.obs.StartSpan(ctx, "usecase.settings.patch_twitch_user")
	defer span.End()

	cur, err := s.repo.GetTwitchUserByID(ctx, id)
	if err != nil {
		s.obs.LogError(ctx, span, "patch twitch user load failed", err, zap.Int64("id", id))
		return entity.TwitchUser{}, err
	}

	// Turning on live-only clears off-stream notifications; apply before validating merged state.
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

	out, err := s.repo.PatchTwitchUser(ctx, id, patch)
	if err != nil {
		s.obs.LogError(ctx, span, "patch twitch user failed", err, zap.Int64("id", id))
	}

	return out, err
}
