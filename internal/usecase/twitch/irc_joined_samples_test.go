package twitch

import (
	"context"
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

func TestService_RecordIrcJoinedSnapshot(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("c", "s"), obs)

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return(nil, nil)
	repo.EXPECT().InsertIrcJoinedSample(gomock.Any(), 0).Return(nil)

	require.NoError(t, svc.RecordIrcJoinedSnapshot(context.Background()))
}

func TestService_ListIrcJoinedSamplesLastDays_defaultsAndCap(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("c", "s"), obs)

	repo.EXPECT().ListIrcJoinedSamples(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, from, to time.Time) ([]entity.IrcJoinedSample, error) {
			d := to.Sub(from)
			assert.GreaterOrEqual(t, d, 7*24*time.Hour-time.Hour)
			assert.LessOrEqual(t, d, 7*24*time.Hour+time.Hour)

			return nil, nil
		})

	_, err := svc.ListIrcJoinedSamplesLastDays(context.Background(), 0)
	require.NoError(t, err)

	repo.EXPECT().ListIrcJoinedSamples(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, from, to time.Time) ([]entity.IrcJoinedSample, error) {
			d := to.Sub(from)
			assert.GreaterOrEqual(t, d, 90*24*time.Hour-time.Hour)
			assert.LessOrEqual(t, d, 90*24*time.Hour+time.Hour)

			return nil, nil
		})

	_, err = svc.ListIrcJoinedSamplesLastDays(context.Background(), 500)
	require.NoError(t, err)
}
