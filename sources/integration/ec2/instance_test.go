package ec2

import (
	"context"
	"fmt"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/ec2"
	"github.com/overmindtech/aws-source/sources/integration"
	"github.com/overmindtech/sdp-go"
)

func TestEC2(t *testing.T) {
	ctx := context.Background()

	t.Log("Running EC2 integration test")

	ec2Cli, err := createEC2Client(ctx)
	if err != nil {
		t.Fatalf("failed to create EC2 client: %v", err)
	}

	awsCfg, err := integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("failed to get AWS settings: %v", err)
	}

	instanceSource := ec2.NewInstanceSource(ec2Cli, awsCfg.AccountID, awsCfg.Region)

	err = instanceSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate EC2 instance source: %v", err)
	}

	scope := sources.FormatScope(awsCfg.AccountID, awsCfg.Region)

	// List instances
	sdpListInstances, err := instanceSource.List(context.Background(), scope, true)
	if err != nil {
		t.Fatalf("failed to list EC2 instances: %v", err)
	}

	if len(sdpListInstances) == 0 {
		t.Fatalf("no instances found")
	}

	uniqueAttribute := sdpListInstances[0].GetUniqueAttribute()

	instanceID, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		sdpListInstances,
		integration.ResourceTags(integration.EC2, instanceSrc),
	)
	if err != nil {
		t.Fatalf("failed to get instance ID: %v", err)
	}

	// Get instance
	sdpInstance, err := instanceSource.Get(context.Background(), scope, instanceID, true)
	if err != nil {
		t.Fatalf("failed to get EC2 instance: %v", err)
	}

	instanceIDFromGet, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		[]*sdp.Item{sdpInstance},
		integration.ResourceTags(integration.EC2, instanceSrc),
	)
	if err != nil {
		t.Fatalf("failed to get instance ID from get: %v", err)
	}

	if instanceIDFromGet != instanceID {
		t.Fatalf("expected instance ID %v, got %v", instanceID, instanceIDFromGet)
	}

	// Search instances
	instanceARN := fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsCfg.Region, awsCfg.AccountID, instanceID)
	sdpSearchInstances, err := instanceSource.Search(context.Background(), scope, instanceARN, true)
	if err != nil {
		t.Fatalf("failed to search EC2 instances: %v", err)
	}

	if len(sdpSearchInstances) == 0 {
		t.Fatalf("no instances found")
	}

	instanceIDFromSearch, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		sdpSearchInstances,
		integration.ResourceTags(integration.EC2, instanceSrc),
	)
	if err != nil {
		t.Fatalf("failed to get instance ID from search: %v", err)
	}

	if instanceIDFromSearch != instanceID {
		t.Fatalf("expected instance ID %v, got %v", instanceID, instanceIDFromSearch)
	}
}
