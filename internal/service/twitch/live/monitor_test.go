package live

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestReconcileIRCJoins_emptyMonitoredNoClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockStore(ctrl)
	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return(nil, nil)

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	r := NewRuntime(Config{
		Repo: repo,
		Obs:  obs,
	})

	require.NotPanics(t, func() {
		r.ReconcileIRCJoins(context.Background())
	})
}
