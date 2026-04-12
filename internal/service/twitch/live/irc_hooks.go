package live

import (
	"context"
	"strings"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"

	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// IRCMonitorChannelStatus is one monitored channel row for the settings UI.
type IRCMonitorChannelStatus struct {
	Login string
	IrcOK bool
}

func (r *Runtime) attachIRCMonitorAppHandlers(client *twitchirc.Client) {
	client.OnConnect(func() {
		r.ircMonitorMu.Lock()
		r.ircMonitorTCP = true
		r.ircChannelOK = make(map[string]bool)
		r.ircMonitorMu.Unlock()
	})

	client.OnSelfJoinMessage(func(m twitchirc.UserJoinMessage) {
		ch := NormalizeTwitchChannel(m.Channel)
		if ch == "" {
			return
		}

		r.ircMonitorMu.Lock()
		if r.ircChannelOK != nil {
			r.ircChannelOK[ch] = true
		}
		r.ircMonitorMu.Unlock()
	})

	client.OnSelfPartMessage(func(m twitchirc.UserPartMessage) {
		ch := NormalizeTwitchChannel(m.Channel)
		if ch == "" {
			return
		}

		r.ircMonitorMu.Lock()
		if r.ircChannelOK != nil {
			r.ircChannelOK[ch] = false
		}
		r.ircMonitorMu.Unlock()
	})

	client.OnUserJoinMessage(func(m twitchirc.UserJoinMessage) {
		ch := NormalizeTwitchChannel(m.Channel)

		u := strings.ToLower(strings.TrimSpace(m.User))
		if ch == "" || u == "" {
			return
		}

		go r.handleIRCChatterPresence(context.Background(), ch, u, true)
	})

	client.OnUserPartMessage(func(m twitchirc.UserPartMessage) {
		ch := NormalizeTwitchChannel(m.Channel)

		u := strings.ToLower(strings.TrimSpace(m.User))
		if ch == "" || u == "" {
			return
		}

		go r.handleIRCChatterPresence(context.Background(), ch, u, false)
	})
}

func (r *Runtime) handleIRCChatterPresence(ctx context.Context, channelLogin, userLogin string, join bool) {
	persistCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	ok, err := r.repo.IsMonitoredChannel(persistCtx, channelLogin)
	if err != nil || !ok {
		return
	}

	chID, err := r.repo.TwitchUserIDByUsername(persistCtx, channelLogin)
	if err != nil {
		return
	}

	idByLogin, err := r.helix.HelixUsersByLogins(persistCtx, []string{userLogin})
	if err != nil {
		r.obs.Logger.Debug("irc join/part helix lookup failed", zap.Error(err), zap.String("channel", channelLogin), zap.String("user", userLogin))
		return
	}

	uid, ok := idByLogin[userLogin]
	if !ok {
		return
	}

	if _, err := r.repo.UpsertTwitchUserFromChat(persistCtx, uid, userLogin); err != nil {
		r.obs.Logger.Debug("irc join/part upsert chatter failed", zap.Error(err), zap.String("user", userLogin))
		return
	}

	ev := entity.UserActivityChatOffline
	if join {
		ev = entity.UserActivityChatOnline
	}

	_ = r.repo.InsertUserActivityEvent(persistCtx, uid, ev, &chID, nil)

	ts := time.Now().UTC()
	createdAt := ts.Format(time.RFC3339Nano)

	if join {
		presentSince, err := r.repo.UpsertChannelChatterPresence(persistCtx, chID, uid)
		if err != nil {
			return
		}

		var accountCreated *time.Time

		if ca, _, err := r.repo.GetHelixMeta(persistCtx, uid); err == nil {
			accountCreated = ca
		}

		payload := map[string]any{
			"type":           "chatter_join",
			"channel":        channelLogin,
			"user":           userLogin,
			"user_twitch_id": uid,
			"present_since":  presentSince.Format(time.RFC3339Nano),
			"created_at":     createdAt,
		}

		if accountCreated != nil {
			payload["account_created_at"] = accountCreated.Format(time.RFC3339Nano)
		}

		r.broadcaster.BroadcastJSON(payload)

		return
	}

	presentSince, hadRow, err := r.repo.DeleteChannelChatterPresence(persistCtx, chID, uid)
	if err != nil || !hadRow {
		return
	}

	sec := int64(ts.Sub(presentSince).Seconds())
	if sec < 0 {
		sec = 0
	}

	r.broadcaster.BroadcastJSON(map[string]any{
		"type":            "chatter_part",
		"channel":         channelLogin,
		"user":            userLogin,
		"user_twitch_id":  uid,
		"present_seconds": sec,
		"present_since":   presentSince.Format(time.RFC3339Nano),
		"created_at":      createdAt,
	})
}

// GetIrcMonitorStatus returns TCP connection and per-channel self-JOIN state (in-memory).
func (r *Runtime) GetIrcMonitorStatus(ctx context.Context) (connected bool, channels []IRCMonitorChannelStatus, err error) {
	monitored, err := r.repo.ListMonitoredTwitchUsers(ctx)
	if err != nil {
		return false, nil, err
	}

	r.monitorMu.Lock()
	clientUp := r.monitorClient != nil
	r.monitorMu.Unlock()

	r.ircMonitorMu.Lock()
	tcp := r.ircMonitorTCP
	chOK := r.ircChannelOK
	r.ircMonitorMu.Unlock()

	if !clientUp {
		tcp = false
	}

	out := make([]IRCMonitorChannelStatus, 0, len(monitored))
	for _, u := range monitored {
		login := NormalizeTwitchChannel(u.Username)
		ok := tcp && chOK != nil && chOK[login]
		out = append(out, IRCMonitorChannelStatus{Login: u.Username, IrcOK: ok})
	}

	return tcp, out, nil
}
