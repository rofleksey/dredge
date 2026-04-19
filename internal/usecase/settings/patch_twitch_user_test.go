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

func TestService_PatchTwitchUser_rejectsNotifyOffWhenLiveOnlyEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	on := true

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(2)).Return(entity.TwitchUser{
		ID:                      2,
		IrcOnlyWhenLive:         true,
		NotifyOffStreamMessages: false,
	}, nil)

	_, err := svc.PatchTwitchUser(context.Background(), 2, entity.TwitchUserPatch{NotifyOffStreamMessages: &on})
	require.ErrorIs(t, err, entity.ErrInvalidTwitchUserMonitorSettings)
}

func TestService_PatchTwitchUser_allowsNotifyOffWhenLiveOnlyDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	on := true

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(4)).Return(entity.TwitchUser{
		ID:                      4,
		IrcOnlyWhenLive:         false,
		NotifyOffStreamMessages: false,
	}, nil)

	repo.EXPECT().PatchTwitchUser(gomock.Any(), int64(4), entity.TwitchUserPatch{
		NotifyOffStreamMessages: &on,
	}).Return(entity.TwitchUser{
		ID:                      4,
		IrcOnlyWhenLive:         false,
		NotifyOffStreamMessages: true,
	}, nil)

	u, err := svc.PatchTwitchUser(context.Background(), 4, entity.TwitchUserPatch{NotifyOffStreamMessages: &on})
	require.NoError(t, err)
	require.True(t, u.NotifyOffStreamMessages)
}

func TestService_PatchTwitchUser_coercesNotifyOffWhenEnablingLiveOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	svc := New(repo, &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")})

	repo.EXPECT().GetTwitchUserByID(gomock.Any(), int64(3)).Return(entity.TwitchUser{
		ID:                      3,
		IrcOnlyWhenLive:         false,
		NotifyOffStreamMessages: true,
	}, nil)

	ircOn := true
	notifyOff := false

	repo.EXPECT().PatchTwitchUser(gomock.Any(), int64(3), entity.TwitchUserPatch{
		IrcOnlyWhenLive:         &ircOn,
		NotifyOffStreamMessages: &notifyOff,
	}).Return(entity.TwitchUser{
		ID:                      3,
		IrcOnlyWhenLive:         true,
		NotifyOffStreamMessages: false,
	}, nil)

	u, err := svc.PatchTwitchUser(context.Background(), 3, entity.TwitchUserPatch{IrcOnlyWhenLive: &ircOn})
	require.NoError(t, err)
	require.True(t, u.IrcOnlyWhenLive)
	require.False(t, u.NotifyOffStreamMessages)
}
