package cmd

import (
	"github.com/overmindtech/aws-source/tracing"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName    = "github.com/overmindtech/aws-source/cmd"
	instrumentationVersion = "0.0.1"
)

// healthCheckTracer is the tracer used for health checks. This is heavily sampled to avoid getting spammed by k8s or ELBs
var healthCheckTracer = tracing.GetHealthCheckTracerProvider().Tracer(
	instrumentationName,
	trace.WithInstrumentationVersion(instrumentationVersion),
	trace.WithSchemaURL(semconv.SchemaURL),
	trace.WithInstrumentationAttributes(
		attribute.Bool("ovm.healthCheck", true),
	),
)
