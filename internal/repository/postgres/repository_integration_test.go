package postgres

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
)

func TestRepository_integration(t *testing.T) {
	if testing.Short() {
		t.Skip("embedded postgres integration test")
	}

	ctx := context.Background()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := uint32(ln.Addr().(*net.TCPAddr).Port)
	require.NoError(t, ln.Close())

	cfg := embeddedpostgres.DefaultConfig().
		Username("postgres").
		Password("postgres").
		Database("dredge").
		Port(port).
		StartTimeout(3 * time.Minute).
		Logger(io.Discard)

	ep := embeddedpostgres.NewDatabase(cfg)
	require.NoError(t, ep.Start())

	defer func() { _ = ep.Stop() }()

	pool, err := pgxpool.New(ctx, cfg.GetConnectionURL())
	require.NoError(t, err)

	defer pool.Close()

	otel.SetTracerProvider(sdktrace.NewTracerProvider())

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	require.NoError(t, RunMigrations(ctx, pool))
	require.NoError(t, RunMigrations(ctx, pool))

	repo := New(pool, obs)

	const (
		channelID = int64(100)
		chatterID = int64(200)
		otherID   = int64(300)
		oauthUID  = int64(400)
	)

	_, err = repo.CreateTwitchUser(ctx, channelID, "channel1")
	require.NoError(t, err)

	_, err = repo.CreateTwitchUser(ctx, chatterID, "chatter1")
	require.NoError(t, err)

	_, err = repo.CreateTwitchUser(ctx, otherID, "otheruser")
	require.NoError(t, err)

	_, err = repo.CreateTwitchUser(ctx, oauthUID, "oauthuser")
	require.NoError(t, err)

	acc, err := repo.CreateTwitchAccount(ctx, oauthUID, "oauthuser", "refresh-token", "main")
	require.NoError(t, err)

	byTwitch, err := repo.GetTwitchAccountByTwitchUserID(ctx, oauthUID)
	require.NoError(t, err)
	assert.Equal(t, acc.ID, byTwitch.ID)

	_, err = repo.GetTwitchAccountByTwitchUserID(ctx, 777_777)
	assert.ErrorIs(t, err, entity.ErrTwitchAccountNotFound)

	_, err = repo.GetTwitchAccountByID(ctx, acc.ID)
	require.NoError(t, err)

	require.NoError(t, repo.UpdateTwitchRefreshToken(ctx, acc.ID, "new-refresh"))

	patchedAcc, err := repo.PatchTwitchAccount(ctx, acc.ID, entity.ToPointer("bot"))
	require.NoError(t, err)
	assert.Equal(t, "bot", patchedAcc.AccountType)

	accounts, err := repo.ListTwitchAccounts(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, accounts)

	nAcc, err := repo.CountTwitchAccounts(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, nAcc, int64(1))

	rule, err := repo.CreateRule(ctx, entity.Rule{
		Regex:            `hello`,
		IncludedUsers:    "*",
		DeniedUsers:      "",
		IncludedChannels: "*",
		DeniedChannels:   "",
	})
	require.NoError(t, err)

	rules, err := repo.ListRules(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, rules)

	nRules, err := repo.CountRules(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, nRules, int64(1))

	_, err = repo.UpdateRule(ctx, rule.ID, entity.Rule{
		Regex:            `world`,
		IncludedUsers:    "*",
		DeniedUsers:      "",
		IncludedChannels: "*",
		DeniedChannels:   "",
	})
	require.NoError(t, err)

	notif, err := repo.CreateNotificationEntry(ctx, "telegram", map[string]any{"k": "v"}, true)
	require.NoError(t, err)

	entries, err := repo.ListNotificationEntries(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, entries)

	enabled, err := repo.ListEnabledNotificationEntries(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, enabled)

	_, err = repo.UpdateNotificationEntry(ctx, notif.ID, entity.ToPointer("webhook"), map[string]any{"u": "x"}, entity.ToPointer(false))
	require.NoError(t, err)

	_, err = repo.InsertChatMessage(ctx, 0, nil, "x", "b", false, "irc", nil)
	require.Error(t, err)

	_, err = repo.InsertChatMessage(ctx, channelID, nil, "x", "b", false, "", nil)
	require.Error(t, err)

	_, err = repo.InsertChatMessage(ctx, channelID, nil, "   ", "b", false, "irc", nil)
	require.Error(t, err)

	_, err = repo.InsertChatMessageForChannelLogin(ctx, "  ", nil, "u", "b", false, "irc", nil)
	require.Error(t, err)

	chatterPtr := chatterID
	msgID, err := repo.InsertChatMessage(ctx, channelID, &chatterPtr, "chatter1", "hello world", true, "irc", []string{"moderator"})
	require.NoError(t, err)
	assert.Greater(t, msgID, int64(0))

	_, err = repo.InsertChatMessageForChannelLogin(ctx, "channel1", &chatterPtr, "chatter1", "second", false, "irc", nil)
	require.NoError(t, err)

	ok, err := repo.IsMonitoredChannel(ctx, "#channel1")
	require.NoError(t, err)
	assert.True(t, ok)

	okEmpty, err := repo.IsMonitoredChannel(ctx, "")
	require.NoError(t, err)
	assert.False(t, okEmpty)

	monID, monOK, err := repo.MonitoredChannelTwitchUserID(ctx, "#channel1")
	require.NoError(t, err)
	assert.True(t, monOK)
	assert.Equal(t, channelID, monID)

	_, notMonOK, err := repo.MonitoredChannelTwitchUserID(ctx, "nosuchchannel_xyz")
	require.NoError(t, err)
	assert.False(t, notMonOK)

	hist, err := repo.ListChatHistory(ctx, "channel1", 10)
	require.NoError(t, err)
	assert.NotEmpty(t, hist)

	msgs, err := repo.ListChatMessages(ctx, entity.ChatMessageListFilter{
		Channel: "channel1",
		Limit:   10,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, msgs)

	cnt, err := repo.CountChatMessages(ctx, entity.ChatMessageListFilter{Channel: "channel1"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cnt, int64(1))

	msgsF, err := repo.ListChatMessages(ctx, entity.ChatMessageListFilter{
		Username: "chatter",
		Channel:  "channel1",
		Text:     "hello",
		Limit:    50,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, msgsF)

	nilHelixCA, nilHelixHF, nilImg, err := repo.GetHelixMeta(ctx, 999_999)
	require.NoError(t, err)
	assert.Nil(t, nilHelixCA)
	assert.Nil(t, nilHelixHF)
	assert.Nil(t, nilImg)

	inserted, err := repo.UpsertTwitchUserFromChat(ctx, 500, "freshuser")
	require.NoError(t, err)
	assert.True(t, inserted)

	_, err = repo.UpsertTwitchUserFromChat(ctx, 500, "freshuser")
	require.NoError(t, err)

	_, err = repo.PatchTwitchUser(ctx, chatterID, entity.TwitchUserPatch{Marked: entity.ToPointer(true)})
	require.NoError(t, err)

	marked, err := repo.IsTwitchUserMarked(ctx, chatterID)
	require.NoError(t, err)
	assert.True(t, marked)

	_, err = repo.SetTwitchUserMonitored(ctx, otherID, false)
	require.NoError(t, err)

	allUsers, err := repo.ListTwitchUsers(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, allUsers)

	mon, err := repo.ListMonitoredTwitchUsers(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, mon)

	idByName, err := repo.TwitchUserIDByUsername(ctx, "channel1")
	require.NoError(t, err)
	assert.Equal(t, channelID, idByName)

	browse, err := repo.ListTwitchUsersBrowse(ctx, entity.TwitchUserBrowseFilter{Username: "chan", Limit: 20})
	require.NoError(t, err)
	assert.NotEmpty(t, browse)

	if len(browse) > 0 {
		last := browse[len(browse)-1]
		cur := last.ID
		m := last.Marked
		_, err = repo.ListTwitchUsersBrowse(ctx, entity.TwitchUserBrowseFilter{
			Username:     "",
			Limit:        10,
			CursorID:     &cur,
			CursorMarked: &m,
		})
		require.NoError(t, err)
	}

	nBrowse, err := repo.CountTwitchUsersBrowse(ctx, entity.TwitchUserBrowseFilter{Username: "ch"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, nBrowse, int64(1))

	u, err := repo.GetTwitchUserByID(ctx, channelID)
	require.NoError(t, err)
	assert.Equal(t, "channel1", u.Username)

	nChatterMsgs, err := repo.CountChatMessagesByChatter(ctx, chatterID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, nChatterMsgs, int64(1))

	require.NoError(t, repo.TruncateChannelChatters(ctx))

	require.NoError(t, repo.ReplaceChannelChattersSnapshot(ctx, channelID, []int64{chatterID, otherID}))

	cids, err := repo.ListChannelChatterIDs(ctx, channelID)
	require.NoError(t, err)
	assert.Len(t, cids, 2)

	nChatters, err := repo.CountChannelChatters(ctx, channelID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), nChatters)

	chatters, err := repo.ListChannelChatterEntries(ctx, channelID)
	require.NoError(t, err)
	assert.Len(t, chatters, 2)

	ch100 := channelID
	require.NoError(t, repo.InsertUserActivityEvent(ctx, chatterID, entity.UserActivityChatOnline, &ch100, map[string]any{"x": 1}))

	now := time.Now().UTC()
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	tl, err := repo.ListUserActivityEventsForTimeline(ctx, chatterID, from, to)
	require.NoError(t, err)
	assert.NotEmpty(t, tl)

	actList, err := repo.ListUserActivityEvents(ctx, entity.UserActivityListFilter{
		ChatterUserID: chatterID,
		Limit:         20,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, actList)

	created := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	require.NoError(t, repo.UpsertHelixMeta(ctx, channelID, &created, nil, now))

	ac, hf, imgURL, err := repo.GetHelixMeta(ctx, channelID)
	require.NoError(t, err)
	require.NotNil(t, ac)
	require.NotNil(t, hf)
	assert.Nil(t, imgURL)

	followed := now.Add(-24 * time.Hour)
	require.NoError(t, repo.UpsertChannelFollow(ctx, chatterID, channelID, &followed, now))

	follows, err := repo.ListFollowedMonitoredChannels(ctx, chatterID)
	require.NoError(t, err)
	assert.NotEmpty(t, follows)

	distinct, err := repo.ListDistinctChattersWithMessages(ctx, 50)
	require.NoError(t, err)
	assert.NotEmpty(t, distinct)

	pairs, err := repo.ListChatterChannelPairsForFollowEnrichment(ctx, 50)
	require.NoError(t, err)
	assert.NotEmpty(t, pairs)

	err = repo.DeleteRule(ctx, 999_999)
	assert.ErrorIs(t, err, entity.ErrRuleNotFound)

	_, err = repo.UpdateRule(ctx, 888_888, entity.Rule{Regex: `z`, IncludedUsers: "*", IncludedChannels: "*"})
	assert.ErrorIs(t, err, entity.ErrRuleNotFound)

	require.NoError(t, repo.DeleteRule(ctx, rule.ID))

	_, err = repo.UpdateNotificationEntry(ctx, 888_888, nil, map[string]any{}, entity.ToPointer(true))
	assert.ErrorIs(t, err, entity.ErrNotificationNotFound)

	require.NoError(t, repo.DeleteNotificationEntry(ctx, notif.ID))

	err = repo.DeleteNotificationEntry(ctx, 999_999)
	assert.ErrorIs(t, err, entity.ErrNotificationNotFound)

	err = repo.UpdateTwitchRefreshToken(ctx, 888_888, "tok")
	assert.ErrorIs(t, err, entity.ErrTwitchAccountNotFound)

	bt := "main"
	_, err = repo.PatchTwitchAccount(ctx, 888_888, &bt)
	assert.ErrorIs(t, err, entity.ErrTwitchAccountNotFound)

	require.NoError(t, repo.DeleteTwitchAccount(ctx, acc.ID))

	_, err = repo.GetTwitchAccountByID(ctx, acc.ID)
	assert.ErrorIs(t, err, entity.ErrTwitchAccountNotFound)

	_, err = repo.GetTwitchUserByID(ctx, 999_999)
	assert.ErrorIs(t, err, entity.ErrTwitchUserNotFound)

	_, err = repo.SetTwitchUserMonitored(ctx, 888_888, true)
	assert.ErrorIs(t, err, entity.ErrTwitchUserNotFound)

	_, err = repo.PatchTwitchUser(ctx, 888_888, entity.TwitchUserPatch{Monitored: entity.ToPointer(true)})
	assert.ErrorIs(t, err, entity.ErrTwitchUserNotFound)

	_, err = repo.PatchTwitchUser(ctx, chatterID, entity.TwitchUserPatch{})
	require.NoError(t, err)

	require.NoError(t, repo.AddChannelBlacklist(ctx, "badstreamer"))
	bl, err := repo.ListChannelBlacklist(ctx)
	require.NoError(t, err)
	assert.Contains(t, bl, "badstreamer")
	require.NoError(t, repo.RemoveChannelBlacklist(ctx, "badstreamer"))
	bl2, err := repo.ListChannelBlacklist(ctx)
	require.NoError(t, err)
	assert.NotContains(t, bl2, "badstreamer")

	ss, err := repo.GetSuspicionSettings(ctx)
	require.NoError(t, err)
	assert.True(t, ss.AutoCheckAccountAge)
	require.NoError(t, repo.UpdateSuspicionSettings(ctx, entity.SuspicionSettings{
		AutoCheckAccountAge: false,
		AccountAgeSusDays:   ss.AccountAgeSusDays,
		AutoCheckBlacklist:  ss.AutoCheckBlacklist,
		AutoCheckLowFollows: ss.AutoCheckLowFollows,
		LowFollowsThreshold: ss.LowFollowsThreshold,
		MaxGQLFollowPages:   ss.MaxGQLFollowPages,
	}))
	ss2, err := repo.GetSuspicionSettings(ctx)
	require.NoError(t, err)
	assert.False(t, ss2.AutoCheckAccountAge)

	require.NoError(t, repo.ReplaceUserFollowedChannels(ctx, chatterID, []entity.FollowedChannelRow{
		{FollowedChannelID: 9001, FollowedChannelLogin: "foo", FollowedAt: nil},
	}))
	gqlFollows, err := repo.ListUserFollowedChannels(ctx, chatterID)
	require.NoError(t, err)
	require.Len(t, gqlFollows, 1)
	assert.Equal(t, "foo", gqlFollows[0].FollowedChannelLogin)

	_ = msgID
}
