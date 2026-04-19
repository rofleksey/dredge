package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func (u *Usecase) toolCountTwitchMessages(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	f := entity.ChatMessageListFilter{}
	if s, ok := raw["username"].(string); ok {
		f.Username = s
	}
	if s, ok := raw["text"].(string); ok {
		f.Text = s
	}
	if s, ok := raw["channel"].(string); ok {
		f.Channel = s
	}
	if v, ok := raw["chatter_user_id"]; ok && v != nil {
		if id, err := int64Field(raw, "chatter_user_id"); err == nil {
			f.ChatterUserID = &id
		}
	}
	if s, ok := raw["created_from"].(string); ok && strings.TrimSpace(s) != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad created_from: " + err.Error()}), err
		}
		f.CreatedFrom = &t
	}
	if s, ok := raw["created_to"].(string); ok && strings.TrimSpace(s) != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad created_to: " + err.Error()}), err
		}
		f.CreatedTo = &t
	}
	n, err := u.tw.CountChatMessages(ctx, f)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]int64{"total": n}), nil
}

func (u *Usecase) toolCountTwitchDirectoryUsers(ctx context.Context, args string) (string, error) {
	var p struct {
		Username       string `json:"username"`
		MonitoredOnly  bool   `json:"monitored_only"`
	}
	_ = json.Unmarshal([]byte(args), &p)
	f := entity.TwitchUserBrowseFilter{Username: p.Username, MonitoredOnly: p.MonitoredOnly}
	n, err := u.tw.CountTwitchUsersBrowse(ctx, f)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]int64{"total": n}), nil
}

func (u *Usecase) toolGetChannelLive(ctx context.Context, args string) (string, error) {
	var p struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	info, chatterCount, err := u.tw.GetChannelLive(ctx, p.Login)
	if err != nil {
		if errors.Is(err, twitchuc.ErrInvalidChannelName) || errors.Is(err, twitchuc.ErrUnknownTwitchChannel) {
			return mustJSON(map[string]string{"error": "unknown or invalid channel"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	out := map[string]any{
		"broadcaster_id":    info.BroadcasterID,
		"broadcaster_login": info.BroadcasterLogin,
		"display_name":      info.DisplayName,
		"profile_image_url": info.ProfileImageURL,
		"is_live":           info.IsLive,
		"title":             info.Title,
		"game_name":         info.GameName,
		"viewer_count":      info.ViewerCount,
	}
	if chatterCount != nil {
		out["channel_chatter_count"] = *chatterCount
	}
	return mustJSON(out), nil
}

func (u *Usecase) toolListChannelChatters(ctx context.Context, args string) (string, error) {
	var p struct {
		AccountID          int64  `json:"account_id"`
		Login              string `json:"login"`
		SessionStartedAt   string `json:"session_started_at"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	var sessionAt *time.Time
	if strings.TrimSpace(p.SessionStartedAt) != "" {
		t, err := time.Parse(time.RFC3339, p.SessionStartedAt)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad session_started_at: " + err.Error()}), err
		}
		sessionAt = &t
	}
	list, err := u.tw.ListChannelChatters(ctx, p.AccountID, p.Login, sessionAt)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return mustJSON(map[string]string{"error": "twitch account not found"}), err
		}
		if errors.Is(err, entity.ErrNoTwitchUserForChannel) {
			return mustJSON(map[string]string{"error": "twitch user not linked for this account"}), err
		}
		if errors.Is(err, twitchuc.ErrInvalidChannelName) || errors.Is(err, twitchuc.ErrUnknownTwitchChannel) {
			return mustJSON(map[string]string{"error": "unknown channel"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func (u *Usecase) toolGetIrcMonitorStatus(ctx context.Context) (string, error) {
	connected, channels, err := u.tw.GetIrcMonitorStatus(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]any{"connected": connected, "channels": channels}), nil
}

func (u *Usecase) toolGetWatchUiHints() (string, error) {
	v, c, m := u.tw.WatchUiHints()
	return mustJSON(map[string]int{
		"viewer_poll_sec":          v,
		"channel_chatters_sync_sec": c,
		"monitored_live_poll_sec":  m,
	}), nil
}

func (u *Usecase) toolListMonitoredStreams(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	_ = json.Unmarshal([]byte(args), &raw)
	f := entity.StreamListFilter{}
	if s, ok := raw["channel_login"].(string); ok {
		f.ChannelLogin = s
	}
	if v, ok := raw["limit"]; ok {
		if n, ok := parseFloatInt(v); ok {
			f.Limit = n
		}
	}
	if s, ok := raw["cursor_started_at"].(string); ok && strings.TrimSpace(s) != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad cursor_started_at: " + err.Error()}), err
		}
		f.CursorStartedAt = &t
		if id, ok := raw["cursor_id"]; ok {
			if cid, err := parseAnyInt64(id); err == nil {
				f.CursorID = &cid
			}
		}
	}
	list, err := u.tw.ListMonitoredStreams(ctx, f)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func parseFloatInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	default:
		return 0, false
	}
}

func parseAnyInt64(v any) (int64, error) {
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case int:
		return int64(t), nil
	case int64:
		return t, nil
	case json.Number:
		return t.Int64()
	default:
		return 0, fmt.Errorf("invalid number")
	}
}

func (u *Usecase) toolGetMonitoredStream(ctx context.Context, args string) (string, error) {
	var p struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	st, err := u.tw.GetMonitoredStream(ctx, p.ID)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return mustJSON(map[string]string{"error": "stream not found"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(st), nil
}

func (u *Usecase) toolListStreamMessages(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	streamID, err := int64Field(raw, "stream_id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	limit := 50
	if v, ok := raw["limit"]; ok {
		if n, ok := parseFloatInt(v); ok && n > 0 {
			limit = n
		}
	}
	if limit > 200 {
		limit = 200
	}
	f := entity.ChatMessageListFilter{Limit: limit}
	if s, ok := raw["username"].(string); ok {
		f.Username = s
	}
	if s, ok := raw["text"].(string); ok {
		f.Text = s
	}
	if v, ok := raw["chatter_user_id"]; ok && v != nil {
		if id, err := int64Field(raw, "chatter_user_id"); err == nil {
			f.ChatterUserID = &id
		}
	}
	if s, ok := raw["cursor_created_at"].(string); ok && strings.TrimSpace(s) != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad cursor_created_at: " + err.Error()}), err
		}
		f.CursorCreatedAt = &t
		if id, err := int64Field(raw, "cursor_id"); err == nil {
			f.CursorID = &id
		}
	}
	list, err := u.tw.ListStreamMessages(ctx, streamID, f)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return mustJSON(map[string]string{"error": "stream not found"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func (u *Usecase) toolListStreamActivity(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	streamID, err := int64Field(raw, "stream_id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	limit := 50
	if v, ok := raw["limit"]; ok {
		if n, ok := parseFloatInt(v); ok && n > 0 {
			limit = n
		}
	}
	if limit > 200 {
		limit = 200
	}
	var cursorCreatedAt *time.Time
	var cursorID *int64
	if s, ok := raw["cursor_created_at"].(string); ok && strings.TrimSpace(s) != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad cursor_created_at: " + err.Error()}), err
		}
		cursorCreatedAt = &t
		if id, err := int64Field(raw, "cursor_id"); err == nil {
			cursorID = &id
		}
	}
	list, err := u.tw.ListStreamActivity(ctx, streamID, limit, cursorCreatedAt, cursorID)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return mustJSON(map[string]string{"error": "stream not found"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func (u *Usecase) toolGetStreamLeaderboard(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	streamID, err := int64Field(raw, "stream_id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	st, err := u.tw.GetMonitoredStream(ctx, streamID)
	if err != nil {
		if errors.Is(err, entity.ErrStreamNotFound) {
			return mustJSON(map[string]string{"error": "stream not found"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	sort := entity.StreamLeaderboardSortPresenceDesc
	if s, ok := raw["sort"].(string); ok && strings.TrimSpace(s) != "" {
		sort = entity.StreamLeaderboardSort(s)
	}
	q := ""
	if s, ok := raw["q"].(string); ok {
		q = s
	}
	rows, err := u.tw.StreamLeaderboard(ctx, st, sort, q)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(rows), nil
}

func (u *Usecase) toolGetTwitchUser(ctx context.Context, args string) (string, error) {
	var p struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	uu, err := u.tw.GetTwitchUser(ctx, p.ID)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return mustJSON(map[string]string{"error": "twitch user not found"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(uu), nil
}

func (u *Usecase) toolCountRules(ctx context.Context) (string, error) {
	n, err := u.rules.CountRules(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]int64{"total": n}), nil
}

func (u *Usecase) toolCreateNotification(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	provider := stringField(raw, "provider")
	if provider == "" {
		return mustJSON(map[string]string{"error": "missing provider"}), fmt.Errorf("missing provider")
	}
	settings := mapField(raw, "settings")
	enabled := true
	if v, ok := raw["enabled"].(bool); ok {
		enabled = v
	}
	e, err := u.sett.CreateNotification(ctx, provider, settings, enabled)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(e), nil
}

func (u *Usecase) toolUpdateNotification(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	id, err := int64Field(raw, "id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	var prov *string
	if _, ok := raw["provider"]; ok {
		s := stringField(raw, "provider")
		prov = &s
	}
	var settings map[string]any
	if _, ok := raw["settings"]; ok {
		settings = mapField(raw, "settings")
	}
	var enabled *bool
	if v, ok := raw["enabled"].(bool); ok {
		enabled = &v
	}
	e, err := u.sett.UpdateNotification(ctx, id, prov, settings, enabled)
	if err != nil {
		if errors.Is(err, entity.ErrNotificationNotFound) {
			return mustJSON(map[string]string{"error": "notification not found"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(e), nil
}

func (u *Usecase) toolDeleteNotification(ctx context.Context, args string) (string, error) {
	var p struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	if err := u.sett.DeleteNotification(ctx, p.ID); err != nil {
		s := err.Error()
		if errors.Is(err, entity.ErrNotificationNotFound) {
			s = "notification not found"
		}
		return mustJSON(map[string]string{"error": s}), err
	}
	return mustJSON(map[string]any{"ok": true, "deleted_id": p.ID}), nil
}

func (u *Usecase) toolSetChannelBlacklist(ctx context.Context, args string) (string, error) {
	var p struct {
		Login string `json:"login"`
		Add   bool   `json:"add"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	if err := u.sett.SetChannelBlacklist(ctx, p.Login, p.Add); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]any{"ok": true}), nil
}

func (u *Usecase) toolUpdateSuspicionSettings(ctx context.Context, args string) (string, error) {
	var wire struct {
		AutoCheckAccountAge bool `json:"auto_check_account_age"`
		AccountAgeSusDays   int  `json:"account_age_sus_days"`
		AutoCheckBlacklist  bool `json:"auto_check_blacklist"`
		AutoCheckLowFollows bool `json:"auto_check_low_follows"`
		LowFollowsThreshold int  `json:"low_follows_threshold"`
		MaxGQLFollowPages   int  `json:"max_gql_follow_pages"`
	}
	if err := json.Unmarshal([]byte(args), &wire); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	in := entity.SuspicionSettings{
		AutoCheckAccountAge: wire.AutoCheckAccountAge,
		AccountAgeSusDays:   wire.AccountAgeSusDays,
		AutoCheckBlacklist:  wire.AutoCheckBlacklist,
		AutoCheckLowFollows: wire.AutoCheckLowFollows,
		LowFollowsThreshold: wire.LowFollowsThreshold,
		MaxGQLFollowPages:   wire.MaxGQLFollowPages,
	}
	out, err := u.sett.UpdateSuspicionSettings(ctx, in)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(out), nil
}

func (u *Usecase) toolUpdateIrcMonitorSettings(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	cur, err := u.sett.GetIrcMonitorSettings(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	in := cur
	if _, ok := raw["oauth_twitch_account_id"]; ok {
		if raw["oauth_twitch_account_id"] == nil {
			in.OauthTwitchAccountID = nil
		} else {
			id, err := parseAnyInt64(raw["oauth_twitch_account_id"])
			if err != nil {
				return mustJSON(map[string]string{"error": "bad oauth_twitch_account_id"}), err
			}
			in.OauthTwitchAccountID = &id
		}
	}
	if v, ok := raw["enrichment_cooldown_hours"]; ok {
		hours, err := parseAnyInt64(v)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad enrichment_cooldown_hours"}), err
		}
		in.EnrichmentCooldown = time.Duration(hours) * time.Hour
	}
	out, err := u.sett.UpdateIrcMonitorSettings(ctx, in)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchAccountNotFound) {
			return mustJSON(map[string]string{"error": "twitch account not found"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	if err := u.tw.RestartMonitor(ctx); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(out), nil
}

func (u *Usecase) toolCreateTwitchUser(ctx context.Context, args string) (string, error) {
	var p struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	resolved, err := u.tw.ResolveChannel(ctx, p.Name)
	if err != nil {
		if errors.Is(err, twitchuc.ErrUnknownTwitchChannel) {
			return mustJSON(map[string]string{"error": "unknown Twitch channel"}), err
		}
		if errors.Is(err, twitchuc.ErrInvalidChannelName) {
			return mustJSON(map[string]string{"error": "invalid channel name"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	uu, err := u.sett.CreateTwitchUser(ctx, resolved.ID, resolved.Username)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	u.tw.ReconcileIRCJoins(ctx)
	return mustJSON(uu), nil
}

func optionalBoolPtr(raw map[string]any, key string) *bool {
	v, ok := raw[key]
	if !ok {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return &b
}

func optionalStringPtr(raw map[string]any, key string) *string {
	v, ok := raw[key]
	if !ok || v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

func (u *Usecase) toolPatchTwitchUser(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	id, err := int64Field(raw, "id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	patch := entity.TwitchUserPatch{}
	if v := optionalBoolPtr(raw, "monitored"); v != nil {
		patch.Monitored = v
	}
	if v := optionalBoolPtr(raw, "marked"); v != nil {
		patch.Marked = v
	}
	if v := optionalBoolPtr(raw, "is_sus"); v != nil {
		patch.IsSus = v
	}
	if _, ok := raw["sus_type"]; ok {
		if raw["sus_type"] == nil {
			empty := ""
			patch.SusType = &empty
		} else if s, ok := raw["sus_type"].(string); ok {
			patch.SusType = &s
		}
	}
	if _, ok := raw["sus_description"]; ok {
		if raw["sus_description"] == nil {
			empty := ""
			patch.SusDescription = &empty
		} else if s, ok := raw["sus_description"].(string); ok {
			patch.SusDescription = &s
		}
	}
	if v := optionalBoolPtr(raw, "sus_auto_suppressed"); v != nil {
		patch.SusAutoSuppressed = v
	}
	if v := optionalBoolPtr(raw, "irc_only_when_live"); v != nil {
		patch.IrcOnlyWhenLive = v
	}
	if v := optionalBoolPtr(raw, "notify_off_stream_messages"); v != nil {
		patch.NotifyOffStreamMessages = v
	}
	if v := optionalBoolPtr(raw, "notify_stream_start"); v != nil {
		patch.NotifyStreamStart = v
	}
	var before entity.TwitchUser
	if patch.Monitored != nil {
		var err error
		before, err = u.tw.GetTwitchUser(ctx, id)
		if err != nil {
			if errors.Is(err, entity.ErrTwitchUserNotFound) {
				return mustJSON(map[string]string{"error": "twitch user not found"}), err
			}
			return mustJSON(map[string]string{"error": err.Error()}), err
		}
	}
	uu, err := u.sett.PatchTwitchUser(ctx, id, patch)
	if err != nil {
		if errors.Is(err, entity.ErrTwitchUserNotFound) {
			return mustJSON(map[string]string{"error": "twitch user not found"}), err
		}
		if errors.Is(err, entity.ErrInvalidTwitchUserMonitorSettings) {
			return mustJSON(map[string]string{"error": "notify_off_stream_messages is only allowed when irc_only_when_live is false"}), err
		}
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	if patch.Monitored != nil || patch.IrcOnlyWhenLive != nil {
		u.tw.ReconcileIRCJoins(ctx)
	}
	if patch.Monitored != nil && before.Monitored != *patch.Monitored {
		u.tw.EnqueueUserEnrichment(id)
	}
	if twitchuc.PatchTouchesSuspicionFields(patch) {
		u.tw.BroadcastTwitchUserSuspicion(uu)
	}
	return mustJSON(uu), nil
}

func (u *Usecase) toolCreateTwitchAccount(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	id, err := int64Field(raw, "id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	username := stringField(raw, "username")
	refresh := stringField(raw, "refresh_token")
	accountType := "main"
	if s := stringField(raw, "account_type"); s != "" {
		accountType = s
	}
	a, err := u.sett.CreateTwitchAccount(ctx, id, username, refresh, accountType)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	type acc struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		AccountType string `json:"account_type"`
		CreatedAt   string `json:"created_at"`
	}
	return mustJSON(acc{ID: a.ID, Username: a.Username, AccountType: a.AccountType, CreatedAt: a.CreatedAt.Format(time.RFC3339)}), nil
}

func (u *Usecase) toolPatchTwitchAccount(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	id, err := int64Field(raw, "id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	var at *string
	if s := optionalStringPtr(raw, "account_type"); s != nil {
		at = s
	}
	a, err := u.sett.PatchTwitchAccount(ctx, id, at)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	type acc struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		AccountType string `json:"account_type"`
		CreatedAt   string `json:"created_at"`
	}
	return mustJSON(acc{ID: a.ID, Username: a.Username, AccountType: a.AccountType, CreatedAt: a.CreatedAt.Format(time.RFC3339)}), nil
}

func (u *Usecase) toolDeleteTwitchAccount(ctx context.Context, args string) (string, error) {
	var p struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	if err := u.sett.DeleteTwitchAccount(ctx, p.ID); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]any{"ok": true, "deleted_id": p.ID}), nil
}
