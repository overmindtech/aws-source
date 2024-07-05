package networkmanager

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

func TestIntegrationNetworkManager(t *testing.T) {
	TestSetup(t)

	TestNetworkManager(t)

	TestTeardown(t)
}

func TestSetup(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	networkmanagerClient, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create NetworkManager client: %v", err)
	}

	if err := setup(ctx, logger, networkmanagerClient); err != nil {
		t.Fatalf("Failed to setup NetworkManager integration tests: %v", err)
	}
}

func TestTeardown(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	networkmanagerClient, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create NetworkManager client: %v", err)
	}

	if err := teardown(ctx, logger, networkmanagerClient); err != nil {
		t.Fatalf("Failed to teardown NetworkManager integration tests: %v", err)
	}
}
