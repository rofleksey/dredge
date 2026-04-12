package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func TestStack_StartSpan(t *testing.T) {
	t.Parallel()

	s := &Stack{
		Logger: zap.NewNop(),
		Tracer: otel.Tracer("test"),
	}

	ctx := context.Background()

	ctx2, span := s.StartSpan(ctx, "op")

	defer span.End()

	assert.NotEqual(t, ctx, ctx2)
	assert.NotNil(t, span)
}
