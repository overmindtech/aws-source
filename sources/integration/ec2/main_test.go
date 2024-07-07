package ec2

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/overmindtech/aws-source/sources/integration"
)

func TestMain(m *testing.M) {
	if integration.ShouldRunIntegrationTests() {
		fmt.Println("Running integration tests")
		os.Exit(m.Run())
	} else {
		fmt.Println("Skipping integration tests, set RUN_INTEGRATION_TESTS=true to run them")
		os.Exit(0)
	}
}

func TestIntegrationEC2(t *testing.T) {
	t.Run("Setup", Setup)
	t.Run("EC2", EC2)
	t.Run("Teardown", Teardown)
}

func Setup(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	ec2Client, err := createEC2Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	if err := setup(ctx, logger, ec2Client); err != nil {
		t.Fatalf("Failed to setup EC2 integration tests: %v", err)
	}
}

func Teardown(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	ec2Client, err := createEC2Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)
	}

	if err := teardown(ctx, logger, ec2Client); err != nil {
		t.Fatalf("Failed to teardown EC2 integration tests: %v", err)
	}
}
