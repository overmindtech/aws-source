package cmd

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/detectors/aws/ec2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// for stdout debugging of traces
// import "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

// use this as template, or for tracing in the cmd package (currently not used)
// import "go.opentelemetry.io/otel/trace"
// const (
// 	instrumentationName    = "github.com/overmindtech/gateway/cmd"
// 	instrumentationVersion = "0.0.1"
// )
// var (
// 	tracer = otel.GetTracerProvider().Tracer(
// 		instrumentationName,
// 		trace.WithInstrumentationVersion(instrumentationVersion),
// 		trace.WithSchemaURL(semconv.SchemaURL),
// 	)
// )

func tracingResource() *resource.Resource {
	res, err := resource.New(context.Background(),
		resource.WithDetectors(ec2.NewResourceDetector()),
		// Keep the default detectors
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithTelemetrySDK(),
		resource.WithSchemaURL(semconv.SchemaURL),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String("aws-source"),
			semconv.ServiceVersionKey.String("0.0.1"),
		),
	)
	if err != nil {
		log.Fatalf("resource.New: %v", err)
		return nil
	}
	return res
}

var tp *sdktrace.TracerProvider

func initTracing(opts ...otlptracehttp.Option) error {
	// for stdout debugging of traces
	// stdoutExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// if err != nil {
	// 	return err
	// }

	client := otlptracehttp.NewClient(opts...)
	otlpExp, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	log.Infof("otlptracehttp client configured itself: %v", client)
	tp = sdktrace.NewTracerProvider(
		// sdktrace.WithSampler(sdktrace.AlwaysSample()),
		// for stdout debugging of traces
		// sdktrace.WithBatcher(stdoutExp),
		sdktrace.WithBatcher(otlpExp),
		sdktrace.WithResource(tracingResource()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return nil
}

func shutdownTracing() {
	if err := tp.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}
