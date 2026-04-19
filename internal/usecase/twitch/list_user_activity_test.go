package twitch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

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
