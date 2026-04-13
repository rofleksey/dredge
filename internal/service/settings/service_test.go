package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestCreateRuleRejectsInvalidRegex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, obs)
	_, err := svc.CreateRule(context.Background(), entity.Rule{Regex: "("})
	require.Error(t, err)
}

func TestService_ListTwitchUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, obs)

	repo.EXPECT().ListTwitchUsers(gomock.Any()).Return([]entity.TwitchUser{{ID: 1}}, nil)

	out, err := svc.ListTwitchUsers(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestService_CreateTwitchUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().CreateTwitchUser(gomock.Any(), int64(10), "name").Return(entity.TwitchUser{ID: 10}, nil)

	u, err := svc.CreateTwitchUser(context.Background(), 10, "name")
	require.NoError(t, err)
	require.Equal(t, int64(10), u.ID)
}

func TestService_PatchTwitchUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	m := true

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(1)).Return(entity.TwitchUser{ID: 1, IrcOnlyWhenLive: true}, nil)
	repo.EXPECT().PatchTwitchUser(gomock.Any(), int64(1), entity.TwitchUserPatch{Monitored: &m}).Return(entity.TwitchUser{ID: 1, Monitored: true}, nil)

	u, err := svc.PatchTwitchUser(context.Background(), 1, entity.TwitchUserPatch{Monitored: &m})
	require.NoError(t, err)
	require.True(t, u.Monitored)
}

func TestService_PatchTwitchUser_rejectsNotifyOffWithoutLiveOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	on := true

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(2)).Return(entity.TwitchUser{
		ID:                      2,
		IrcOnlyWhenLive:         false,
		NotifyOffStreamMessages: false,
	}, nil)

	_, err := svc.PatchTwitchUser(context.Background(), 2, entity.TwitchUserPatch{NotifyOffStreamMessages: &on})
	require.ErrorIs(t, err, entity.ErrInvalidTwitchUserMonitorSettings)
}

func TestService_PatchTwitchUser_coercesNotifyOffWhenDisablingLiveOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(3)).Return(entity.TwitchUser{
		ID:                      3,
		IrcOnlyWhenLive:         true,
		NotifyOffStreamMessages: true,
	}, nil)

	ircOff := false
	notifyOff := false

	repo.EXPECT().PatchTwitchUser(gomock.Any(), int64(3), entity.TwitchUserPatch{
		IrcOnlyWhenLive:         &ircOff,
		NotifyOffStreamMessages: &notifyOff,
	}).Return(entity.TwitchUser{
		ID:                      3,
		IrcOnlyWhenLive:         false,
		NotifyOffStreamMessages: false,
	}, nil)

	u, err := svc.PatchTwitchUser(context.Background(), 3, entity.TwitchUserPatch{IrcOnlyWhenLive: &ircOff})
	require.NoError(t, err)
	require.False(t, u.IrcOnlyWhenLive)
	require.False(t, u.NotifyOffStreamMessages)
}

func TestService_ListRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().ListRules(gomock.Any()).Return([]entity.Rule{{ID: 1, Regex: "x"}}, nil)

	rules, err := svc.ListRules(context.Background())
	require.NoError(t, err)
	require.Len(t, rules, 1)
}

func TestService_CountRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().CountRules(gomock.Any()).Return(int64(4), nil)

	n, err := svc.CountRules(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(4), n)
}

func TestService_CreateRule_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	r := entity.Rule{Regex: `foo`}
	repo.EXPECT().CreateRule(gomock.Any(), r).Return(entity.Rule{ID: 1, Regex: "foo"}, nil)

	out, err := svc.CreateRule(context.Background(), r)
	require.NoError(t, err)
	require.Equal(t, int64(1), out.ID)
}

func TestService_UpdateRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	r := entity.Rule{Regex: `bar`}
	repo.EXPECT().UpdateRule(gomock.Any(), int64(2), r).Return(entity.Rule{ID: 2}, nil)

	out, err := svc.UpdateRule(context.Background(), 2, r)
	require.NoError(t, err)
	require.Equal(t, int64(2), out.ID)
}

func TestService_DeleteRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().DeleteRule(gomock.Any(), int64(3)).Return(nil)

	require.NoError(t, svc.DeleteRule(context.Background(), 3))
}

func TestService_Notifications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().ListNotificationEntries(gomock.Any()).Return([]entity.NotificationEntry{{ID: 1}}, nil)

	list, err := svc.ListNotifications(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)

	repo.EXPECT().CreateNotificationEntry(gomock.Any(), "telegram", map[string]any{}, true).Return(entity.NotificationEntry{ID: 2}, nil)

	created, err := svc.CreateNotification(context.Background(), "telegram", map[string]any{}, true)
	require.NoError(t, err)
	require.Equal(t, int64(2), created.ID)

	en := true
	repo.EXPECT().UpdateNotificationEntry(gomock.Any(), int64(2), nil, map[string]any{"a": 1}, &en).Return(entity.NotificationEntry{ID: 2}, nil)
	_, err = svc.UpdateNotification(context.Background(), 2, nil, map[string]any{"a": 1}, &en)
	require.NoError(t, err)

	repo.EXPECT().DeleteNotificationEntry(gomock.Any(), int64(2)).Return(nil)
	require.NoError(t, svc.DeleteNotification(context.Background(), 2))
}

func TestService_TwitchAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().ListTwitchAccounts(gomock.Any()).Return([]entity.TwitchAccount{{ID: 1}}, nil)
	accs, err := svc.ListTwitchAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, accs, 1)

	repo.EXPECT().CountTwitchAccounts(gomock.Any()).Return(int64(1), nil)
	n, err := svc.CountTwitchAccounts(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	repo.EXPECT().CreateTwitchAccount(gomock.Any(), int64(2), "u", "rt", "bot").Return(entity.TwitchAccount{ID: 2}, nil)
	a, err := svc.CreateTwitchAccount(context.Background(), 2, "u", "rt", "bot")
	require.NoError(t, err)
	require.Equal(t, int64(2), a.ID)

	bt := "main"
	repo.EXPECT().PatchTwitchAccount(gomock.Any(), int64(2), &bt).Return(entity.TwitchAccount{ID: 2}, nil)
	_, err = svc.PatchTwitchAccount(context.Background(), 2, &bt)
	require.NoError(t, err)

	repo.EXPECT().DeleteTwitchAccount(gomock.Any(), int64(2)).Return(nil)
	require.NoError(t, svc.DeleteTwitchAccount(context.Background(), 2))
}
