package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	traceExporter sdktrace.SpanExporter
	traceProvider *sdktrace.TracerProvider
	propagator    propagation.TextMapPropagator
}

func NewTracer(jaegerEndpoint string, serviceName string, environment string) (*Tracer, error) {
	//create trace exporter
	traceExporter, err := newExporter(context.Background(), jaegerEndpoint)
	if err != nil {
		return nil, err
	}
	// create trace provider
	traceProvider, err := newTraceProvider(serviceName, environment, traceExporter)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(traceProvider)

	// create and set global propagator
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	return &Tracer{
		traceExporter: traceExporter,
		traceProvider: traceProvider,
		propagator:    prop,
	}, nil
}

func (t *Tracer) Shutdown(ctx context.Context) error {
	return t.traceProvider.Shutdown(ctx)
}

func GetTracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

func newExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	return otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(serviceName, environment string, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName), semconv.DeploymentEnvironmentName(environment)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	), nil
}
