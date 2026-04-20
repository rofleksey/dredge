package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/usecase/rules"
)

// ExecTool runs one tool and returns JSON text for the model (or an error string payload).
func (u *Usecase) ExecTool(ctx context.Context, name, argumentsJSON string) (string, error) {
	switch name {
	case ToolListTwitchMessages:
		return u.toolListTwitchMessages(ctx, argumentsJSON)
	case ToolGetTwitchUserProfile:
		return u.toolGetTwitchUserProfile(ctx, argumentsJSON)
	case ToolListTwitchUserActivity:
		return u.toolListTwitchUserActivity(ctx, argumentsJSON)
	case ToolGetTwitchUserActivityTimeline:
		return u.toolGetTwitchUserActivityTimeline(ctx, argumentsJSON)
	case ToolListTwitchDirectoryUsers:
		return u.toolListTwitchDirectoryUsers(ctx, argumentsJSON)
	case ToolListChatHistory:
		return u.toolListChatHistory(ctx, argumentsJSON)
	case ToolListRules:
		return u.toolListRules(ctx)
	case ToolRuleTemplateVariables:
		return u.toolRuleTemplateVariables()
	case ToolTestRuleRegex:
		return u.toolTestRuleRegex(argumentsJSON)
	case ToolListTwitchUsers:
		return u.toolListTwitchUsers(ctx, argumentsJSON)
	case ToolListNotifications:
		return u.toolListNotifications(ctx)
	case ToolListChannelBlacklist:
		return u.toolListChannelBlacklist(ctx)
	case ToolGetSuspicionSettings:
		return u.toolGetSuspicionSettings(ctx)
	case ToolGetIrcMonitorSettings:
		return u.toolGetIrcMonitorSettings(ctx)
	case ToolListTwitchAccounts:
		return u.toolListTwitchAccounts(ctx)
	case ToolCreateRule:
		return u.toolCreateRule(ctx, argumentsJSON)
	case ToolUpdateRule:
		return u.toolUpdateRule(ctx, argumentsJSON)
	case ToolDeleteRule:
		return u.toolDeleteRule(ctx, argumentsJSON)
	case ToolSendTwitchMessage:
		return u.toolSendTwitchMessage(ctx, argumentsJSON)
	case ToolCountTwitchMessages:
		return u.toolCountTwitchMessages(ctx, argumentsJSON)
	case ToolCountTwitchDirectoryUsers:
		return u.toolCountTwitchDirectoryUsers(ctx, argumentsJSON)
	case ToolGetChannelLive:
		return u.toolGetChannelLive(ctx, argumentsJSON)
	case ToolListChannelChatters:
		return u.toolListChannelChatters(ctx, argumentsJSON)
	case ToolGetIrcMonitorStatus:
		return u.toolGetIrcMonitorStatus(ctx)
	case ToolGetWatchUiHints:
		return u.toolGetWatchUiHints()
	case ToolListMonitoredStreams:
		return u.toolListMonitoredStreams(ctx, argumentsJSON)
	case ToolGetMonitoredStream:
		return u.toolGetMonitoredStream(ctx, argumentsJSON)
	case ToolListStreamMessages:
		return u.toolListStreamMessages(ctx, argumentsJSON)
	case ToolListStreamActivity:
		return u.toolListStreamActivity(ctx, argumentsJSON)
	case ToolGetStreamLeaderboard:
		return u.toolGetStreamLeaderboard(ctx, argumentsJSON)
	case ToolGetTwitchUser:
		return u.toolGetTwitchUser(ctx, argumentsJSON)
	case ToolCountRules:
		return u.toolCountRules(ctx)
	case ToolCreateNotification:
		return u.toolCreateNotification(ctx, argumentsJSON)
	case ToolUpdateNotification:
		return u.toolUpdateNotification(ctx, argumentsJSON)
	case ToolDeleteNotification:
		return u.toolDeleteNotification(ctx, argumentsJSON)
	case ToolSetChannelBlacklist:
		return u.toolSetChannelBlacklist(ctx, argumentsJSON)
	case ToolUpdateSuspicionSettings:
		return u.toolUpdateSuspicionSettings(ctx, argumentsJSON)
	case ToolUpdateIrcMonitorSettings:
		return u.toolUpdateIrcMonitorSettings(ctx, argumentsJSON)
	case ToolCreateTwitchUser:
		return u.toolCreateTwitchUser(ctx, argumentsJSON)
	case ToolPatchTwitchUser:
		return u.toolPatchTwitchUser(ctx, argumentsJSON)
	case ToolCreateTwitchAccount:
		return u.toolCreateTwitchAccount(ctx, argumentsJSON)
	case ToolPatchTwitchAccount:
		return u.toolPatchTwitchAccount(ctx, argumentsJSON)
	case ToolDeleteTwitchAccount:
		return u.toolDeleteTwitchAccount(ctx, argumentsJSON)
	default:
		b, _ := json.Marshal(map[string]string{"error": "unknown tool: " + name})
		return string(b), fmt.Errorf("unknown tool %q", name)
	}
}

func mustJSON(v any) string {
	normalized := normalizeToolOutputJSON(v)
	b, err := json.Marshal(normalized)
	if err != nil {
		return `{"error":"marshal failed"}`
	}
	return string(b)
}

var (
	snakeKeyBoundary1 = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	snakeKeyBoundary2 = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func normalizeToolOutputJSON(v any) any {
	// Marshal/unmarshal first so custom MarshalJSON logic (time.Time, etc.) is preserved.
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}

	var decoded any
	if err := json.Unmarshal(b, &decoded); err != nil {
		return v
	}

	return normalizeJSONKeys(decoded)
}

func normalizeJSONKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[toSnakeCaseKey(k)] = normalizeJSONKeys(vv)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = normalizeJSONKeys(t[i])
		}
		return out
	default:
		return v
	}
}

func toSnakeCaseKey(s string) string {
	if s == "" {
		return s
	}
	if strings.Contains(s, "_") {
		return strings.ToLower(s)
	}
	k := snakeKeyBoundary1.ReplaceAllString(s, `${1}_${2}`)
	k = snakeKeyBoundary2.ReplaceAllString(k, `${1}_${2}`)
	return strings.ToLower(k)
}

func (u *Usecase) toolListTwitchMessages(ctx context.Context, args string) (string, error) {
	var p struct {
		Username        string `json:"username"`
		Channel         string `json:"channel"`
		Text            string `json:"text"`
		Limit           int    `json:"limit"`
		ChatterUserID   *int64 `json:"chatter_user_id"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	f := entity.ChatMessageListFilter{
		Username:      p.Username,
		Text:          p.Text,
		Channel:       p.Channel,
		Limit:         limit,
		ChatterUserID: p.ChatterUserID,
	}
	msgs, err := u.tw.ListChatMessages(ctx, f)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(msgs), nil
}

func (u *Usecase) toolGetTwitchUserProfile(ctx context.Context, args string) (string, error) {
	var p struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	uu, msgCount, pres, ac, img, monF, gqlF, bl, err := u.tw.GetTwitchUserProfile(ctx, p.ID)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	out := map[string]any{
		"user":                      uu,
		"message_count":             msgCount,
		"presence_seconds_this_week": pres,
		"account_created_at":        ac,
		"profile_image_url":         img,
		"followed_monitored_channels": monF,
		"followed_channels_gql":     gqlF,
		"channel_blacklist":         bl,
	}
	return mustJSON(out), nil
}

func (u *Usecase) toolListTwitchUserActivity(ctx context.Context, args string) (string, error) {
	var p struct {
		ID    int64 `json:"id"`
		Limit int   `json:"limit"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	lim := p.Limit
	if lim <= 0 {
		lim = 50
	}
	if lim > 200 {
		lim = 200
	}
	f := entity.UserActivityListFilter{ChatterUserID: p.ID, Limit: lim}
	ev, err := u.tw.ListUserActivity(ctx, f)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(ev), nil
}

func (u *Usecase) toolGetTwitchUserActivityTimeline(ctx context.Context, args string) (string, error) {
	var p struct {
		ID   int64  `json:"id"`
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	to := time.Now()
	if strings.TrimSpace(p.To) != "" {
		t, err := time.Parse(time.RFC3339, p.To)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad to: " + err.Error()}), err
		}
		to = t
	}
	from := to.Add(-7 * 24 * time.Hour)
	if strings.TrimSpace(p.From) != "" {
		t, err := time.Parse(time.RFC3339, p.From)
		if err != nil {
			return mustJSON(map[string]string{"error": "bad from: " + err.Error()}), err
		}
		from = t
	}
	segs, err := u.tw.GetUserActivityTimeline(ctx, p.ID, from, to)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(segs), nil
}

func (u *Usecase) toolListTwitchDirectoryUsers(ctx context.Context, args string) (string, error) {
	var p struct {
		Username       string `json:"username"`
		Limit          int    `json:"limit"`
		MonitoredOnly  bool   `json:"monitored_only"`
	}
	_ = json.Unmarshal([]byte(args), &p)
	lim := p.Limit
	if lim <= 0 {
		lim = 50
	}
	if lim > 200 {
		lim = 200
	}
	f := entity.TwitchUserBrowseFilter{Username: p.Username, Limit: lim, MonitoredOnly: p.MonitoredOnly}
	rows, err := u.tw.ListTwitchUsersBrowse(ctx, f)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(rows), nil
}

func (u *Usecase) toolListChatHistory(ctx context.Context, args string) (string, error) {
	var p struct {
		Channel string `json:"channel"`
		Limit   int    `json:"limit"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	lim := p.Limit
	if lim <= 0 {
		lim = 50
	}
	if lim > 200 {
		lim = 200
	}
	msgs, err := u.tw.ListChatHistory(ctx, p.Channel, lim)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(msgs), nil
}

func (u *Usecase) toolListRules(ctx context.Context) (string, error) {
	list, err := u.rules.ListRules(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func (u *Usecase) toolRuleTemplateVariables() (string, error) {
	v := rules.RuleTemplateVariables()
	return mustJSON(v), nil
}

func (u *Usecase) toolTestRuleRegex(args string) (string, error) {
	var p struct {
		Pattern          string `json:"pattern"`
		Sample           string `json:"sample"`
		CaseInsensitive  bool   `json:"case_insensitive"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	pat := p.Pattern
	if p.CaseInsensitive {
		pat = "(?i)" + pat
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return mustJSON(map[string]any{"matches": false, "compile_error": err.Error()}), nil
	}
	return mustJSON(map[string]any{"matches": re.MatchString(p.Sample)}), nil
}

func (u *Usecase) toolListTwitchUsers(ctx context.Context, args string) (string, error) {
	var p struct {
		MonitoredOnly bool `json:"monitored_only"`
	}
	_ = json.Unmarshal([]byte(args), &p)
	list, err := u.sett.ListTwitchUsers(ctx, p.MonitoredOnly)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func (u *Usecase) toolListNotifications(ctx context.Context) (string, error) {
	list, err := u.sett.ListNotifications(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func (u *Usecase) toolListChannelBlacklist(ctx context.Context) (string, error) {
	list, err := u.sett.ListChannelBlacklist(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(list), nil
}

func (u *Usecase) toolGetSuspicionSettings(ctx context.Context) (string, error) {
	s, err := u.sett.GetSuspicionSettings(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(s), nil
}

func (u *Usecase) toolGetIrcMonitorSettings(ctx context.Context) (string, error) {
	s, err := u.sett.GetIrcMonitorSettings(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(s), nil
}

func (u *Usecase) toolListTwitchAccounts(ctx context.Context) (string, error) {
	list, err := u.sett.ListTwitchAccounts(ctx)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	// Redact refresh tokens
	type acc struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		AccountType string `json:"account_type"`
		CreatedAt   string `json:"created_at"`
	}
	out := make([]acc, 0, len(list))
	for _, a := range list {
		out = append(out, acc{ID: a.ID, Username: a.Username, AccountType: a.AccountType, CreatedAt: a.CreatedAt.Format(time.RFC3339)})
	}
	return mustJSON(out), nil
}

func (u *Usecase) toolCreateRule(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	r := entity.Rule{
		Name:           stringField(raw, "name"),
		EventType:      stringField(raw, "event_type"),
		EventSettings:  mapField(raw, "event_settings"),
		ActionType:     stringField(raw, "action_type"),
		ActionSettings: mapField(raw, "action_settings"),
	}
	if v, ok := raw["enabled"].(bool); ok {
		r.Enabled = v
	} else {
		r.Enabled = true
	}
	if v, ok := raw["use_shared_pool"].(bool); ok {
		r.UseSharedPool = v
	} else {
		r.UseSharedPool = true
	}
	r.Middlewares = middlewaresFromRaw(raw["middlewares"])
	out, err := u.rules.CreateRule(ctx, r)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(out), nil
}

func (u *Usecase) toolUpdateRule(ctx context.Context, args string) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(args), &raw); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	id, err := int64Field(raw, "id")
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	existing, err := u.findRuleByID(ctx, id)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	merged := mergeRulePatch(existing, raw)
	out, err := u.rules.UpdateRule(ctx, id, merged)
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(out), nil
}

func (u *Usecase) findRuleByID(ctx context.Context, id int64) (entity.Rule, error) {
	list, err := u.rules.ListRules(ctx)
	if err != nil {
		return entity.Rule{}, err
	}
	for _, r := range list {
		if r.ID == id {
			return r, nil
		}
	}
	return entity.Rule{}, entity.ErrRuleNotFound
}

func cloneStringMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	return maps.Clone(m)
}

func mergeStringMaps(base, patch map[string]any) map[string]any {
	out := map[string]any{}
	if base != nil {
		out = maps.Clone(base)
	}
	for k, v := range patch {
		out[k] = v
	}
	return out
}

func copyRuleForMerge(r entity.Rule) entity.Rule {
	out := r
	out.EventSettings = cloneStringMap(r.EventSettings)
	out.ActionSettings = cloneStringMap(r.ActionSettings)
	if len(r.Middlewares) > 0 {
		out.Middlewares = make([]entity.RuleMiddleware, len(r.Middlewares))
		for i := range r.Middlewares {
			out.Middlewares[i].Type = r.Middlewares[i].Type
			out.Middlewares[i].Settings = cloneStringMap(r.Middlewares[i].Settings)
		}
	}
	return out
}

// mergeRulePatch overlays JSON keys present in raw onto a copy of existing (patch semantics).
func mergeRulePatch(existing entity.Rule, raw map[string]any) entity.Rule {
	r := copyRuleForMerge(existing)
	if _, ok := raw["name"]; ok {
		r.Name = stringField(raw, "name")
	}
	if _, ok := raw["enabled"]; ok {
		if v, ok := raw["enabled"].(bool); ok {
			r.Enabled = v
		}
	}
	if _, ok := raw["use_shared_pool"]; ok {
		if v, ok := raw["use_shared_pool"].(bool); ok {
			r.UseSharedPool = v
		}
	}
	if _, ok := raw["event_type"]; ok {
		r.EventType = stringField(raw, "event_type")
	}
	if _, ok := raw["action_type"]; ok {
		r.ActionType = stringField(raw, "action_type")
	}
	if _, ok := raw["event_settings"]; ok {
		patch := mapField(raw, "event_settings")
		base := r.EventSettings
		if base == nil {
			base = map[string]any{}
		}
		r.EventSettings = mergeStringMaps(base, patch)
	}
	if _, ok := raw["action_settings"]; ok {
		patch := mapField(raw, "action_settings")
		base := r.ActionSettings
		if base == nil {
			base = map[string]any{}
		}
		r.ActionSettings = mergeStringMaps(base, patch)
	}
	if _, ok := raw["middlewares"]; ok {
		r.Middlewares = middlewaresFromRaw(raw["middlewares"])
	}
	return r
}

func (u *Usecase) toolDeleteRule(ctx context.Context, args string) (string, error) {
	var p struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	if err := u.rules.DeleteRule(ctx, p.ID); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]any{"ok": true, "deleted_id": p.ID}), nil
}

func (u *Usecase) toolSendTwitchMessage(ctx context.Context, args string) (string, error) {
	var p struct {
		AccountID int64  `json:"account_id"`
		Channel   string `json:"channel"`
		Message   string `json:"message"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	if err := u.tw.SendMessage(ctx, p.AccountID, p.Channel, p.Message); err != nil {
		return mustJSON(map[string]string{"error": err.Error()}), err
	}
	return mustJSON(map[string]any{"ok": true}), nil
}

func stringField(m map[string]any, k string) string {
	v, ok := m[k]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		return fmt.Sprint(t)
	}
}

func int64Field(m map[string]any, k string) (int64, error) {
	v, ok := m[k]
	if !ok || v == nil {
		return 0, fmt.Errorf("missing %s", k)
	}
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
		return 0, fmt.Errorf("invalid %s", k)
	}
}

func boolField(m map[string]any, k string, def bool) bool {
	v, ok := m[k]
	if !ok {
		return def
	}
	b, ok := v.(bool)
	if !ok {
		return def
	}
	return b
}

func mapField(m map[string]any, k string) map[string]any {
	v, ok := m[k]
	if !ok || v == nil {
		return map[string]any{}
	}
	mm, ok := v.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return mm
}

func middlewaresFromRaw(v any) []entity.RuleMiddleware {
	arr, ok := v.([]any)
	if !ok || len(arr) == 0 {
		return nil
	}
	out := make([]entity.RuleMiddleware, 0, len(arr))
	for _, x := range arr {
		mm, ok := x.(map[string]any)
		if !ok {
			continue
		}
		t := stringField(mm, "type")
		st := mapField(mm, "settings")
		out = append(out, entity.RuleMiddleware{Type: t, Settings: st})
	}
	return out
}
