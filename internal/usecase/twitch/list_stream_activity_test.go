package twitch

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_ListStreamActivity(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("c", "s"), obs)

	st := time.Now().UTC().Add(-time.Hour)
	repo.EXPECT().GetMonitoredStreamByID(gomock.Any(), int64(7)).Return(entity.Stream{
		ChannelTwitchUserID: 1,
		StartedAt:           st,
	}, nil)
	repo.EXPECT().ListUserActivityForStream(gomock.Any(), gomock.Any()).Return(nil, nil)

	_, err := svc.ListStreamActivity(context.Background(), 7, 20, nil, nil)
	require.NoError(t, err)
}
