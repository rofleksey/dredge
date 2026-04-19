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
	// go-twitch-irc Client stores a single OnConnect func; attachIRCMonitorDebug must not register
	// another OnConnect after this or ircMonitorTCP will never flip true (Settings IRC status breaks).
	client.OnConnect(func() {
		r.ircMonitorMu.Lock()
		r.ircMonitorTCP = true
		r.ircMonitorMu.Unlock()
		r.obs.Logger.Debug("irc monitor: on_connect")
		r.broadcastIRCMonitorTCP(true)
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
	if r.onEnqueue != nil {
		r.onEnqueue(uid)
	}

	ev := entity.UserActivityChatOffline
	if join {
		ev = entity.UserActivityChatOnline
	}

	if join {
		_, upErr := r.repo.UpsertChannelChatterPresence(persistCtx, chID, uid)
		if upErr != nil {
			return
		}
		_ = r.repo.InsertUserActivityEvent(persistCtx, uid, ev, &chID, nil)

		return
	}

	since, hadRow, delErr := r.repo.DeleteChannelChatterPresence(persistCtx, chID, uid)
	if delErr != nil || !hadRow {
		return
	}
	_ = r.repo.InsertUserActivityEvent(persistCtx, uid, ev, &chID, nil)

	_ = since
}

// GetIrcMonitorStatus returns TCP connection and per-channel join state from Join/Depart calls
// (see applyJoinDiffs); it does not infer joins from IRC userlists or JOIN/PART events.
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
	r.ircMonitorMu.Unlock()

	if !clientUp {
		tcp = false
	}

	r.joinStateMu.RLock()

	out := make([]IRCMonitorChannelStatus, 0, len(monitored))

	for _, u := range monitored {
		login := NormalizeTwitchChannel(u.Username)
		ok := tcp && clientUp && login != "" && r.reconcilerJoined[login]
		out = append(out, IRCMonitorChannelStatus{Login: u.Username, IrcOK: ok})
	}
	r.joinStateMu.RUnlock()

	return tcp, out, nil
}

// LiveWebSocketWelcomePayloads returns one JSON message with IRC monitor TCP + per-channel join state for new browser clients.
func (r *Runtime) LiveWebSocketWelcomePayloads(ctx context.Context) (any, error) {
	tcp, rows, err := r.GetIrcMonitorStatus(ctx)
	if err != nil {
		return nil, err
	}

	joined := make([]string, 0, len(rows))
	for _, row := range rows {
		if !row.IrcOK {
			continue
		}

		ch := NormalizeTwitchChannel(row.Login)
		if ch != "" {
			joined = append(joined, ch)
		}
	}

	return map[string]any{
		"type":            "irc_monitor_snapshot",
		"tcp_connected":   tcp,
		"joined_channels": joined,
	}, nil
}
