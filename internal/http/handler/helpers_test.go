package handler

import (
	"context"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/authctx"
	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/rules"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func testTwitchServiceConfig(clientID, clientSecret string) config.Config {
	var c config.Config

	c.Twitch.ClientID = clientID
	c.Twitch.ClientSecret = clientSecret
	c.Twitch.OAuthRedirectURI = "http://localhost:8080/oauth/twitch/callback"
	c.Twitch.OAuthReturnURL = "http://localhost:5173/#/settings"

	return c
}

type noopBroadcaster struct{}

func (noopBroadcaster) BroadcastJSON(any) {}

func testHandler(t *testing.T) (*Handler, *gomock.Controller, *repomocks.MockStore) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockStore(ctrl)

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	twSvc := twitchuc.New(repo, noopBroadcaster{}, testTwitchServiceConfig("cid", "sec"), obs)
	setSvc := settings.New(repo, obs)

	rulesSvc := rules.NewUsecase(repo, obs, nil, nil)

	h := NewHandler(authSvc, setSvc, rulesSvc, twSvc, nil, nil, obs)

	return h, ctrl, repo
}

func adminCtx() context.Context {
	return authctx.WithRole(authctx.WithUserID(context.Background(), int64(1)), "admin")
}

func viewerCtx() context.Context {
	return authctx.WithRole(authctx.WithUserID(context.Background(), int64(2)), "viewer")
}

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
		FirstMessage:        true,
		CreatedAt:           time.Unix(1, 0).UTC(),
		BadgeTags:           []string{"moderator"},
	}
	g := chatHistoryEntityToGen(m)
	assert.Equal(t, gen.ChatHistoryEntrySourceSent, g.Source)
	assert.True(t, g.FirstMessage)
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

	req := &gen.CreateRuleRequest{
		Name:           "my rule",
		EventType:      gen.RuleEventTypeChatMessage,
		ActionType:     gen.RuleActionTypeNotify,
		EventSettings:  gen.CreateRuleRequestEventSettings{},
		Middlewares:    nil,
		ActionSettings: gen.CreateRuleRequestActionSettings{},
	}
	ent := createRuleReqToEntity(req)
	assert.Equal(t, "my rule", ent.Name)
	assert.True(t, ent.Enabled)
	assert.Equal(t, "chat_message", ent.EventType)
	assert.Equal(t, "notify", ent.ActionType)
	assert.True(t, ent.UseSharedPool)
}

func TestRuleEntityToGen(t *testing.T) {
	t.Parallel()

	r := entity.Rule{
		ID:             3,
		Name:           "n",
		Enabled:        true,
		EventType:      "chat_message",
		EventSettings:  map[string]any{},
		Middlewares:    nil,
		ActionType:     "notify",
		ActionSettings: map[string]any{},
		UseSharedPool:  true,
		CreatedAt:      time.Unix(1, 0).UTC(),
		UpdatedAt:      time.Unix(1, 0).UTC(),
	}
	g := ruleEntityToGen(r)
	assert.Equal(t, int64(3), g.ID)
	assert.Equal(t, "n", g.Name)
	assert.Equal(t, gen.RuleEventTypeChatMessage, g.EventType)
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

	req := &gen.UpdateRulePostRequest{
		ID:             1,
		Name:           "upd",
		Enabled:        true,
		EventType:      gen.RuleEventTypeChatMessage,
		EventSettings:  gen.UpdateRulePostRequestEventSettings{},
		Middlewares:    nil,
		ActionType:     gen.RuleActionTypeNotify,
		ActionSettings: gen.UpdateRulePostRequestActionSettings{},
		UseSharedPool:  true,
	}

	ent := updateRulePostReqToEntity(req)
	assert.Equal(t, "upd", ent.Name)
	assert.Equal(t, "chat_message", ent.EventType)
}

func ptrInt64(v int64) *int64 {
	return &v
}
