package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
)

func TestLogin(t *testing.T) {
	var cfg config.Config

	cfg.Admin.Email = "a@b.c"
	cfg.Admin.Password = "secret12345"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc, err := New(cfg, "1234567890123456", time.Hour, obs)
	require.NoError(t, err)

	token, err := svc.Login(context.Background(), "a@b.c", "secret12345")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestLogin_wrongEmail(t *testing.T) {
	var cfg config.Config

	cfg.Admin.Email = "a@b.c"
	cfg.Admin.Password = "secret12345"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc, err := New(cfg, "1234567890123456", time.Hour, obs)
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), "other@b.c", "secret12345")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidCredentials))
	require.True(t, errors.Is(err, entity.ErrNoSentry))
}

func TestLogin_wrongPassword(t *testing.T) {
	var cfg config.Config

	cfg.Admin.Email = "a@b.c"
	cfg.Admin.Password = "secret12345"

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	svc, err := New(cfg, "1234567890123456", time.Hour, obs)
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), "a@b.c", "wrong-password")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidCredentials))
}
