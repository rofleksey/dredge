package httptransport

import (
	"testing"
	"time"

	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/usecase/auth"
)

func TestIsUnauthorized(t *testing.T) {
	t.Parallel()

	assert.True(t, IsUnauthorized(ogenerrors.ErrSecurityRequirementIsNotSatisfied))
	assert.False(t, IsUnauthorized(assert.AnError))
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
