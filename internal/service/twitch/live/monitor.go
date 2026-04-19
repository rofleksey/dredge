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

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

const sendMessageTimeout = 45 * time.Second

type streamStartNotify struct {
	login string
	title string
}

func (r *Runtime) ircOAuthCredentials(ctx context.Context, accountID int64) (username string, oauthIRC string, acc entity.TwitchAccount, err error) {
	acc, err = r.repo.GetTwitchAccountByID(ctx, accountID)
	if err != nil {
		return "", "", entity.TwitchAccount{}, err
	}

	at, newRT, err := r.helix.CachedUserAccessTokenForAccount(ctx, acc.ID, acc.RefreshToken)
	if err != nil {
		return "", "", acc, err
	}

	if newRT != "" && newRT != acc.RefreshToken {
		_ = r.repo.UpdateTwitchRefreshToken(ctx, acc.ID, newRT)
	}

	return acc.Username, "oauth:" + at, acc, nil
}

func (r *Runtime) buildIRCMonitorClient(ctx context.Context) (client *twitchirc.Client, oauthTokenForSync string, useOAuthSync bool, err error) {
	settings, err := r.repo.GetIrcMonitorSettings(ctx)
	if err != nil {
		return nil, "", false, err
	}

	if settings.OauthTwitchAccountID == nil {
		return twitchirc.NewAnonymousClient(), "", false, nil
	}

	username, oauthIRC, _, err := r.ircOAuthCredentials(ctx, *settings.OauthTwitchAccountID)
	if err != nil {
		return nil, "", false, err
	}

	return twitchirc.NewClient(username, oauthIRC), oauthIRC, true, nil
}

func (r *Runtime) wirePrivateMessageHandlers(client *twitchirc.Client) {
	client.OnPrivateMessage(func(msg twitchirc.PrivateMessage) {
		ch := NormalizeTwitchChannel(msg.Channel)
		if ch == "" {
			return
		}

		chatterLogin := strings.ToLower(strings.TrimSpace(msg.User.Name))

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
			_, err := r.repo.UpsertTwitchUserFromChat(persistCtx, tid, chatterLogin)
			if err != nil {
				r.obs.Logger.Warn("upsert chatter from irc failed", zap.Error(err), zap.String("channel", ch))
			} else {
				chatterID = &tid
				if r.onEnqueue != nil {
					r.onEnqueue(tid)
				}
			}
		}

		keyword := false

		if re := r.ruleEng(); re != nil {
			keyword = re.KeywordMatchChat(persistCtx, ch, chatterLogin, msg.Message)
			re.HandleChatMessage(ch, chatterLogin, msg.Message)
		}

		_, err := r.repo.InsertChatMessageForChannelLogin(persistCtx, ch, chatterID, chatterLogin, msg.Message, keyword, "irc", badgeTags, msg.FirstMessage)
		if err != nil {
			r.obs.Logger.Warn("persist chat message failed", zap.Error(err), zap.String("channel", ch))
		}

		var chatterMarked bool

		var chatterIsSus bool

		if chatterID != nil {
			if m, err := r.repo.IsTwitchUserMarked(persistCtx, *chatterID); err == nil {
				chatterMarked = m
			}
			if s, err := r.repo.IsTwitchUserSuspicious(persistCtx, *chatterID); err == nil {
				chatterIsSus = s
			}
		}

		wsPayload := map[string]any{
			"type":           "chat_message",
			"channel":        ch,
			"user":           chatterLogin,
			"message":        msg.Message,
			"keyword_match":  keyword,
			"chatter_marked": chatterMarked,
			"chatter_is_sus": chatterIsSus,
			"first_message":  msg.FirstMessage,
			"badge_tags":     badgeTags,
			"created_at":     ts.Format(time.RFC3339Nano),
		}
		if chatterID != nil {
			wsPayload["user_twitch_id"] = *chatterID
		}

		r.broadcaster.BroadcastJSON(wsPayload)
	})
}

// StartMonitor connects the IRC client (anonymous or OAuth per settings) and ingests chat for monitored channels (join set reconciled against Helix).
func (r *Runtime) StartMonitor(ctx context.Context) error {
	ctx, span := r.obs.StartSpan(ctx, "service.twitch.start_monitor")
	defer span.End()

	r.obs.Logger.Debug("start twitch monitor")

	client, oauthIRC, useOAuthSync, err := r.buildIRCMonitorClient(ctx)
	if err != nil {
		r.obs.LogError(ctx, span, "irc monitor credentials failed", err)
		return err
	}

	client.Capabilities = []string{twitchirc.TagsCapability, twitchirc.CommandsCapability, twitchirc.MembershipCapability}

	r.attachIRCMonitorAppHandlers(client)
	r.attachIRCMonitorDebug(client)
	r.wirePrivateMessageHandlers(client)

	if err := r.repo.TruncateChannelChatters(ctx); err != nil {
		r.obs.Logger.Warn("truncate channel chatters failed", zap.Error(err))
	}

	r.monitorMu.Lock()
	r.monitorClient = client
	r.monitorMu.Unlock()

	r.joinStateMu.Lock()
	r.reconcilerJoined = make(map[string]bool)
	r.streamEdge = make(map[int64]streamLiveEdge)

	if useOAuthSync {
		r.lastIRCOAuthToken = oauthIRC
	} else {
		r.lastIRCOAuthToken = ""
	}
	r.joinStateMu.Unlock()

	loopCtx, cancel := context.WithCancel(context.Background())

	r.monitorLoopsMu.Lock()
	if r.monitorLoopsCancel != nil {
		r.monitorLoopsCancel()
		r.monitorLoopsWG.Wait()
	}

	r.monitorLoopsCancel = cancel
	r.monitorLoopsMu.Unlock()

	r.monitorLoopsWG.Add(1)

	go func() {
		defer r.monitorLoopsWG.Done()

		r.runJoinReconcileLoop(loopCtx)
	}()

	if useOAuthSync {
		r.monitorLoopsWG.Add(1)

		go func() {
			defer r.monitorLoopsWG.Done()

			r.runOAuthTokenSyncLoop(loopCtx)
		}()
	}

	go func() {
		if err := client.Connect(); err != nil {
			r.obs.Logger.Warn("irc monitor: connect ended", zap.Error(err))
		}
	}()

	return nil
}

func (r *Runtime) runOAuthTokenSyncLoop(ctx context.Context) {
	t := time.NewTicker(r.oauthTokenSyncInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			syncCtx, cancel := context.WithTimeout(r.persistContext(), 45*time.Second)

			settings, err := r.repo.GetIrcMonitorSettings(syncCtx)
			if err != nil || settings.OauthTwitchAccountID == nil {
				cancel()
				continue
			}

			_, oauthIRC, _, err := r.ircOAuthCredentials(syncCtx, *settings.OauthTwitchAccountID)

			cancel()

			if err != nil {
				r.obs.Logger.Debug("irc oauth token sync skipped", zap.Error(err))
				continue
			}

			r.joinStateMu.Lock()

			prev := r.lastIRCOAuthToken
			if oauthIRC != prev {
				r.lastIRCOAuthToken = oauthIRC
			}
			r.joinStateMu.Unlock()

			if oauthIRC == prev {
				continue
			}

			r.monitorMu.Lock()

			cl := r.monitorClient
			if cl != nil {
				cl.SetIRCToken(oauthIRC)
			}
			r.monitorMu.Unlock()
		}
	}
}

func (r *Runtime) runJoinReconcileLoop(ctx context.Context) {
	t := time.NewTicker(r.joinReconcileInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.reconcileIRCJoinsOnce(ctx)
		}
	}
}

func (r *Runtime) reconcileIRCJoinsOnce(ctx context.Context) {
	r.monitorMu.Lock()
	client := r.monitorClient
	r.monitorMu.Unlock()

	if client == nil {
		return
	}

	reconcileCtx, cancel := context.WithTimeout(r.persistContext(), 60*time.Second)
	defer cancel()

	monitored, err := r.repo.ListMonitoredTwitchUsers(reconcileCtx)
	if err != nil {
		r.obs.Logger.Debug("join reconcile: list monitored failed", zap.Error(err))
		return
	}

	if len(monitored) == 0 {
		r.applyJoinDiffs(reconcileCtx, map[string]bool{})
		return
	}

	ids := make([]int64, 0, len(monitored))
	for _, u := range monitored {
		ids = append(ids, u.ID)
	}

	liveMap, err := r.helix.HelixStreamsLiveByBroadcasterIDs(reconcileCtx, ids)
	if err != nil {
		r.obs.Logger.Debug("join reconcile: helix streams failed", zap.Error(err))
		return
	}

	metaByID, err := r.helix.HelixStreamsMetadataByBroadcasterIDs(reconcileCtx, ids)
	if err != nil {
		metaByID = nil
	}

	streamStarts := make([]streamStartNotify, 0)
	streamEnds := make([]string, 0)

	r.joinStateMu.Lock()

	for _, u := range monitored {
		nowLive := liveMap[u.ID]

		edge := r.streamEdge[u.ID]
		if !edge.initialized {
			r.streamEdge[u.ID] = streamLiveEdge{initialized: true, wasLive: nowLive}
			continue
		}

		if edge.initialized && edge.wasLive && !nowLive {
			streamEnds = append(streamEnds, u.Username)
		}

		if u.NotifyStreamStart && !edge.wasLive && nowLive {
			title := ""

			if metaByID != nil {
				if m, ok := metaByID[u.ID]; ok {
					title = m.Title
				}
			}

			streamStarts = append(streamStarts, streamStartNotify{login: u.Username, title: title})
		}

		r.streamEdge[u.ID] = streamLiveEdge{initialized: true, wasLive: nowLive}
	}
	r.joinStateMu.Unlock()

	re := r.ruleEng()

	for _, ev := range streamStarts {
		ev := ev
		if re != nil {
			go func(login, title string) {
				re.HandleStreamStart(login, title)
			}(ev.login, ev.title)
		}
	}

	for _, login := range streamEnds {
		login := login
		if re != nil {
			go func(l string) {
				re.HandleStreamEnd(l)
			}(login)
		}
	}

	want := make(map[string]bool, len(monitored))
	for _, u := range monitored {
		ch := NormalizeTwitchChannel(u.Username)
		if ch == "" {
			continue
		}

		if ircChannelJoinWanted(u, liveMap[u.ID]) {
			want[ch] = true
		}
	}

	r.applyJoinDiffs(reconcileCtx, want)
}

func (r *Runtime) applyJoinDiffs(ctx context.Context, want map[string]bool) {
	r.applyJoinSerialMu.Lock()
	defer r.applyJoinSerialMu.Unlock()

	r.monitorMu.Lock()
	client := r.monitorClient
	r.monitorMu.Unlock()

	if client == nil {
		return
	}

	r.joinStateMu.Lock()

	toLeave := make([]string, 0)

	for ch := range r.reconcilerJoined {
		if !want[ch] {
			toLeave = append(toLeave, ch)
		}
	}

	toJoin := make([]string, 0)

	for ch := range want {
		if !r.reconcilerJoined[ch] {
			toJoin = append(toJoin, ch)
		}
	}

	r.joinStateMu.Unlock()

	// Depart/Join can block on IRC I/O; must not hold joinStateMu so GetIrcMonitorStatus can RLock.
	for _, ch := range toLeave {
		client.Depart(ch)
	}

	for _, ch := range toJoin {
		client.Join(ch)
	}

	r.joinStateMu.Lock()

	for _, ch := range toLeave {
		delete(r.reconcilerJoined, ch)
	}

	for _, ch := range toJoin {
		r.reconcilerJoined[ch] = true
	}

	r.joinStateMu.Unlock()

	for _, ch := range toLeave {
		r.broadcastIRCMonitorPart(ch)
	}

	for _, ch := range toJoin {
		r.broadcastIRCMonitorJoin(ch)
	}

	_ = ctx
}

func (r *Runtime) broadcastIRCMonitorJoin(channel string) {
	if r.broadcaster == nil {
		return
	}

	ch := NormalizeTwitchChannel(channel)
	if ch == "" {
		return
	}

	r.broadcaster.BroadcastJSON(map[string]any{
		"type":    "irc_monitor_join",
		"channel": ch,
	})
}

func (r *Runtime) broadcastIRCMonitorPart(channel string) {
	if r.broadcaster == nil {
		return
	}

	ch := NormalizeTwitchChannel(channel)
	if ch == "" {
		return
	}

	r.broadcaster.BroadcastJSON(map[string]any{
		"type":    "irc_monitor_part",
		"channel": ch,
	})
}

func (r *Runtime) broadcastIRCMonitorTCP(connected bool) {
	if r.broadcaster == nil {
		return
	}

	r.broadcaster.BroadcastJSON(map[string]any{
		"type":      "irc_monitor_tcp",
		"connected": connected,
	})
}

// RestartMonitor stops the IRC client (if any) and starts it again with current DB state.
func (r *Runtime) RestartMonitor(ctx context.Context) error {
	r.StopMonitor()
	return r.StartMonitor(ctx)
}

// StopMonitor disconnects the IRC monitor client, if running.
func (r *Runtime) StopMonitor() {
	r.monitorLoopsMu.Lock()
	if r.monitorLoopsCancel != nil {
		r.monitorLoopsCancel()
		r.monitorLoopsCancel = nil
	}
	r.monitorLoopsMu.Unlock()

	r.monitorLoopsWG.Wait()

	// Wait for any in-flight applyJoinDiffs (e.g. HTTP ReconcileIRCJoins) before tearing down maps/client.
	r.applyJoinSerialMu.Lock()
	defer r.applyJoinSerialMu.Unlock()

	r.monitorMu.Lock()
	if r.monitorClient != nil {
		_ = r.monitorClient.Disconnect()
		r.monitorClient = nil
	}
	r.monitorMu.Unlock()

	r.joinStateMu.Lock()

	prevJoined := make([]string, 0, len(r.reconcilerJoined))
	for ch := range r.reconcilerJoined {
		prevJoined = append(prevJoined, ch)
	}

	r.reconcilerJoined = make(map[string]bool)
	r.streamEdge = make(map[int64]streamLiveEdge)
	r.lastIRCOAuthToken = ""
	r.joinStateMu.Unlock()

	r.ircMonitorMu.Lock()
	r.ircMonitorTCP = false
	r.ircMonitorMu.Unlock()

	for _, ch := range prevJoined {
		r.broadcastIRCMonitorPart(ch)
	}

	r.broadcastIRCMonitorTCP(false)
}

// ReconcileIRCJoins refreshes IRC channel membership from the database and Helix live state
// (Join/Depart diffs) without reconnecting the client. It is a no-op when the monitor client
// is not running.
func (r *Runtime) ReconcileIRCJoins(ctx context.Context) {
	r.reconcileIRCJoinsOnce(ctx)
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
