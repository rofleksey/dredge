package twitch

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
)

func TestEnrichSingleUser_skipsWithinCooldown(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	recent := time.Now().UTC().Add(-1 * time.Hour)
	repo.EXPECT().GetIrcMonitorSettings(gomock.Any()).Return(entity.IrcMonitorSettings{
		EnrichmentCooldown: 24 * time.Hour,
	}, nil)
	repo.EXPECT().GetHelixMeta(gomock.Any(), int64(42)).Return(nil, &recent, nil, nil)

	svc.enrichSingleUser(context.Background(), 42)
}

func TestEnqueueMonitoredAndMarkedUsersForEnrichment(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc := New(repo, stopNoopBC{}, testTwitchCfg("cid", "csec"), obs)

	repo.EXPECT().ListMonitoredOrMarkedTwitchUserIDs(gomock.Any()).Return([]int64{11, 22}, nil)
	svc.EnqueueMonitoredAndMarkedUsersForEnrichment(context.Background())

	got1 := <-svc.enrichQueue
	got2 := <-svc.enrichQueue
	if got1 != 11 || got2 != 22 {
		t.Fatalf("unexpected queued ids: %d, %d", got1, got2)
	}
}
