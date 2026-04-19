package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/observability"
)

func TestNewEngine(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	e := NewEngine(Config{Obs: obs})
	require.NotNil(t, e)
}

func TestEngine_Reload(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	e := NewEngine(Config{Obs: obs})
	e.Reload(context.Background(), []entity.Rule{{ID: 1, EventType: EventChatMessage, Enabled: true}})
	require.NotNil(t, e.snapshot())
}

func TestEngine_StartStop(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	e := NewEngine(Config{Obs: obs})
	e.Start(context.Background())
	e.Stop()
	require.NotNil(t, e)
}

func TestEngine_Stop_withoutStart(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	e := NewEngine(Config{Obs: obs})
	e.Stop()
	require.NotNil(t, e)
}

func TestEngine_KeywordMatchChat_noRules(t *testing.T) {
	t.Parallel()

	obs := &observability.Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	e := NewEngine(Config{Obs: obs})
	ok := e.KeywordMatchChat(context.Background(), "ch", "u", "x")
	require.False(t, ok)
}
