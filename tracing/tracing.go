package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "embed"

	"github.com/MrAlias/otel-schema-utils/schema"
	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/detectors/aws/ec2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

//go:generate sh -c "echo -n $(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD) > commit.txt"
//go:embed commit.txt
var ServiceVersion string

func tracingResource() *resource.Resource {
	// Identify your application using resource detection
	resources := []*resource.Resource{}

	// the EC2 detector takes ~10s to time out outside EC2
	// disable it if we're running from a git checkout
	_, err := os.Stat(".git")
	if os.IsNotExist(err) {
		ec2Res, err := resource.New(context.Background(), resource.WithDetectors(ec2.NewResourceDetector()))
		if err != nil {
			log.WithError(err).Error("error initialising EC2 resource detector")
			return nil
		}
		resources = append(resources, ec2Res)
	}

	// Needs https://github.com/open-telemetry/opentelemetry-go-contrib/issues/1856 fixed first
	// // the EKS detector is temperamental and doesn't like running outside of kube
	// // hence we need to keep it from running when we know there's no kube
	// if !viper.GetBool("disable-kube") {
	// 	// Use the AWS resource detector to detect information about the runtime environment
	// 	detectors = append(detectors, eks.NewResourceDetector())
	// }

	hostRes, err := resource.New(context.Background(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		log.WithError(err).Error("error initialising host resource")
		return nil
	}
	resources = append(resources, hostRes)

	localRes, err := resource.New(context.Background(),
		resource.WithSchemaURL(semconv.SchemaURL),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String("aws-source"),
			semconv.ServiceVersionKey.String(ServiceVersion),
		),
	)
	if err != nil {
		log.WithError(err).Error("error initialising local resource")
		return nil
	}
	resources = append(resources, localRes)

	conv := schema.NewConverter(schema.DefaultClient)
	res, err := conv.MergeResources(context.Background(), semconv.SchemaURL, resources...)

	if err != nil {
		log.WithError(err).Error("error merging resource")
		return nil
	}
	return res
}

var tp *sdktrace.TracerProvider
var healthTp *sdktrace.TracerProvider

func GetHealthCheckTracerProvider() *sdktrace.TracerProvider {
	if healthTp == nil {
		panic("healthTp not initialised")
	}
	return healthTp
}

func InitTracing(opts ...otlptracehttp.Option) error {
	if sentry_dsn := viper.GetString("sentry-dsn"); sentry_dsn != "" {
		var environment string
		if viper.GetString("run-mode") == "release" {
			environment = "prod"
		} else {
			environment = "dev"
		}
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentry_dsn,
			AttachStacktrace: true,
			EnableTracing:    false,
			Environment:      environment,
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for performance monitoring.
			// We recommend adjusting this value in production,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Errorf("sentry.Init: %s", err)
		}
		// setup recovery for an unexpected panic in this function
		defer sentry.Flush(2 * time.Second)
		defer sentry.Recover()
		log.Info("sentry configured")
	}

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
		// for stdout debugging of traces
		// sdktrace.WithBatcher(stdoutExp),
		sdktrace.WithBatcher(otlpExp),
		sdktrace.WithResource(tracingResource()),
	)
	healthTp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased((0.1)))),
		// for stdout debugging of traces
		// sdktrace.WithBatcher(stdoutExp),
		sdktrace.WithBatcher(otlpExp),
		sdktrace.WithResource(tracingResource()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return nil
}

func ShutdownTracing() {
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	if err := tp.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
	}
}
