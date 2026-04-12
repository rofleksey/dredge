package httptransport

import (
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func TestChatHistoryBadgeTags(t *testing.T) {
	t.Parallel()

	out := chatHistoryBadgeTags([]string{"moderator", "vip", "bot", "other", "unknown"})
	require.Len(t, out, 4)
	assert.Equal(t, gen.ChatHistoryEntryBadgeTagsItemModerator, out[0])
	assert.Equal(t, gen.ChatHistoryEntryBadgeTagsItemVip, out[1])
	assert.Equal(t, gen.ChatHistoryEntryBadgeTagsItemBot, out[2])
	assert.Equal(t, gen.ChatHistoryEntryBadgeTagsItemOther, out[3])
}

func TestChatHistoryEntityToGen(t *testing.T) {
	t.Parallel()

	m := entity.ChatHistoryMessage{
		ID:                  1,
		Channel:             "c",
		Username:            "u",
		ChatterTwitchUserID: ptrInt64(9),
		Message:             "hi",
		MsgType:             "sent",
		CreatedAt:           time.Unix(1, 0).UTC(),
		BadgeTags:           []string{"moderator"},
	}
	g := chatHistoryEntityToGen(m)
	assert.Equal(t, gen.ChatHistoryEntrySourceSent, g.Source)
	assert.True(t, g.ChatterUserID.IsSet())
	v, _ := g.ChatterUserID.Get()
	assert.Equal(t, int64(9), v)

	m2 := entity.ChatHistoryMessage{
		ID:        2,
		Channel:   "c",
		Username:  "u",
		Message:   "irc",
		MsgType:   "irc",
		CreatedAt: time.Unix(2, 0).UTC(),
	}
	g2 := chatHistoryEntityToGen(m2)
	assert.Equal(t, gen.ChatHistoryEntrySourceIrc, g2.Source)
}

func TestCreateRuleReqToEntity_defaults(t *testing.T) {
	t.Parallel()

	req := &gen.CreateRuleRequest{Regex: `foo`}
	ent := createRuleReqToEntity(req)
	assert.Equal(t, `foo`, ent.Regex)
	assert.Equal(t, "*", ent.IncludedUsers)
	assert.Equal(t, "*", ent.IncludedChannels)
	assert.Empty(t, ent.DeniedUsers)
	assert.Empty(t, ent.DeniedChannels)

	req2 := &gen.CreateRuleRequest{
		Regex:            `bar`,
		IncludedUsers:    gen.NewOptString("a"),
		IncludedChannels: gen.NewOptString("c"),
		DeniedUsers:      gen.NewOptString("d"),
		DeniedChannels:   gen.NewOptString("x"),
	}
	ent2 := createRuleReqToEntity(req2)
	assert.Equal(t, "a", ent2.IncludedUsers)
	assert.Equal(t, "c", ent2.IncludedChannels)
	assert.Equal(t, "d", ent2.DeniedUsers)
	assert.Equal(t, "x", ent2.DeniedChannels)
}

func TestRuleEntityToGen(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		ID:               3,
		Regex:            "x",
		IncludedUsers:    "u",
		DeniedUsers:      "",
		IncludedChannels: "c",
		DeniedChannels:   "",
	}
	g := ruleEntityToGen(r)
	assert.Equal(t, int64(3), g.ID)
	assert.Equal(t, "x", g.Regex)
}

func TestNotificationEntityToGen(t *testing.T) {
	t.Parallel()

	e := entity.NotificationEntry{
		ID:       1,
		Provider: "telegram",
		Settings: map[string]any{"k": "v"},
		Enabled:  true,
	}
	g := notificationEntityToGen(e)
	assert.Equal(t, gen.NotificationEntryProviderTelegram, g.Provider)

	e.Provider = "webhook"
	g2 := notificationEntityToGen(e)
	assert.Equal(t, gen.NotificationEntryProviderWebhook, g2.Provider)
}

func TestRawSettingsToMap(t *testing.T) {
	t.Parallel()

	raw := map[string]jx.Raw{
		"bad": jx.Raw(`invalid`),
		"ok":  jx.Raw(`"x"`),
	}
	m := rawSettingsToMap(raw)
	require.Len(t, m, 1)
	assert.Equal(t, "x", m["ok"])
}

func TestTwitchAccountToAPI(t *testing.T) {
	t.Parallel()

	a := entity.TwitchAccount{ID: 1, Username: "u", AccountType: "main"}
	g := twitchAccountToAPI(a)
	assert.Equal(t, gen.TwitchAccountAccountTypeMain, g.AccountType)

	a.AccountType = "bot"
	g2 := twitchAccountToAPI(a)
	assert.Equal(t, gen.TwitchAccountAccountTypeBot, g2.AccountType)
}

func TestEntityActivityToGen(t *testing.T) {
	t.Parallel()

	ev := entity.UserActivityEvent{
		ID:           1,
		EventType:    entity.UserActivityChatOnline,
		CreatedAt:    time.Unix(0, 0).UTC(),
		ChannelLogin: "chan",
	}
	g := entityActivityToGen(ev, "profile")
	assert.Equal(t, gen.UserActivityEventEventTypeChatOnline, g.EventType)
	assert.Equal(t, "profile", g.Username)

	ev2 := entity.UserActivityEvent{
		ID:        2,
		EventType: entity.UserActivityChatOffline,
		CreatedAt: time.Unix(0, 0).UTC(),
		Details:   map[string]any{"a": 1},
	}
	g2 := entityActivityToGen(ev2, "u")
	assert.Equal(t, gen.UserActivityEventEventTypeChatOffline, g2.EventType)

	ev3 := entity.UserActivityEvent{
		ID:        3,
		EventType: "other",
		Details:   map[string]any{},
	}
	g3 := entityActivityToGen(ev3, "u")
	assert.Equal(t, gen.UserActivityEventEventTypeMessage, g3.EventType)

	ev4 := entity.UserActivityEvent{
		ID:           4,
		EventType:    entity.UserActivityChatOnline,
		CreatedAt:    time.Unix(0, 0).UTC(),
		ChannelLogin: "chan",
		ChatterLogin: "chatter",
	}
	g4 := entityActivityToGen(ev4, "profile")
	assert.Equal(t, "chatter", g4.Username)
}

func TestUpdateRulePostReqToEntity(t *testing.T) {
	t.Parallel()

	req := &gen.UpdateRulePostRequest{}
	req.SetRegex(`x`)
	req.SetIncludedUsers("a")
	req.SetDeniedUsers("b")
	req.SetIncludedChannels("c")
	req.SetDeniedChannels("d")

	ent := updateRulePostReqToEntity(req)
	assert.Equal(t, `x`, ent.Regex)
	assert.Equal(t, "a", ent.IncludedUsers)
	assert.Equal(t, "b", ent.DeniedUsers)
	assert.Equal(t, "c", ent.IncludedChannels)
	assert.Equal(t, "d", ent.DeniedChannels)
}

func ptrInt64(v int64) *int64 {
	return &v
}
