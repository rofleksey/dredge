package observability

import (
	"context"
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rofleksey/dredge/internal/config"
)

func Setup(cfg config.Config) (*Stack, error) {
	if cfg.Observability.LogLevel == "" {
		cfg.Observability.LogLevel = "debug"
	}

	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Observability.LogLevel)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	ctx := context.Background()

	encCfg := zap.NewDevelopmentEncoderConfig()
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encCfg),
		zapcore.AddSync(os.Stderr),
		level,
	)

	logExporter := cfg.Observability.LogExporter
	if logExporter == "" {
		logExporter = "none"
	}

	traceExporter := cfg.Observability.TraceExporter
	if traceExporter == "" {
		traceExporter = "none"
	}

	var otelRes *resource.Resource

	if traceExporter == "stdout" || logExporter == "otlp" {
		var err error

		otelRes, err = resource.New(ctx,
			resource.WithAttributes(semconv.ServiceName(cfg.Observability.ServiceName)),
		)
		if err != nil {
			return nil, fmt.Errorf("otel resource: %w", err)
		}
	}

	var logLP *sdklog.LoggerProvider

	core := consoleCore

	switch logExporter {
	case "none":
	case "otlp":
		exp, err := otlploghttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("otlp log exporter: %w", err)
		}

		processor := sdklog.NewBatchProcessor(exp)
		logLP = sdklog.NewLoggerProvider(
			sdklog.WithResource(otelRes),
			sdklog.WithProcessor(processor),
		)
		otelCore := otelzap.NewCore(
			"github.com/rofleksey/dredge",
			otelzap.WithLoggerProvider(logLP),
		)
		core = zapcore.NewTee(consoleCore, otelCore)
	default:
		return nil, fmt.Errorf("unknown log_exporter: %q", logExporter)
	}

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	var tp *sdktrace.TracerProvider

	switch traceExporter {
	case "none":
		tp = sdktrace.NewTracerProvider()
	case "stdout":
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("stdout trace exporter: %w", err)
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(otelRes),
		)
	default:
		return nil, fmt.Errorf("unknown trace_exporter: %q", traceExporter)
	}

	otel.SetTracerProvider(tp)

	tracer := otel.Tracer(cfg.Observability.ServiceName)

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dredge_http_requests_total",
			Help: "Total number of incoming HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dredge_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	prometheus.MustRegister(requestsTotal, requestLatency)

	if cfg.Observability.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Observability.SentryDSN,
			EnableTracing:    true,
			TracesSampleRate: 0.2,
		}); err != nil {
			return nil, fmt.Errorf("init sentry: %w", err)
		}
	}

	return &Stack{
		Logger:            logger,
		LogLoggerProvider: logLP,
		TracerProvider:    tp,
		Tracer:            tracer,
		requestsTotal:     requestsTotal,
		requestLatency:    requestLatency,
	}, nil
}
