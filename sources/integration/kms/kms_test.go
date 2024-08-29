package kms

import (
	"context"
	"fmt"
	"strings"
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

	aliasSource := kms.NewAliasSource(testClient, accountID, testAWSConfig.Region)

	grantSource := kms.NewGrantSource(testClient, accountID, testAWSConfig.Region)

	keyPolicySource := kms.NewKeyPolicySource(testClient, accountID, testAWSConfig.Region)

	err = keySource.Validate()
	if err != nil {
		t.Fatalf("failed to validate KMS key source: %v", err)
	}

	err = aliasSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate KMS alias source: %v", err)
	}

	err = grantSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate KMS grant source: %v", err)
	}

	err = keyPolicySource.Validate()
	if err != nil {
		t.Fatalf("failed to validate KMS key policy source: %v", err)
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

	keyUniqueAttribute := sdpListKeys[0].GetUniqueAttribute()

	keyID, err := integration.GetUniqueAttributeValue(
		keyUniqueAttribute,
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
		keyUniqueAttribute,
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
		keyUniqueAttribute,
		sdpSearchKeys,
		integration.ResourceTags(integration.KMS, keySrc),
	)
	if err != nil {
		t.Fatalf("failed to get key ID from search: %v", err)
	}

	if keyIDFromSearch != keyID {
		t.Fatalf("expected key ID %v, got %v", keyID, keyIDFromSearch)
	}

	// List aliases
	sdpListAliases, err := aliasSource.List(context.Background(), scope, true)
	if err != nil {
		t.Fatalf("failed to list KMS aliases: %v", err)
	}

	if len(sdpListAliases) == 0 {
		t.Fatalf("no aliases found")
	}

	// Get the alias for this key
	var aliasUniqueAttributeValue interface{}

	for _, alias := range sdpListAliases {
		// Check if the alias is for the key
		for _, query := range alias.GetLinkedItemQueries() {
			if query.GetQuery().GetQuery() == keyID {
				aliasUniqueAttributeValue, err = alias.GetAttributes().Get(alias.GetUniqueAttribute())
				if err != nil {
					t.Fatalf("failed to get alias unique attribute values: %v", err)
				}
				break
			}
		}
	}

	if aliasUniqueAttributeValue == nil {
		t.Fatalf("no alias found for key %v", keyID)
	}

	sdpAlias, err := aliasSource.Get(context.Background(), scope, aliasUniqueAttributeValue.(string), true)
	if err != nil {
		t.Fatalf("failed to get KMS alias: %v", err)
	}

	aliasName, err := sdpAlias.GetAttributes().Get("aliasName")
	if err != nil {
		t.Fatalf("failed to get alias name: %v", err)
	}

	if aliasName != genAliasName() {
		t.Fatalf("expected alias %v, got %v", genAliasName(), aliasName)
	}

	// Search aliases
	sdpSearchAliases, err := aliasSource.Search(context.Background(), scope, keyID, true)
	if err != nil {
		t.Fatalf("failed to search KMS aliases: %v", err)
	}

	if len(sdpSearchAliases) == 0 {
		t.Fatalf("no aliases found")
	}

	searchAliasName, err := sdpSearchAliases[0].GetAttributes().Get("aliasName")
	if err != nil {
		t.Fatalf("failed to get alias name: %v", err)
	}

	if searchAliasName != genAliasName() {
		t.Fatalf("expected alias %v, got %v", genAliasName(), searchAliasName)
	}

	// List grants is not supported
	sdpListGrants, err := grantSource.List(context.Background(), scope, true)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if len(sdpListGrants) != 0 {
		t.Fatalf("expected 0 grants, got %v", len(sdpListGrants))
	}

	// Search grants
	sdpSearchGrants, err := grantSource.Search(context.Background(), scope, keyID, true)
	if err != nil {
		t.Fatalf("failed to search KMS grants: %v", err)
	}

	if len(sdpSearchGrants) == 0 {
		t.Fatal("no grants found")
	}
	searchGrantID, err := sdpSearchGrants[0].GetAttributes().Get("grantId")
	if err != nil {
		t.Fatalf("failed to get grant ID: %v", err)
	}

	// Get grant
	grantUniqueAttribute := sdpSearchGrants[0].GetUniqueAttribute()
	grantUniqueAttributeValue, err := sdpSearchGrants[0].GetAttributes().Get(grantUniqueAttribute)
	if err != nil {
		t.Fatalf("failed to get grant unique attribute values: %v", err)
	}

	sdpGrant, err := grantSource.Get(context.Background(), scope, grantUniqueAttributeValue.(string), true)
	if err != nil {
		t.Fatalf("failed to get KMS grant: %v", err)
	}

	grantID, err := sdpGrant.GetAttributes().Get("grantId")
	if err != nil {
		t.Fatalf("failed to get grant ID: %v", err)
	}

	expectedGrantID := strings.Split(grantUniqueAttributeValue.(string), "/")[1]

	if grantID != expectedGrantID {
		t.Fatalf("expected grant ID %v, got %v", expectedGrantID, grantID)
	}

	if searchGrantID != expectedGrantID {
		t.Fatalf("expected grant ID %v, got %v", expectedGrantID, searchGrantID)
	}

	// Search key policy by key ID
	sdpSearchKeyPolicies, err := keyPolicySource.Search(context.Background(), scope, keyID, true)
	if err != nil {
		t.Fatalf("failed to search KMS key policies: %v", err)
	}

	if len(sdpSearchKeyPolicies) == 0 {
		t.Fatalf("no key policies found")
	}

	searchKeyPolicyKeyID, err := sdpSearchKeyPolicies[0].GetAttributes().Get("keyId")
	if err != nil {
		t.Fatalf("failed to get key ID: %v", err)
	}

	if searchKeyPolicyKeyID != keyID {
		t.Fatalf("expected key ID %v, got %v", keyID, searchKeyPolicyKeyID)
	}

	// Get key policy
	keyPolicyUniqueAttribute := sdpSearchKeyPolicies[0].GetUniqueAttribute()
	keyPolicyUniqueAttributeValue, err := sdpSearchKeyPolicies[0].GetAttributes().Get(keyPolicyUniqueAttribute)
	if err != nil {
		t.Fatalf("failed to get key policy unique attribute values: %v", err)
	}

	sdpKeyPolicy, err := keyPolicySource.Get(context.Background(), scope, keyPolicyUniqueAttributeValue.(string), true)
	if err != nil {
		t.Fatalf("failed to get KMS key policy: %v", err)
	}

	keyPolicyKeyID, err := sdpKeyPolicy.GetAttributes().Get("keyId")
	if err != nil {
		t.Fatalf("failed to get key ID: %v", err)
	}

	if keyPolicyKeyID != keyID {
		t.Fatalf("expected key ID %v, got %v", keyID, keyPolicyKeyID)
	}
}
