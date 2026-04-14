package twitch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_GetTwitchUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(5)).Return(entity.TwitchUser{ID: 5, Username: "u"}, nil)

	u, err := svc.GetTwitchUser(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), u.ID)
}

func TestService_ListChatHistory_notMonitored(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "x").Return(false, nil)

	_, err := svc.ListChatHistory(context.Background(), "x", 10)
	require.ErrorIs(t, err, ErrChannelNotMonitored)
}

func TestService_ListChatHistory_ok(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	repo.EXPECT().IsMonitoredChannel(gomock.Any(), "chan").Return(true, nil)
	repo.EXPECT().ListChatHistory(gomock.Any(), "chan", 5).Return([]entity.ChatHistoryMessage{{ID: 1}}, nil)

	msgs, err := svc.ListChatHistory(context.Background(), "chan", 5)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
}

func TestService_ListChatMessages(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	f := entity.ChatMessageListFilter{Limit: 10}
	repo.EXPECT().ListChatMessages(gomock.Any(), f).Return(nil, errors.New("db"))

	_, err := svc.ListChatMessages(context.Background(), f)
	require.Error(t, err)
}

func TestService_ListTwitchUsersBrowse(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	f := entity.TwitchUserBrowseFilter{Limit: 20}
	repo.EXPECT().ListTwitchUsersBrowse(gomock.Any(), f).Return([]entity.TwitchDirectoryEntry{{User: entity.TwitchUser{ID: 1}}}, nil)

	out, err := svc.ListTwitchUsersBrowse(context.Background(), f)
	require.NoError(t, err)
	assert.Len(t, out, 1)
}

func TestService_CountChatMessages(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	f := entity.ChatMessageListFilter{}
	repo.EXPECT().CountChatMessages(gomock.Any(), f).Return(int64(3), nil)

	n, err := svc.CountChatMessages(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestService_CountTwitchUsersBrowse(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	f := entity.TwitchUserBrowseFilter{}
	repo.EXPECT().CountTwitchUsersBrowse(gomock.Any(), f).Return(int64(9), nil)

	n, err := svc.CountTwitchUsersBrowse(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, int64(9), n)
}

func TestService_GetTwitchUserProfile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	now := time.Now().UTC()

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(1)).Return(entity.TwitchUser{ID: 1}, nil)
	repo.EXPECT().CountChatMessagesByChatter(gomock.Any(), int64(1)).Return(int64(2), nil)
	repo.EXPECT().ListUserActivityEventsForTimeline(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().GetHelixMeta(gomock.Any(), int64(1)).Return(&now, &now, nil, nil)
	repo.EXPECT().ListFollowedMonitoredChannels(gomock.Any(), int64(1)).Return(nil, nil)
	repo.EXPECT().ListUserFollowedChannels(gomock.Any(), int64(1)).Return(nil, nil)
	repo.EXPECT().ListChannelBlacklist(gomock.Any()).Return(nil, nil)

	u, n, pres, ac, _, follows, gqlFollows, bl, err := svc.GetTwitchUserProfile(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.Equal(t, int64(2), n)
	assert.Equal(t, int64(0), pres)
	assert.NotNil(t, ac)
	assert.Empty(t, follows)
	assert.Empty(t, gqlFollows)
	assert.Empty(t, bl)
}

func TestService_ListUserActivity(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	f := entity.UserActivityListFilter{ChatterUserID: 1, Limit: 10}
	repo.EXPECT().ListUserActivityEvents(gomock.Any(), f).Return([]entity.UserActivityEvent{{ID: 1}}, nil)

	out, err := svc.ListUserActivity(context.Background(), f)
	require.NoError(t, err)
	assert.Len(t, out, 1)
}

func TestService_GetUserActivityTimeline_swapsWindow(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("id", "secret"), obs)

	to := time.Now().UTC()
	from := to.Add(time.Hour) // after to — should be swapped internally

	repo.EXPECT().ListUserActivityEventsForTimeline(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return([]entity.UserActivityEvent{}, nil)

	segs, err := svc.GetUserActivityTimeline(context.Background(), 1, from, to)
	require.NoError(t, err)
	assert.Empty(t, segs)
}

func TestService_New(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, nil, testTwitchCfg("cid", "sec"), obs)
	require.NotNil(t, svc)
}
