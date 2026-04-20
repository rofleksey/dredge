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

func TestService_StartIrcJoinedSnapshotLoop_exitsOnCancel(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("c", "s"), obs)

	repo.EXPECT().ListMonitoredTwitchUsers(gomock.Any()).Return(nil, nil)
	repo.EXPECT().InsertIrcJoinedSample(gomock.Any(), 0).Return(nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})

	go func() {
		svc.StartIrcJoinedSnapshotLoop(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("irc joined snapshot loop did not exit")
	}
}
