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
	"github.com/rofleksey/dredge/internal/http/authctx"
	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/observability"
	repomocks "github.com/rofleksey/dredge/internal/repository/mocks"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/usecase/settings"
	twitchuc "github.com/rofleksey/dredge/internal/usecase/twitch"
)

func TestHandler_Me_ok(t *testing.T) {
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

	tok, err := authSvc.Login(context.Background(), "admin@example.com", "password123")
	require.NoError(t, err)

	uid, role, err := authSvc.ParseToken(context.Background(), tok)
	require.NoError(t, err)

	ctx := authctx.WithRole(authctx.WithUserID(context.Background(), uid), role)

	res, err := h.Me(ctx)
	require.NoError(t, err)

	ac, ok := res.(*gen.Account)
	require.True(t, ok)
	assert.Equal(t, uid, ac.ID)
	assert.Equal(t, "admin", ac.Role)
}

func TestHandler_Me_unauthorized(t *testing.T) {
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

	res, err := h.Me(context.Background())
	require.NoError(t, err)

	_, ok := res.(*gen.MeUnauthorized)
	require.True(t, ok)
}
