package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/overmindtech/sdp-go"
)

const (
	TagTestKey       = "test"
	TagTestValue     = "true"
	TagTestIDKey     = "test-id"
	TagTestTypeKey   = "test-type"
	TagResourceIDKey = "resource-id"
)

type resourceGroup int

const (
	NetworkManager resourceGroup = iota
	EC2
	KMS
)

func (rg resourceGroup) String() string {
	switch rg {
	case NetworkManager:
		return "network-manager"
	case EC2:
		return "ec2"
	case KMS:
		return "kms"
	default:
		return "unknown"
	}
}

func ShouldRunIntegrationTests() bool {
	run, found := os.LookupEnv("RUN_INTEGRATION_TESTS")

	if !found {
		return false
	}

	shouldRun, err := strconv.ParseBool(run)
	if err != nil {
		return false
	}

	return shouldRun
}

func TestID() string {
	tagTestID, found := os.LookupEnv("INTEGRATION_TEST_ID")
	if !found {
		var err error
		tagTestID, err = os.Hostname()
		if err != nil {
			panic("failed to get hostname")
		}
	}

	return tagTestID
}

func TestName(resourceGroup resourceGroup) string {
	return fmt.Sprintf("%s-integration-tests", resourceGroup.String())
}

type AWSCfg struct {
	AccountID string
	Region    string
	Config    aws.Config
}

func AWSSettings(ctx context.Context) (*AWSCfg, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	callerIdentity, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	accountID := aws.ToString(callerIdentity.Account)

	return &AWSCfg{
		AccountID: accountID,
		Region:    cfg.Region,
		Config:    cfg,
	}, nil
}

func removeUnhealthy(sdpInstances []*sdp.Item) []*sdp.Item {
	var filteredInstances []*sdp.Item
	for _, instance := range sdpInstances {
		if instance.GetHealth() != sdp.Health_HEALTH_OK {
			continue
		}
		filteredInstances = append(filteredInstances, instance)
	}
	return filteredInstances
}

func GetUniqueAttributeValue(uniqueAttrKey string, items []*sdp.Item, filterTags map[string]string) (string, error) {
	var filteredItems []*sdp.Item
	for _, item := range removeUnhealthy(items) {
		if hasTags(item.GetTags(), filterTags) {
			filteredItems = append(filteredItems, item)
		}
	}

	if len(filteredItems) != 1 {
		return "", fmt.Errorf("expected 1 item, got %v", len(filteredItems))
	}

	uniqueAttrValue, err := filteredItems[0].GetAttributes().Get(uniqueAttrKey)
	if err != nil {
		return "", fmt.Errorf("failed to get %s: %w", uniqueAttrKey, err)
	}

	uniqueAttrValueStr := uniqueAttrValue.(string)
	if uniqueAttrValueStr == "" {
		return "", fmt.Errorf("%s is empty", uniqueAttrKey)
	}

	return uniqueAttrValueStr, nil
}

// ResourceName returns a unique resource name for integration tests
// I.e., integration-test-networkmanager-global-network-1
func ResourceName(resourceGroup resourceGroup, resourceName string, additionalAttr ...string) string {
	name := []string{"integration-test", resourceGroup.String(), resourceName}

	name = append(name, additionalAttr...)

	return strings.Join(name, "-")
}

func ResourceTags(resourceGroup resourceGroup, resourceName string, additionalAttr ...string) map[string]string {
	return map[string]string{
		TagTestKey:       TagTestValue,
		TagTestTypeKey:   TestName(resourceGroup),
		TagTestIDKey:     TestID(),
		TagResourceIDKey: ResourceName(resourceGroup, resourceName, additionalAttr...),
	}
}

func hasTags(tags map[string]string, requiredTags map[string]string) bool {
	for k, v := range requiredTags {
		if tags[k] != v {
			return false
		}
	}

	return true
}

func AssertEqualItems(t *testing.T, expected, actual []*sdp.Item, uniqueAttrKey string) {
	if len(expected) != len(actual) {
		t.Fatalf("expected %d items, got %d", len(expected), len(actual))
	}

	expectedUnqAttrValSet, err := uniqueAttributeValueSet(expected, uniqueAttrKey)
	if err != nil {
		t.Fatalf("failed to get unique attribute value set: %v", err)
	}

	actualUnqAttrValSet, err := uniqueAttributeValueSet(actual, uniqueAttrKey)
	if err != nil {
		t.Fatalf("failed to get unique attribute value set: %v", err)
	}

	if len(expectedUnqAttrValSet) != len(actualUnqAttrValSet) {
		t.Fatalf("expected %d unique values, got %d", len(expectedUnqAttrValSet), len(actualUnqAttrValSet))
	}

	for val := range expectedUnqAttrValSet {
		if _, ok := actualUnqAttrValSet[val]; !ok {
			t.Fatalf("expected value %v not found in actual", val)
		}
	}
}

func uniqueAttributeValueSet(items []*sdp.Item, key string) (map[any]bool, error) {
	uniqueValues := make(map[any]bool)
	for _, item := range items {
		value, err := item.GetAttributes().Get(key)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", key, err)
		}
		uniqueValues[value] = true
	}
	return uniqueValues, nil
}
