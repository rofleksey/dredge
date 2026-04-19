package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/observability"
)

func TestMe_okAndWrongAccount(t *testing.T) {
	var cfg config.Config

	cfg.Admin.Email = "a@b.c"
	cfg.Admin.Password = "secret12345"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc, err := New(cfg, "1234567890123456", time.Hour, obs)
	require.NoError(t, err)

	acc, err := svc.Me(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "a@b.c", acc.Email)
	require.Equal(t, "admin", acc.Role)

	_, err = svc.Me(context.Background(), 99)
	require.Error(t, err)
}
