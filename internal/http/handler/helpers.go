package handler

import (
	"encoding/json"
	"time"

	"github.com/go-faster/jx"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

func chatHistoryBadgeTags(tags []string) []gen.ChatHistoryEntryBadgeTagsItem {
	out := make([]gen.ChatHistoryEntryBadgeTagsItem, 0, len(tags))

	for _, t := range tags {
		switch t {
		case "moderator":
			out = append(out, gen.ChatHistoryEntryBadgeTagsItemModerator)
		case "vip":
			out = append(out, gen.ChatHistoryEntryBadgeTagsItemVip)
		case "bot":
			out = append(out, gen.ChatHistoryEntryBadgeTagsItemBot)
		case "other":
			out = append(out, gen.ChatHistoryEntryBadgeTagsItemOther)
		}
	}

	return out
}

func chatHistoryEntityToGen(m entity.ChatHistoryMessage) gen.ChatHistoryEntry {
	src := gen.ChatHistoryEntrySourceIrc
	if m.MsgType == "sent" {
		src = gen.ChatHistoryEntrySourceSent
	}

	var chatter gen.OptNilInt64
	if m.ChatterTwitchUserID != nil {
		chatter.SetTo(*m.ChatterTwitchUserID)
	} else {
		chatter.SetToNull()
	}

	return gen.ChatHistoryEntry{
		ID:            m.ID,
		Channel:       m.Channel,
		User:          m.Username,
		ChatterUserID: chatter,
		ChatterMarked: m.ChatterMarked,
		ChatterIsSus:  m.ChatterIsSus,
		FirstMessage:  m.FirstMessage,
		Message:       m.Message,
		KeywordMatch:  m.KeywordMatch,
		Source:        src,
		CreatedAt:     m.CreatedAt,
		BadgeTags:     chatHistoryBadgeTags(m.BadgeTags),
	}
}

func ruleEntityToGen(r entity.Rule) gen.Rule {
	return gen.Rule{
		ID:             r.ID,
		Enabled:        r.Enabled,
		EventType:      gen.RuleEventType(r.EventType),
		EventSettings:  anyMapToRuleEventSettings(r.EventSettings),
		Middlewares:    middlewaresEntityToGen(r.Middlewares),
		ActionType:     gen.RuleActionType(r.ActionType),
		ActionSettings: anyMapToActionSettings(r.ActionSettings),
		UseSharedPool:  r.UseSharedPool,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func anyMapToRuleEventSettings(m map[string]any) gen.RuleEventSettings {
	out := make(map[string]jx.Raw)

	for k, v := range m {
		raw, err := json.Marshal(v)
		if err != nil {
			continue
		}

		out[k] = jx.Raw(raw)
	}

	return out
}

func anyMapToActionSettings(m map[string]any) gen.RuleActionSettings {
	return gen.RuleActionSettings(anyMapToRuleEventSettings(m))
}

func middlewaresEntityToGen(in []entity.RuleMiddleware) []gen.RuleMiddleware {
	out := make([]gen.RuleMiddleware, 0, len(in))

	for _, m := range in {
		out = append(out, gen.RuleMiddleware{
			Type:     m.Type,
			Settings: anyMapToMiddlewareSettings(m.Settings),
		})
	}

	return out
}

func anyMapToMiddlewareSettings(m map[string]any) gen.RuleMiddlewareSettings {
	return gen.RuleMiddlewareSettings(anyMapToRuleEventSettings(m))
}

func createRuleReqToEntity(req *gen.CreateRuleRequest) entity.Rule {
	return entity.Rule{
		Enabled:        req.Enabled.Or(true),
		EventType:      string(req.EventType),
		EventSettings:  rawSettingsToMap(req.EventSettings),
		Middlewares:    middlewaresGenToEntity(req.Middlewares),
		ActionType:     string(req.ActionType),
		ActionSettings: rawSettingsToMap(req.ActionSettings),
		UseSharedPool:  req.UseSharedPool.Or(true),
	}
}

func middlewaresGenToEntity(in []gen.RuleMiddleware) []entity.RuleMiddleware {
	out := make([]entity.RuleMiddleware, 0, len(in))

	for _, m := range in {
		out = append(out, entity.RuleMiddleware{
			Type:     m.Type,
			Settings: rawSettingsToMap(m.Settings),
		})
	}

	return out
}

func updateRulePostReqToEntity(req *gen.UpdateRulePostRequest) entity.Rule {
	return entity.Rule{
		Enabled:        req.Enabled,
		EventType:      string(req.EventType),
		EventSettings:  rawSettingsToMap(req.EventSettings),
		Middlewares:    middlewaresGenToEntity(req.Middlewares),
		ActionType:     string(req.ActionType),
		ActionSettings: rawSettingsToMap(req.ActionSettings),
		UseSharedPool:  req.UseSharedPool,
	}
}

func notificationEntityToGen(e entity.NotificationEntry) gen.NotificationEntry {
	settings := gen.NotificationEntrySettings{}

	for k, v := range e.Settings {
		raw, err := json.Marshal(v)
		if err != nil {
			continue
		}

		settings[k] = jx.Raw(raw)
	}

	prov := gen.NotificationEntryProviderTelegram
	if e.Provider == "webhook" {
		prov = gen.NotificationEntryProviderWebhook
	}

	return gen.NotificationEntry{
		ID:        e.ID,
		Provider:  prov,
		Settings:  settings,
		Enabled:   e.Enabled,
		CreatedAt: e.CreatedAt,
	}
}

func rawSettingsToMap(s map[string]jx.Raw) map[string]any {
	out := map[string]any{}

	for k, v := range s {
		var val any
		if err := json.Unmarshal([]byte(v), &val); err == nil {
			out[k] = val
		}
	}
	return out
}

func streamEntityToGen(s entity.Stream) gen.RecordedStream {
	out := gen.RecordedStream{
		ID:            s.ID,
		ChannelID:     s.ChannelTwitchUserID,
		ChannelLogin:  s.ChannelLogin,
		HelixStreamID: s.HelixStreamID,
		StartedAt:     s.StartedAt,
		CreatedAt:     s.CreatedAt,
	}

	if s.EndedAt != nil {
		out.SetEndedAt(gen.NewOptNilDateTime(*s.EndedAt))
	} else {
		var z gen.OptNilDateTime
		z.SetToNull()
		out.SetEndedAt(z)
	}

	if s.Title != "" {
		out.SetTitle(gen.NewOptNilString(s.Title))
	} else {
		var t gen.OptNilString
		t.SetToNull()
		out.SetTitle(t)
	}

	if s.GameName != "" {
		out.SetGameName(gen.NewOptNilString(s.GameName))
	} else {
		var g gen.OptNilString
		g.SetToNull()
		out.SetGameName(g)
	}

	return out
}

func optNilStringFromPtr(s *string) gen.OptNilString {
	if s == nil || *s == "" {
		var z gen.OptNilString
		z.SetToNull()
		return z
	}
	return gen.NewOptNilString(*s)
}

func entityTwitchUserToGen(u entity.TwitchUser) gen.TwitchUser {
	out := gen.TwitchUser{
		ID:                      u.ID,
		Username:                u.Username,
		Monitored:               u.Monitored,
		Marked:                  u.Marked,
		IsSus:                   u.IsSus,
		SusType:                 optNilStringFromPtr(u.SusType),
		SusDescription:          optNilStringFromPtr(u.SusDescription),
		SusAutoSuppressed:       u.SusAutoSuppressed,
		IrcOnlyWhenLive:         u.IrcOnlyWhenLive,
		NotifyOffStreamMessages: u.NotifyOffStreamMessages,
		NotifyStreamStart:       u.NotifyStreamStart,
	}

	return out
}

func directoryEntryToGen(e entity.TwitchDirectoryEntry) gen.TwitchUser {
	out := entityTwitchUserToGen(e.User)

	if e.ProfileImageURL != nil && *e.ProfileImageURL != "" {
		out.SetProfileImageURL(gen.NewOptNilString(*e.ProfileImageURL))
	}

	if e.ChannelLive == nil {
		return out
	}

	l := e.ChannelLive
	prof := ""
	if e.ProfileImageURL != nil {
		prof = *e.ProfileImageURL
	}

	cl := gen.ChannelLive{
		BroadcasterID:    e.User.ID,
		BroadcasterLogin: e.User.Username,
		DisplayName:      e.User.Username,
		ProfileImageURL:  prof,
		IsLive:           l.IsLive,
	}

	if l.Title != nil && *l.Title != "" {
		cl.SetTitle(gen.NewOptNilString(*l.Title))
	} else {
		var t gen.OptNilString
		t.SetToNull()
		cl.SetTitle(t)
	}

	if l.GameName != nil && *l.GameName != "" {
		cl.SetGameName(gen.NewOptNilString(*l.GameName))
	} else {
		var g gen.OptNilString
		g.SetToNull()
		cl.SetGameName(g)
	}

	if l.IsLive && l.ViewerCount != nil {
		cl.SetViewerCount(gen.NewOptNilInt64(*l.ViewerCount))
	} else {
		var v gen.OptNilInt64
		v.SetToNull()
		cl.SetViewerCount(v)
	}

	if l.ChannelChatterCount != nil {
		cl.SetChannelChatterCount(gen.NewOptNilInt64(*l.ChannelChatterCount))
	} else {
		var cc gen.OptNilInt64
		cc.SetToNull()
		cl.SetChannelChatterCount(cc)
	}

	if l.StartedAt != nil {
		cl.SetStartedAt(gen.NewOptNilDateTime(*l.StartedAt))
	} else {
		var s gen.OptNilDateTime
		s.SetToNull()
		cl.SetStartedAt(s)
	}

	out.SetChannelLive(gen.NewOptNilChannelLive(cl))

	return out
}

func twitchAccountToAPI(a entity.TwitchAccount) gen.TwitchAccount {
	at := gen.TwitchAccountAccountTypeMain
	if a.AccountType == "bot" {
		at = gen.TwitchAccountAccountTypeBot
	}

	return gen.TwitchAccount{
		ID:          a.ID,
		Username:    a.Username,
		AccountType: at,
		CreatedAt:   a.CreatedAt,
	}
}

func entityActivityToGen(e entity.UserActivityEvent, profileUsername string) gen.UserActivityEvent {
	var et gen.UserActivityEventEventType

	switch e.EventType {
	case entity.UserActivityChatOnline:
		et = gen.UserActivityEventEventTypeChatOnline
	case entity.UserActivityChatOffline:
		et = gen.UserActivityEventEventTypeChatOffline
	case entity.UserActivityMessage:
		et = gen.UserActivityEventEventTypeMessage
	default:
		et = gen.UserActivityEventEventTypeMessage
	}

	uname := profileUsername
	if e.ChatterLogin != "" {
		uname = e.ChatterLogin
	}

	ge := gen.UserActivityEvent{
		ID:        e.ID,
		Username:  uname,
		EventType: et,
		CreatedAt: e.CreatedAt,
	}

	if e.ChannelLogin != "" {
		ge.SetChannel(gen.NewOptNilString(e.ChannelLogin))
	} else {
		var c gen.OptNilString
		c.SetToNull()
		ge.SetChannel(c)
	}

	if len(e.Details) > 0 {
		raw, err := json.Marshal(e.Details)
		if err == nil {
			var rm map[string]json.RawMessage
			if err := json.Unmarshal(raw, &rm); err == nil {
				dm := make(gen.UserActivityEventDetails)
				for k, v := range rm {
					dm[k] = jx.Raw(v)
				}

				ge.SetDetails(gen.NewOptNilUserActivityEventDetails(dm))
			}
		}
	} else {
		var d gen.OptNilUserActivityEventDetails
		d.SetToNull()
		ge.SetDetails(d)
	}

	return ge
}

func ircMonitorEntityToGen(s entity.IrcMonitorSettings) *gen.IrcMonitorSettings {
	g := &gen.IrcMonitorSettings{}
	if s.OauthTwitchAccountID == nil {
		g.OAuthTwitchAccountID.SetToNull()
	} else {
		g.OAuthTwitchAccountID.SetTo(*s.OauthTwitchAccountID)
	}

	g.EnrichmentCooldownHours = int(s.EnrichmentCooldown / time.Hour)

	return g
}

func ircMonitorGenToEntity(req *gen.IrcMonitorSettings) entity.IrcMonitorSettings {
	if req == nil {
		return entity.IrcMonitorSettings{}
	}

	n := req.GetOAuthTwitchAccountID()
	cooldown := time.Duration(req.EnrichmentCooldownHours) * time.Hour
	if cooldown <= 0 {
		cooldown = 24 * time.Hour
	}

	if n.IsNull() {
		return entity.IrcMonitorSettings{
			OauthTwitchAccountID: nil,
			EnrichmentCooldown:   cooldown,
		}
	}

	v, ok := n.Get()
	if !ok {
		return entity.IrcMonitorSettings{
			OauthTwitchAccountID: nil,
			EnrichmentCooldown:   cooldown,
		}
	}

	return entity.IrcMonitorSettings{
		OauthTwitchAccountID: &v,
		EnrichmentCooldown:   cooldown,
	}
}

func suspicionEntityToGen(s entity.SuspicionSettings) *gen.SuspicionSettings {
	return &gen.SuspicionSettings{
		AutoCheckAccountAge: s.AutoCheckAccountAge,
		AccountAgeSusDays:   s.AccountAgeSusDays,
		AutoCheckBlacklist:  s.AutoCheckBlacklist,
		AutoCheckLowFollows: s.AutoCheckLowFollows,
		LowFollowsThreshold: s.LowFollowsThreshold,
		MaxGqlFollowPages:   s.MaxGQLFollowPages,
	}
}

func suspicionGenToEntity(s *gen.SuspicionSettings) entity.SuspicionSettings {
	if s == nil {
		return entity.SuspicionSettings{}
	}

	return entity.SuspicionSettings{
		AutoCheckAccountAge: s.AutoCheckAccountAge,
		AccountAgeSusDays:   s.AccountAgeSusDays,
		AutoCheckBlacklist:  s.AutoCheckBlacklist,
		AutoCheckLowFollows: s.AutoCheckLowFollows,
		LowFollowsThreshold: s.LowFollowsThreshold,
		MaxGQLFollowPages:   s.MaxGqlFollowPages,
	}
}
