package otel

import (
	"context"

	"github.com/ovinc/zerotrust/internal/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/ovinc/zerotrust"

var tracer trace.Tracer
var tracerProvider *sdkTrace.TracerProvider

func init() {
	// init
	ctx := context.Background()

	// init tracer
	tracer = otel.Tracer(instrumentationName)

	// load config
	cfg := config.Get().OTel

	// return empty instance if otel is disabled
	if !cfg.Enabled {
		return
	}

	// build resource with configured attributes
	res, err := buildResource(ctx, &cfg.Resource)
	if err != nil {
		logrus.Fatalf("failed to create resource: %s", err)
	}

	// create trace exporter using grpc
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(cfg.Endpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("failed to create trace exporter: %s", err)
	}

	// create and set tracer provider
	tracerProvider = sdkTrace.NewTracerProvider(sdkTrace.WithBatcher(traceExporter), sdkTrace.WithResource(res))
	otel.SetTracerProvider(tracerProvider)

	// set text map propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func buildResource(ctx context.Context, cfg *config.ResourceConfig) (*resource.Resource, error) {
	// build base attributes from config
	attrs := []resource.Option{resource.WithAttributes(semconv.ServiceName(cfg.ServiceName))}

	// add custom attributes from config
	for key, value := range cfg.Attributes {
		attrs = append(attrs, resource.WithAttributes(attribute.String(key, value)))
	}

	return resource.New(ctx, attrs...)
}

func Shutdown(ctx context.Context) {
	if tracerProvider != nil {
		_ = tracerProvider.Shutdown(ctx)
	}
}
