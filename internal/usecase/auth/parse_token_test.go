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

func TestParseToken_roundTrip(t *testing.T) {
	var cfg config.Config

	cfg.Admin.Email = "a@b.c"
	cfg.Admin.Password = "secret12345"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc, err := New(cfg, "1234567890123456", time.Hour, obs)
	require.NoError(t, err)

	token, err := svc.Login(context.Background(), "a@b.c", "secret12345")
	require.NoError(t, err)

	id, role, err := svc.ParseToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, int64(1), id)
	require.Equal(t, "admin", role)
}

func TestParseToken_invalid(t *testing.T) {
	var cfg config.Config

	cfg.Admin.Email = "a@b.c"
	cfg.Admin.Password = "secret12345"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc, err := New(cfg, "1234567890123456", time.Hour, obs)
	require.NoError(t, err)

	_, _, err = svc.ParseToken(context.Background(), "not-a-jwt")
	require.Error(t, err)
}
