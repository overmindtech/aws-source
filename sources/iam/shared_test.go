package iam

import (
	"log"
	"os"
	"testing"

	"github.com/micahhausler/aws-iam-policy/policy"
	"github.com/overmindtech/aws-source/tracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// TestIAMClient Test client that returns three pages
type TestIAMClient struct{}

func TestMain(m *testing.M) {
	// Add tracing if present
	key, _ := os.LookupEnv("HONEYCOMB_API_KEY")
	opts := make([]otlptracehttp.Option, 0)
	if key != "" {
		opts = []otlptracehttp.Option{
			otlptracehttp.WithEndpoint("api.honeycomb.io"),
			otlptracehttp.WithHeaders(map[string]string{"x-honeycomb-team": key}),
		}
	}

	if err := tracing.InitTracing(opts...); err != nil {
		log.Fatal(err)
	}
	defer tracing.ShutdownTracing()

	os.Exit(m.Run())
}

func TestLinksFromPolicy(t *testing.T) {
	action := policy.NewStringOrSlice(true, "sts:AssumeRole")
	pol := policy.Policy{
		Statements: policy.NewSingularStatementOrSlice(policy.Statement{
			Action:    action,
			Effect:    "Allow",
			Principal: policy.NewAWSPrincipal("arn:aws:iam::123456789:role/aws-controltower-AuditAdministratorRole"),
		}),
	}

	queries := LinksFromPolicy(&pol)

	if len(queries) != 1 {
		t.Fatalf("expected 1 query got %v", len(queries))
	}

}
