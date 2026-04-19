package handler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/config"
	"github.com/rofleksey/dredge/internal/http/gen"
	"github.com/rofleksey/dredge/internal/observability"
	"github.com/rofleksey/dredge/internal/usecase/auth"
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
