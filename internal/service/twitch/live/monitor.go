package live

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"

	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

const sendMessageTimeout = 45 * time.Second

// StartMonitor connects the anonymous IRC client and ingests chat for monitored channels.
func (r *Runtime) StartMonitor(ctx context.Context) error {
	ctx, span := r.obs.StartSpan(ctx, "service.twitch.start_monitor")
	defer span.End()

	r.obs.Logger.Debug("start twitch monitor")

	channels, err := r.repo.ListMonitoredTwitchUsers(ctx)
	if err != nil {
		r.obs.LogError(ctx, span, "list monitored twitch users failed", err)
		return err
	}

	rules, err := r.repo.ListRules(ctx)
	if err != nil {
		r.obs.LogError(ctx, span, "list rules failed", err)
		return err
	}

	compiled, bad := compileRules(rules)
	for _, e := range bad {
		r.obs.LogError(ctx, span, "compile monitor rule regex failed", e)
	}

	client := twitchirc.NewAnonymousClient()
	client.Capabilities = []string{twitchirc.TagsCapability, twitchirc.CommandsCapability, twitchirc.MembershipCapability}

	r.attachIRCMonitorAppHandlers(client)
	r.attachIRCMonitorDebug(client)

	if err := r.repo.TruncateChannelChatters(ctx); err != nil {
		r.obs.Logger.Warn("truncate channel chatters failed", zap.Error(err))
	}

	client.OnPrivateMessage(func(msg twitchirc.PrivateMessage) {
		ch := NormalizeTwitchChannel(msg.Channel)
		if ch == "" {
			return
		}

		chatterLogin := strings.ToLower(strings.TrimSpace(msg.User.Name))

		keyword := false

		for i := range compiled {
			if compiled[i].matches(chatterLogin, ch, msg.Message) {
				keyword = true
				break
			}
		}

		badgeTags := badgeTagsFromIRC(msg.User)

		ts := msg.Time
		if ts.IsZero() {
			ts = time.Now().UTC()
		} else {
			ts = ts.UTC()
		}

		persistCtx, cancel := context.WithTimeout(r.persistContext(), 5*time.Second)
		defer cancel()

		var chatterID *int64

		if tid, err := strconv.ParseInt(msg.User.ID, 10, 64); err == nil && tid > 0 {
			inserted, err := r.repo.UpsertTwitchUserFromChat(persistCtx, tid, chatterLogin)
			if err != nil {
				r.obs.Logger.Warn("upsert chatter from irc failed", zap.Error(err), zap.String("channel", ch))
			} else {
				chatterID = &tid
				if inserted && r.onEnqueue != nil {
					r.onEnqueue(tid)
				}
			}
		}

		_, err := r.repo.InsertChatMessageForChannelLogin(persistCtx, ch, chatterID, chatterLogin, msg.Message, keyword, "irc", badgeTags)
		if err != nil {
			r.obs.Logger.Warn("persist chat message failed", zap.Error(err), zap.String("channel", ch))
		}

		if keyword {
			go r.dispatchRuleHitNotifications(context.Background(), ch, chatterLogin, msg.Message)
		}

		var chatterMarked bool

		if chatterID != nil {
			if m, err := r.repo.IsTwitchUserMarked(persistCtx, *chatterID); err == nil {
				chatterMarked = m
			}
		}

		wsPayload := map[string]any{
			"type":           "chat_message",
			"channel":        ch,
			"user":           chatterLogin,
			"message":        msg.Message,
			"keyword_match":  keyword,
			"chatter_marked": chatterMarked,
			"badge_tags":     badgeTags,
			"created_at":     ts.Format(time.RFC3339Nano),
		}
		if chatterID != nil {
			wsPayload["user_twitch_id"] = *chatterID
		}

		r.broadcaster.BroadcastJSON(wsPayload)
	})

	for _, c := range channels {
		client.Join(c.Username)
	}

	r.monitorMu.Lock()
	r.monitorClient = client
	r.monitorMu.Unlock()

	go func() { _ = client.Connect() }()

	return nil
}

// RestartMonitor stops the IRC client (if any) and starts it again with current DB state.
func (r *Runtime) RestartMonitor(ctx context.Context) error {
	r.StopMonitor()
	return r.StartMonitor(ctx)
}

// StopMonitor disconnects the anonymous IRC monitor client, if running.
func (r *Runtime) StopMonitor() {
	r.monitorMu.Lock()
	defer r.monitorMu.Unlock()

	if r.monitorClient != nil {
		_ = r.monitorClient.Disconnect()
		r.monitorClient = nil
	}

	r.ircMonitorMu.Lock()
	r.ircMonitorTCP = false
	r.ircChannelOK = nil
	r.ircMonitorMu.Unlock()
}

// SendMessage sends a chat line using a linked OAuth account (Helix Send Chat Message).
func (r *Runtime) SendMessage(ctx context.Context, accountID int64, channel, message string) error {
	ctx, span := r.obs.StartSpan(ctx, "service.twitch.send_message")
	defer span.End()

	r.obs.Logger.Debug("send twitch message", zap.Int64("account_id", accountID), zap.String("channel", channel))

	sendCtx, cancel := context.WithTimeout(ctx, sendMessageTimeout)
	defer cancel()

	acc, err := r.repo.GetTwitchAccountByID(sendCtx, accountID)
	if err != nil {
		r.obs.LogError(sendCtx, span, "load twitch account failed", err, zap.Int64("account_id", accountID))
		return err
	}

	accessToken, newRefreshToken, err := r.helix.CachedUserAccessTokenForAccount(sendCtx, accountID, acc.RefreshToken)
	if err != nil {
		r.obs.LogError(sendCtx, span, "refresh access token failed", err, zap.Int64("account_id", accountID))
		return err
	}

	if newRefreshToken != "" && newRefreshToken != acc.RefreshToken {
		_ = r.repo.UpdateTwitchRefreshToken(sendCtx, acc.ID, newRefreshToken)
	}

	targetCh := NormalizeTwitchChannel(channel)
	if targetCh == "" {
		return fmt.Errorf("empty channel")
	}

	var broadcasterID int64

	if bid, ok, err := r.repo.MonitoredChannelTwitchUserID(sendCtx, targetCh); err != nil {
		r.obs.LogError(sendCtx, span, "monitored channel twitch id lookup failed", err, zap.String("channel", targetCh))
		return err
	} else if ok {
		broadcasterID = bid
	} else {
		resolved, err := r.helix.ResolveChannel(sendCtx, targetCh)
		if err != nil {
			r.obs.LogError(sendCtx, span, "resolve channel failed", err, zap.String("channel", targetCh))
			return err
		}

		broadcasterID = resolved.ID
	}

	err = r.helix.SendChatMessage(sendCtx, accessToken, broadcasterID, acc.ID, message)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("%w: %w", helix.ErrSendChatTimeout, err)
		}

		r.obs.LogError(sendCtx, span, "helix send chat failed", err, zap.String("channel", targetCh))
		return err
	}

	return nil
}
