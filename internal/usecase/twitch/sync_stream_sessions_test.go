package twitch

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestUsecase_syncStreamSessions_listError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	s := New(repo, nil, config.Config{}, obs)

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return(nil, errors.New("boom"))

	err := s.syncStreamSessions(context.Background())
	require.Error(t, err)
}
