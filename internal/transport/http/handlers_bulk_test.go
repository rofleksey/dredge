package httptransport

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func adminCtx() context.Context {
	return context.WithValue(context.WithValue(context.Background(), userIDCtxKey, int64(1)), roleCtxKey, "admin")
}

func TestHandler_ListChatHistory_ok(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "chan").Return(true, nil)
	repo.EXPECT().ListChatHistory(gomock.Any(), "chan", 50).Return([]entity.ChatHistoryMessage{
		{ID: 1, Channel: "chan", Username: "u", Message: "m", MsgType: "irc"},
	}, nil)

	res, err := h.ListChatHistory(context.Background(), gen.ListChatHistoryParams{Channel: "chan"})
	require.NoError(t, err)

	_, ok := res.(*gen.ListChatHistoryOKApplicationJSON)
	require.True(t, ok)
}

func TestHandler_ListChatHistory_notMonitored(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "x").Return(false, nil)

	res, err := h.ListChatHistory(context.Background(), gen.ListChatHistoryParams{Channel: "x"})
	require.NoError(t, err)

	_, ok := res.(*gen.ErrorMessage)
	require.True(t, ok)
}

func TestHandler_ListTwitchMessages(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListChatMessages(gomock.Any(), gomock.Any()).Return([]entity.ChatHistoryMessage{
		{ID: 1, Channel: "c", Username: "u", Message: "hi", MsgType: "irc"},
	}, nil)

	out, err := h.ListTwitchMessages(context.Background(), gen.ListTwitchMessagesParams{})
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestHandler_GetIrcMonitorStatus(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return(nil, nil)

	st, err := h.GetIrcMonitorStatus(adminCtx())
	require.NoError(t, err)
	assert.NotNil(t, st)
}

func TestHandler_ListNotifications(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().ListNotificationEntries(gomock.Any()).Return([]entity.NotificationEntry{
		{ID: 1, Provider: "telegram", Settings: map[string]any{}, Enabled: true},
	}, nil)

	out, err := h.ListNotifications(adminCtx())
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestHandler_GetTwitchUserProfile(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	now := time.Now().UTC()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(9)).Return(entity.TwitchUser{ID: 9, Username: "u"}, nil)
	repo.EXPECT().CountChatMessagesByChatter(gomock.Any(), int64(9)).Return(int64(1), nil)
	repo.EXPECT().ListUserActivityEventsForTimeline(gomock.Any(), int64(9), gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().GetHelixMeta(gomock.Any(), int64(9)).Return(&now, &now, nil, nil)
	repo.EXPECT().ListFollowedMonitoredChannels(gomock.Any(), int64(9)).Return(nil, nil)
	repo.EXPECT().ListUserFollowedChannels(gomock.Any(), int64(9)).Return(nil, nil)
	repo.EXPECT().ListChannelBlacklist(gomock.Any()).Return(nil, nil)

	res, err := h.GetTwitchUserProfile(context.Background(), &gen.GetTwitchUserProfileRequest{ID: 9})
	require.NoError(t, err)

	_, ok := res.(*gen.TwitchUserProfile)
	require.True(t, ok)
}
