package handler

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
	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func TestHandler_Login_ok(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	twRepo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	twSvc := twitchuc.New(twRepo, noopBroadcaster{}, testTwitchServiceConfig("c", "s"), obs)
	h := NewHandler(authSvc, settings.New(nil, obs), nil, twSvc, nil, nil, obs, nil)

	res, err := h.Login(context.Background(), &gen.LoginRequest{Email: "admin@example.com", Password: "password123"})
	require.NoError(t, err)

	ok, ok2 := res.(*gen.LoginResponse)
	require.True(t, ok2)
	assert.NotEmpty(t, ok.Token)
}

func TestHandler_Login_unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	twRepo := repomocks.NewMockStore(ctrl)
	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	twSvc := twitchuc.New(twRepo, noopBroadcaster{}, testTwitchServiceConfig("c", "s"), obs)
	h := NewHandler(authSvc, settings.New(nil, obs), nil, twSvc, nil, nil, obs, nil)

	res, err := h.Login(context.Background(), &gen.LoginRequest{Email: "admin@example.com", Password: "wrong"})
	require.NoError(t, err)

	_, ok := res.(*gen.LoginUnauthorized)
	require.True(t, ok)
}
