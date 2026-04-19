package twitch

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestService_StartEnrichmentWorker_exitsOnCancel(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("c", "s"), obs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc.StartEnrichmentWorker(ctx)

	time.Sleep(10 * time.Millisecond)
}
