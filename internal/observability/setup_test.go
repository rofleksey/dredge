package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rofleksey/dredge/internal/config"
)

func TestSetup_defaults(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Observability.ServiceName = "dredge-test"
	cfg.Observability.LogLevel = ""
	cfg.Observability.LogExporter = "none"
	cfg.Observability.TraceExporter = "none"

	s, err := Setup(cfg)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NotNil(t, s.Logger)
	require.NotNil(t, s.Tracer)

	if s.TracerProvider != nil {
		_ = s.TracerProvider.Shutdown(t.Context())
	}
}

func TestSetup_invalidLogLevel(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Observability.ServiceName = "x"
	cfg.Observability.LogLevel = "not-a-real-level"

	_, err := Setup(cfg)
	assert.Error(t, err)
}

func TestSetup_unknownLogExporter(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Observability.ServiceName = "x"
	cfg.Observability.LogLevel = "info"
	cfg.Observability.LogExporter = "nope"
	cfg.Observability.TraceExporter = "none"

	_, err := Setup(cfg)
	assert.Error(t, err)
}

func TestSetup_unknownTraceExporter(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	cfg.Observability.ServiceName = "x"
	cfg.Observability.LogLevel = "info"
	cfg.Observability.LogExporter = "none"
	cfg.Observability.TraceExporter = "teleport"

	_, err := Setup(cfg)
	assert.Error(t, err)
}
