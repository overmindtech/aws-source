package ec2

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources/integration"
)

var testAWSConfig *integration.AWSCfg
var testClient *awsec2.Client

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
	var err error
	testAWSConfig, err = integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS settings: %v", err)
	}
	testClient = awsec2.NewFromConfig(testAWSConfig.Config)

	if err := setup(ctx, logger, testClient); err != nil {
		t.Fatalf("Failed to setup EC2 integration tests: %v", err)
	}
}

func Teardown(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	if err := teardown(ctx, logger, testClient); err != nil {
		t.Fatalf("Failed to teardown EC2 integration tests: %v", err)
	}
}
