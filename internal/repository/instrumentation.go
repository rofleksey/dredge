package repository

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Instrumentation is the subset of observability used by Store implementations (e.g. postgres).
type Instrumentation interface {
	StartSpan(ctx context.Context, name string) (context.Context, trace.Span)
	LogError(ctx context.Context, span trace.Span, msg string, err error, fields ...zap.Field)
	Zap() *zap.Logger
}
