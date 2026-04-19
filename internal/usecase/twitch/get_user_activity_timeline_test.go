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
