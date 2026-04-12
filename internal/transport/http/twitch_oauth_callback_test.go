package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/observability"
)

func TestTwitchOAuthCallback_methodNotAllowed(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	h := NewTwitchOAuthCallback(nil, nil, obs)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/oauth/twitch/callback", nil)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestTwitchOAuthCallback_oauthNotConfigured(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	h := NewTwitchOAuthCallback(nil, nil, obs)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/oauth/twitch/callback", nil)
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestWithTwitchOAuthQuery(t *testing.T) {
	t.Parallel()

	u := withTwitchOAuthQuery("http://localhost:8080/#/settings", "error", "bad")
	assert.Contains(t, u, "#/settings?")
	assert.Contains(t, u, "twitch_oauth_error=bad")
}

func TestWithTwitchOAuthQuery_invalidBase(t *testing.T) {
	t.Parallel()

	u := withTwitchOAuthQuery(":", "k", "v")
	require.Equal(t, ":", u)
}
