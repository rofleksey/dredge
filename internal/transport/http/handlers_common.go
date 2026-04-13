package httptransport

import (
	"encoding/json"

	"github.com/go-faster/jx"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
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
		Message:       m.Message,
		KeywordMatch:  m.KeywordMatch,
		Source:        src,
		CreatedAt:     m.CreatedAt,
		BadgeTags:     chatHistoryBadgeTags(m.BadgeTags),
	}
}

func ruleEntityToGen(r entity.Rule) gen.Rule {
	return gen.Rule{
		ID:               r.ID,
		Regex:            r.Regex,
		IncludedUsers:    r.IncludedUsers,
		DeniedUsers:      r.DeniedUsers,
		IncludedChannels: r.IncludedChannels,
		DeniedChannels:   r.DeniedChannels,
	}
}

func createRuleReqToEntity(req *gen.CreateRuleRequest) entity.Rule {
	r := entity.Rule{Regex: req.Regex}
	if req.IncludedUsers.IsSet() {
		r.IncludedUsers = req.IncludedUsers.Value
	} else {
		r.IncludedUsers = "*"
	}

	if req.DeniedUsers.IsSet() {
		r.DeniedUsers = req.DeniedUsers.Value
	}

	if req.IncludedChannels.IsSet() {
		r.IncludedChannels = req.IncludedChannels.Value
	} else {
		r.IncludedChannels = "*"
	}

	if req.DeniedChannels.IsSet() {
		r.DeniedChannels = req.DeniedChannels.Value
	}

	return r
}

func updateRulePostReqToEntity(req *gen.UpdateRulePostRequest) entity.Rule {
	return entity.Rule{
		Regex:            req.GetRegex(),
		IncludedUsers:    req.GetIncludedUsers(),
		DeniedUsers:      req.GetDeniedUsers(),
		IncludedChannels: req.GetIncludedChannels(),
		DeniedChannels:   req.GetDeniedChannels(),
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
