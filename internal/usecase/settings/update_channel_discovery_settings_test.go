package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func TestService_UpdateChannelDiscoverySettings_invalidWhenEnabledEmptyGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, obs)

	_, err := svc.UpdateChannelDiscoverySettings(context.Background(), entity.ChannelDiscoverySettings{
		Enabled:             true,
		PollIntervalSeconds: 3600,
		GameID:              "   ",
	})
	require.ErrorIs(t, err, entity.ErrInvalidChannelDiscoverySettings)
}

func TestService_UpdateChannelDiscoverySettings_ok(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, obs)

	in := entity.ChannelDiscoverySettings{
		Enabled:              false,
		PollIntervalSeconds:  120,
		GameID:               "1",
		MinLiveViewers:       10,
		RequiredStreamTags:   []string{" A ", ""},
		MaxStreamPagesPerRun: 5,
	}

	repo.EXPECT().UpdateChannelDiscoverySettings(gomock.Any(), entity.ChannelDiscoverySettings{
		Enabled:              false,
		PollIntervalSeconds:  120,
		GameID:               "1",
		MinLiveViewers:       10,
		RequiredStreamTags:   []string{"A"},
		MaxStreamPagesPerRun: 5,
	}).Return(nil)

	repo.EXPECT().GetChannelDiscoverySettings(gomock.Any()).Return(entity.ChannelDiscoverySettings{
		Enabled:              false,
		PollIntervalSeconds:  120,
		GameID:               "1",
		MinLiveViewers:       10,
		RequiredStreamTags:   []string{"A"},
		MaxStreamPagesPerRun: 5,
	}, nil)

	out, err := svc.UpdateChannelDiscoverySettings(context.Background(), in)
	require.NoError(t, err)
	require.Equal(t, []string{"A"}, out.RequiredStreamTags)
}
