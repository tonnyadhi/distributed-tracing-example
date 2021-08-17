package tracing

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

func InitTracer(ctx context.Context, kind string, serviceName string, endpoint string) (func(), error) {
	var exporter sdktrace.SpanExporter

	log.Printf("Endpoint %s", endpoint)

	traceLabels := []attribute.KeyValue{
		attribute.String("service.name", serviceName),
		attribute.String("Host", os.Getenv("HOSTNAME")),
	}
	processDetector, err := resource.New(ctx, resource.WithOSType(), resource.WithProcess(), resource.WithProcessExecutableName())
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(kind, "stdouttrace") {
		exporterStdout, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		exporter = sdktrace.SpanExporter(exporterStdout)
	} else if strings.EqualFold(kind, "jaeger") {
		exporterJaeger, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
		if err != nil {
			return nil, err
		}
		exporter = sdktrace.SpanExporter(exporterJaeger)
	} else if strings.EqualFold(kind, "oteltrace") {
		exporterOtel, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		)
		if err != nil {
			return nil, err
		}
		exporter = sdktrace.SpanExporter(exporterOtel)
	} else {
		return nil, errors.New("unrecognized tracer kind")
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(processDetector),
		sdktrace.WithResource(resource.NewSchemaless(traceLabels...)),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown provider")
		handleErr(exporter.Shutdown(ctx), "failed to stop exporter")
	}, nil
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
