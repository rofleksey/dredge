package httptransport

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetTwitchUserProfile(ctx context.Context, req *gen.GetTwitchUserProfileRequest) (gen.GetTwitchUserProfileRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_twitch_user_profile")
	defer span.End()

	u, n, presenceSec, accountCreated, profileImageURL, monitoredFollows, gqlFollows, blacklist, err := h.twitch.GetTwitchUserProfile(ctx, req.GetID())
	if err != nil {
		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return &gen.ErrorMessage{Message: "twitch user not found"}, nil
		}

		h.obs.LogError(ctx, span, "get twitch user profile failed", err, zap.Int64("id", req.GetID()))
		return nil, err
	}

	prof := gen.TwitchUserProfile{
		ID:                      u.ID,
		Username:                u.Username,
		Monitored:               u.Monitored,
		Marked:                  u.Marked,
		IsSus:                   u.IsSus,
		SusType:                 optNilStringFromPtr(u.SusType),
		SusDescription:          optNilStringFromPtr(u.SusDescription),
		SusAutoSuppressed:       u.SusAutoSuppressed,
		MessageCount:            n,
		PresenceSecondsThisWeek: presenceSec,
		ChannelBlacklist:        append([]string(nil), blacklist...),
		IrcOnlyWhenLive:         u.IrcOnlyWhenLive,
		NotifyOffStreamMessages: u.NotifyOffStreamMessages,
		NotifyStreamStart:       u.NotifyStreamStart,
	}

	if profileImageURL != nil && *profileImageURL != "" {
		prof.SetProfileImageURL(gen.NewOptNilString(*profileImageURL))
	} else {
		var z gen.OptNilString
		z.SetToNull()
		prof.SetProfileImageURL(z)
	}

	if accountCreated != nil {
		prof.SetAccountCreatedAt(gen.NewOptNilDateTime(*accountCreated))
	} else {
		var z gen.OptNilDateTime
		z.SetToNull()
		prof.SetAccountCreatedAt(z)
	}

	followsGen := make([]gen.FollowedMonitoredChannel, 0, len(monitoredFollows))

	for _, f := range monitoredFollows {
		fc := gen.FollowedMonitoredChannel{
			ChannelID:    f.ChannelTwitchUserID,
			ChannelLogin: f.ChannelLogin,
		}

		if f.FollowedAt != nil {
			fc.SetFollowedAt(gen.NewOptNilDateTime(*f.FollowedAt))
		} else {
			var fa gen.OptNilDateTime
			fa.SetToNull()
			fc.SetFollowedAt(fa)
		}

		followsGen = append(followsGen, fc)
	}

	prof.SetFollowedMonitoredChannels(followsGen)

	blSet := make(map[string]struct{}, len(blacklist))

	for _, login := range blacklist {
		blSet[strings.ToLower(login)] = struct{}{}
	}

	fcFull := make([]gen.FollowedChannelEntry, 0, len(gqlFollows))

	for _, row := range gqlFollows {
		_, onBL := blSet[strings.ToLower(row.FollowedChannelLogin)]

		entry := gen.FollowedChannelEntry{
			ChannelID:    row.FollowedChannelID,
			ChannelLogin: row.FollowedChannelLogin,
			OnBlacklist:  onBL,
		}

		if row.FollowedAt != nil {
			entry.SetFollowedAt(gen.NewOptNilDateTime(*row.FollowedAt))
		} else {
			var fa gen.OptNilDateTime
			fa.SetToNull()
			entry.SetFollowedAt(fa)
		}

		fcFull = append(fcFull, entry)
	}

	prof.SetFollowedChannels(fcFull)

	return &prof, nil
}
