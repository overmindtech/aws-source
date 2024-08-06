package kms

import (
	"context"
	"fmt"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/integration"
	"github.com/overmindtech/aws-source/sources/kms"
	"github.com/overmindtech/sdp-go"
)

func KMS(t *testing.T) {
	ctx := context.Background()

	var err error
	testClient, err := kmsClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create KMS client: %v", err)
	}

	testAWSConfig, err := integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS settings: %v", err)
	}

	accountID := testAWSConfig.AccountID

	t.Log("Running KMS integration test")

	keySource := kms.NewKeySource(testClient, accountID, testAWSConfig.Region)

	err = keySource.Validate()
	if err != nil {
		t.Fatalf("failed to validate KMS key source: %v", err)
	}

	scope := sources.FormatScope(accountID, testAWSConfig.Region)

	// List keys
	sdpListKeys, err := keySource.List(context.Background(), scope, true)
	if err != nil {
		t.Fatalf("failed to list KMS keys: %v", err)
	}

	if len(sdpListKeys) == 0 {
		t.Fatalf("no keys found")
	}

	uniqueAttribute := sdpListKeys[0].GetUniqueAttribute()

	keyID, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		sdpListKeys,
		integration.ResourceTags(integration.KMS, keySrc),
	)
	if err != nil {
		t.Fatalf("failed to get key ID: %v", err)
	}

	// Get key
	sdpKey, err := keySource.Get(context.Background(), scope, keyID, true)
	if err != nil {
		t.Fatalf("failed to get KMS key: %v", err)
	}

	keyIDFromGet, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		[]*sdp.Item{sdpKey},
		integration.ResourceTags(integration.KMS, keySrc),
	)
	if err != nil {
		t.Fatalf("failed to get key ID from get: %v", err)
	}

	if keyIDFromGet != keyID {
		t.Fatalf("expected key ID %v, got %v", keyID, keyIDFromGet)
	}

	// Search keys
	keyARN := fmt.Sprintf("arn:aws:kms:%s:%s:key/%s", testAWSConfig.Region, accountID, keyID)
	sdpSearchKeys, err := keySource.Search(context.Background(), scope, keyARN, true)
	if err != nil {
		t.Fatalf("failed to search KMS keys: %v", err)
	}

	if len(sdpSearchKeys) == 0 {
		t.Fatalf("no keys found")
	}

	keyIDFromSearch, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		sdpSearchKeys,
		integration.ResourceTags(integration.KMS, keySrc),
	)
	if err != nil {
		t.Fatalf("failed to get key ID from search: %v", err)
	}

	if keyIDFromSearch != keyID {
		t.Fatalf("expected key ID %v, got %v", keyID, keyIDFromSearch)
	}
}
