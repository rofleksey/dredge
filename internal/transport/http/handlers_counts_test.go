package httptransport

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"github.com/rofleksey/dredge/internal/service/auth"
	"github.com/rofleksey/dredge/internal/service/settings"
	"github.com/rofleksey/dredge/internal/service/twitch"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

type noopBroadcaster struct{}

func (noopBroadcaster) BroadcastJSON(any) {}

func testHandler(t *testing.T) (*Handler, *gomock.Controller, *repomocks.MockStore) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockStore(ctrl)

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	twSvc := twitch.New(repo, noopBroadcaster{}, testTwitchServiceConfig("cid", "sec"), obs)
	setSvc := settings.New(repo, obs)

	h := NewHandler(authSvc, setSvc, twSvc, nil, obs)

	return h, ctrl, repo
}

func TestHandler_CountTwitchMessages(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().CountChatMessages(gomock.Any(), gomock.Any()).Return(int64(42), nil)

	res, err := h.CountTwitchMessages(context.Background(), gen.CountTwitchMessagesParams{})
	require.NoError(t, err)
	assert.Equal(t, int64(42), res.Total)
}

func TestHandler_CountTwitchDirectoryUsers(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().CountTwitchUsersBrowse(gomock.Any(), gomock.Any()).Return(int64(3), nil)

	res, err := h.CountTwitchDirectoryUsers(context.Background(), gen.CountTwitchDirectoryUsersParams{})
	require.NoError(t, err)
	assert.Equal(t, int64(3), res.Total)
}

func TestHandler_CountTwitchAccounts(t *testing.T) {
	h, ctrl, repo := testHandler(t)
	defer ctrl.Finish()

	repo.EXPECT().CountTwitchAccounts(gomock.Any()).Return(int64(2), nil)

	res, err := h.CountTwitchAccounts(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), res.Total)
}
