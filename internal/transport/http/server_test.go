package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/service/auth"
	"github.com/rofleksey/dredge/internal/transport/http/gen"
	"github.com/rofleksey/dredge/internal/transport/ws"
)

func TestSecurity_HandleBearerAuth(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	tok, err := authSvc.Login(context.Background(), "admin@example.com", "password123")
	require.NoError(t, err)

	sec := NewSecurity(authSvc, obs)

	ctx, err := sec.HandleBearerAuth(context.Background(), gen.LoginOperation, gen.BearerAuth{Token: tok})
	require.NoError(t, err)
	require.NotNil(t, ctx)
}

func TestSecurity_HandleBearerAuth_invalid(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	sec := NewSecurity(authSvc, obs)

	_, err = sec.HandleBearerAuth(context.Background(), gen.LoginOperation, gen.BearerAuth{Token: "bad"})
	require.Error(t, err)
}

func TestIsUnauthorized(t *testing.T) {
	t.Parallel()

	assert.True(t, IsUnauthorized(ogenerrors.ErrSecurityRequirementIsNotSatisfied))
	assert.False(t, IsUnauthorized(assert.AnError))
}

func TestLiveWebsocketHandler_missingToken(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	h := LiveWebsocketHandler(authSvc, ws.NewHub(""), nil, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)

	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Admin.Email = "admin@example.com"
	cfg.Admin.Password = "password123"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	authSvc, err := auth.New(cfg, "12345678901234567890", time.Hour, obs)
	require.NoError(t, err)

	h := NewHandler(authSvc, nil, nil, nil, obs)
	require.NotNil(t, h)
}
