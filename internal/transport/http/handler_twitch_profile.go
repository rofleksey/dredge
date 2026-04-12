package httptransport

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	twitchsvc "github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func (h *Handler) GetTwitchUserProfile(ctx context.Context, req *gen.GetTwitchUserProfileRequest) (gen.GetTwitchUserProfileRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_twitch_user_profile")
	defer span.End()

	u, n, presenceSec, accountCreated, monitoredFollows, gqlFollows, blacklist, err := h.twitch.GetTwitchUserProfile(ctx, req.GetID())
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

func (h *Handler) GetChannelLive(ctx context.Context, req *gen.GetChannelLiveRequest) (gen.GetChannelLiveRes, error) {
	info, chatterCount, err := h.twitch.GetChannelLive(ctx, req.Login)
	if err != nil {
		if errors.Is(err, twitchsvc.ErrInvalidChannelName) || errors.Is(err, twitchsvc.ErrUnknownTwitchChannel) {
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

func (h *Handler) GetTwitchUserActivityTimeline(ctx context.Context, req *gen.GetTwitchUserActivityTimelineRequest) (gen.GetTwitchUserActivityTimelineRes, error) {
	ctx, span := h.obs.StartSpan(ctx, "handler.get_twitch_user_activity_timeline")
	defer span.End()

	if _, err := h.twitch.GetTwitchUser(ctx, req.GetID()); err != nil {
		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return &gen.ErrorMessage{Message: "twitch user not found"}, nil
		}
		return nil, err
	}

	to := time.Now().UTC()
	if req.To.IsSet() {
		to = req.To.Value
	}

	from := to.Add(-7 * 24 * time.Hour)
	if req.From.IsSet() {
		from = req.From.Value
	}

	segs, err := h.twitch.GetUserActivityTimeline(ctx, req.GetID(), from, to)
	if err != nil {
		h.obs.LogError(ctx, span, "timeline failed", err)
		return nil, err
	}

	tl := make(gen.GetTwitchUserActivityTimelineOKApplicationJSON, 0, len(segs))
	for _, seg := range segs {
		tl = append(tl, gen.ActivityTimelineSegment{
			ChannelID:    seg.ChannelTwitchUserID,
			ChannelLogin: seg.ChannelLogin,
			Start:        seg.Start,
			End:          seg.End,
		})
	}

	return &tl, nil
}
