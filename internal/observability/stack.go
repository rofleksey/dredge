package observability

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Stack struct {
	Logger            *zap.Logger
	LogLoggerProvider *sdklog.LoggerProvider
	TracerProvider    *sdktrace.TracerProvider
	Tracer            trace.Tracer
	requestsTotal     *prometheus.CounterVec
	requestLatency    *prometheus.HistogramVec
}

func (s *Stack) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return s.Tracer.Start(ctx, name)
}

// Zap implements repository.Instrumentation (field is named Logger to avoid clashing with this method).
func (s *Stack) Zap() *zap.Logger {
	return s.Logger
}
