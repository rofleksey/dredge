package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/usecase/auth"
	"github.com/rofleksey/dredge/internal/ws"
)

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
