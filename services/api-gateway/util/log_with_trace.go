package util

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func LoggerWithTrace(ctx context.Context, logger *zap.SugaredLogger) *zap.SugaredLogger {
	span := trace.SpanFromContext(ctx)

	if !span.SpanContext().IsValid() {
		return logger
	}

	sc := span.SpanContext()

	return logger.With(
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	)
}
