package live

import (
	"context"
	"strings"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"

	"github.com/rofleksey/dredge/internal/entity"
	"go.uber.org/zap"
)

// isGoTwitchIRCUserlistMissing reports whether err is go-twitch-irc's error for a channel
// that has no internal userlist entry (never joined, departed, or key mismatch). An actual
// empty NAMES result is an empty slice with a nil error; we treat this error like empty NAMES
// and sync an empty chatter snapshot.
func isGoTwitchIRCUserlistMissing(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "could not find userlist for channel") && strings.Contains(msg, "in client")
}

// StartPresenceTicker runs until ctx is cancelled; periodically syncs IRC NAMES lists into channel_chatters.
func (r *Runtime) StartPresenceTicker(ctx context.Context) {
	t := time.NewTicker(r.channelChattersSyncPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			snapshotCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			r.runPresenceSnapshot(snapshotCtx)
			cancel()
		}
	}
}

func (r *Runtime) runPresenceSnapshot(ctx context.Context) {
	r.monitorMu.Lock()
	client := r.monitorClient
	r.monitorMu.Unlock()

	if client == nil {
		return
	}

	channels, err := r.repo.ListMonitoredTwitchUsers(ctx)
	if err != nil {
		r.obs.Logger.Warn("presence: list monitored failed", zap.Error(err))
		return
	}

	for _, ch := range channels {
		login := NormalizeTwitchChannel(ch.Username)
		if login == "" {
			continue
		}
		if err := r.snapshotChannelPresence(ctx, client, ch, login); err != nil {
			r.obs.Logger.Debug("presence snapshot channel skipped", zap.String("channel", login), zap.Error(err))
		}
	}
}

func (r *Runtime) snapshotChannelPresence(ctx context.Context, client *twitchirc.Client, channel entity.TwitchUser, ircLogin string) error {
	logins, err := client.Userlist(ircLogin)
	if err != nil {
		if !isGoTwitchIRCUserlistMissing(err) {
			return err
		}
		logins = nil
	}

	prev, err := r.repo.ListChannelChatterIDs(ctx, channel.ID)
	if err != nil {
		return err
	}

	prevSet := make(map[int64]struct{}, len(prev))
	for _, id := range prev {
		prevSet[id] = struct{}{}
	}

	loginBatch := append([]string{}, logins...)

	idByLogin, err := r.helix.HelixUsersByLogins(ctx, loginBatch)
	if err != nil {
		return err
	}

	curr := make([]int64, 0, len(idByLogin))
	currSet := make(map[int64]struct{})

	for _, login := range logins {
		ln := NormalizeTwitchChannel(login)
		if ln == "" {
			continue
		}

		id, ok := idByLogin[ln]
		if !ok {
			continue
		}

		if _, dup := currSet[id]; dup {
			continue
		}

		// channel_chatters and activity rows FK to twitch_users; Helix IDs are not inserted elsewhere for NAMES-only users.
		if _, err := r.repo.UpsertTwitchUserFromChat(ctx, id, ln); err != nil {
			return err
		}

		currSet[id] = struct{}{}
		curr = append(curr, id)
	}

	chID := channel.ID

	if len(prevSet) == 0 && len(currSet) > 0 {
		if err := r.repo.ReplaceChannelChattersSnapshot(ctx, chID, curr); err != nil {
			return err
		}
		return nil
	}

	return r.repo.ReplaceChannelChattersSnapshot(ctx, chID, curr)
}
