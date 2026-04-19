package live

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/repository"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"github.com/rofleksey/dredge/internal/service/twitch/helix"
)

type noopBC struct{}

func (noopBC) BroadcastJSON(any) {}

func testRuntime(t *testing.T, repo repository.Store) *Runtime {
	t.Helper()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	h := helix.NewClient(repo, obs, "c", "s")

	return NewRuntime(Config{
		Helix:          h,
		Repo:           repo,
		Obs:            obs,
		Broadcaster:    noopBC{},
		OnEnqueueUser:  func(int64) {},
		PersistContext: context.Background,
	})
}

func TestDispatchRuleHitNotifications_empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	r := testRuntime(t, repo)

	repo.EXPECT().ListEnabledNotificationEntries(gomock.Any()).Return(nil, nil)

	r.NotifyChatKeyword(context.Background(), "ch", "u", "msg", "")
}

func TestDispatchRuleHitNotifications_webhook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	r := testRuntime(t, repo)
	r.helix.HTTPClient = srv.Client()

	repo.EXPECT().ListEnabledNotificationEntries(gomock.Any()).Return([]entity.NotificationEntry{
		{Provider: "webhook", Settings: map[string]any{"url": srv.URL}},
	}, nil)

	r.NotifyChatKeyword(context.Background(), "ch", "u", "msg", "")

	time.Sleep(150 * time.Millisecond)
}

func TestNotifyRuleText_webhook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockStore(ctrl)
	r := testRuntime(t, repo)
	r.helix.HTTPClient = srv.Client()

	repo.EXPECT().ListEnabledNotificationEntries(gomock.Any()).Return([]entity.NotificationEntry{
		{Provider: "webhook", Settings: map[string]any{"url": srv.URL}},
	}, nil)

	r.NotifyRuleText(context.Background(), "ch", "hello interval")

	time.Sleep(150 * time.Millisecond)
}
