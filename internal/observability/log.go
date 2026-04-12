package observability

import (
	"context"
	"errors"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/rofleksey/dredge/internal/entity"
)

func (s *Stack) LogError(ctx context.Context, span trace.Span, msg string, err error, fields ...zap.Field) {
	if err == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	if !errors.Is(err, entity.ErrNoSentry) {
		sentry.CaptureException(err)
	}

	s.Logger.Error(msg, append(fields, zap.Error(err))...)
}
