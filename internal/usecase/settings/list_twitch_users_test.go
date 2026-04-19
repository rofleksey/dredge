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

func TestService_ListTwitchUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, obs)

	repo.EXPECT().ListTwitchUsers(gomock.Any()).Return([]entity.TwitchUser{{ID: 1}}, nil)

	out, err := svc.ListTwitchUsers(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestService_ListTwitchUsers_monitoredOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, obs)

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return([]entity.TwitchUser{{ID: 2, Monitored: true}}, nil)

	out, err := svc.ListTwitchUsers(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(2), out[0].ID)
	require.True(t, out[0].Monitored)
}
