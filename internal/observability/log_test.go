package observability

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func TestStack_LogError_nil(t *testing.T) {
	t.Parallel()

	s := &Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	ctx := context.Background()

	_, span := s.StartSpan(ctx, "t")

	defer span.End()

	s.LogError(ctx, span, "msg", nil)
}

func TestStack_LogError_withErr(t *testing.T) {
	t.Parallel()

	s := &Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	ctx := context.Background()

	_, span := s.StartSpan(ctx, "t")

	defer span.End()

	s.LogError(ctx, span, "msg", errors.New("boom"))
}

func TestStack_LogError_noSentry(t *testing.T) {
	t.Parallel()

	s := &Stack{Logger: zap.NewNop(), Tracer: otel.Tracer("test")}
	ctx := context.Background()

	_, span := s.StartSpan(ctx, "t")

	defer span.End()

	err := entity.ErrNoSentry
	require.Error(t, err)

	s.LogError(ctx, span, "expected", err)
}
