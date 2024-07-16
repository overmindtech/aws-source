package networkmanager

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	awsnetworkmanager "github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources/integration"
)

var testAWSConfig *integration.AWSCfg
var testClient *awsnetworkmanager.Client

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
	t.Run("Setup", Setup)
	t.Run("NetworkManager", NetworkManager)
	t.Run("Teardown", Teardown)
}

func Setup(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	var err error
	testAWSConfig, err = integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS settings: %v", err)
	}
	testClient = awsnetworkmanager.NewFromConfig(testAWSConfig.Config)

	if err := setup(ctx, logger, testClient); err != nil {
		t.Fatalf("Failed to setup NetworkManager integration tests: %v", err)
	}
}

func Teardown(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	if err := teardown(ctx, logger, testClient); err != nil {
		t.Fatalf("Failed to teardown NetworkManager integration tests: %v", err)
	}
}
