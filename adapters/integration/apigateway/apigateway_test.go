package apigateway

import (
	"context"
	"fmt"
	"testing"

	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/aws-source/adapters/integration"
	"github.com/overmindtech/sdp-go"
)

func APIGateway(t *testing.T) {
	ctx := context.Background()

	var err error
	testClient, err := apigatewayClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create APIGateway client: %v", err)
	}

	testAWSConfig, err := integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("Failed to get AWS settings: %v", err)
	}

	accountID := testAWSConfig.AccountID

	t.Log("Running APIGateway integration test")

	restApiSource := adapters.NewAPIGatewayRestApiAdapter(testClient, accountID, testAWSConfig.Region)

	err = restApiSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate APIGateway restApi adapter: %v", err)
	}

	resourceApiSource := adapters.NewAPIGatewayResourceAdapter(testClient, accountID, testAWSConfig.Region)

	err = resourceApiSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate APIGateway resource adapter: %v", err)
	}

	methodSource := adapters.NewAPIGatewayMethodAdapter(testClient, accountID, testAWSConfig.Region)

	err = methodSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate APIGateway method adapter: %v", err)
	}

	methodResponseSource := adapters.NewAPIGatewayMethodResponseAdapter(testClient, accountID, testAWSConfig.Region)

	err = methodResponseSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate APIGateway method response adapter: %v", err)
	}

	integrationSource := adapters.NewAPIGatewayIntegrationAdapter(testClient, accountID, testAWSConfig.Region)

	err = integrationSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate APIGateway integration adapter: %v", err)
	}

	apiKeySource := adapters.NewAPIGatewayApiKeyAdapter(testClient, accountID, testAWSConfig.Region)

	err = apiKeySource.Validate()
	if err != nil {
		t.Fatalf("failed to validate APIGateway API key adapter: %v", err)
	}

	scope := adapterhelpers.FormatScope(accountID, testAWSConfig.Region)

	// List restApis
	restApis, err := restApiSource.List(ctx, scope, true)
	if err != nil {
		t.Fatalf("failed to list APIGateway restApis: %v", err)
	}

	if len(restApis) == 0 {
		t.Fatalf("no restApis found")
	}

	restApiUniqueAttribute := restApis[0].GetUniqueAttribute()

	restApiID, err := integration.GetUniqueAttributeValueByTags(
		restApiUniqueAttribute,
		restApis,
		integration.ResourceTags(integration.APIGateway, restAPISrc),
		true,
	)
	if err != nil {
		t.Fatalf("failed to get restApi ID: %v", err)
	}

	// Get restApi
	restApi, err := restApiSource.Get(ctx, scope, restApiID, true)
	if err != nil {
		t.Fatalf("failed to get APIGateway restApi: %v", err)
	}

	restApiIDFromGet, err := integration.GetUniqueAttributeValueByTags(
		restApiUniqueAttribute,
		[]*sdp.Item{restApi},
		integration.ResourceTags(integration.APIGateway, restAPISrc),
		true,
	)
	if err != nil {
		t.Fatalf("failed to get restApi ID from get: %v", err)
	}

	if restApiID != restApiIDFromGet {
		t.Fatalf("expected restApi ID %s, got %s", restApiID, restApiIDFromGet)
	}

	// Search restApis
	restApiName := integration.ResourceName(integration.APIGateway, restAPISrc, integration.TestID())
	restApisFromSearch, err := restApiSource.Search(ctx, scope, restApiName, true)
	if err != nil {
		t.Fatalf("failed to search APIGateway restApis: %v", err)
	}

	if len(restApis) == 0 {
		t.Fatalf("no restApis found")
	}

	restApiIDFromSearch, err := integration.GetUniqueAttributeValueBySignificantAttribute(
		restApiUniqueAttribute,
		"Name",
		integration.ResourceName(integration.APIGateway, restAPISrc, integration.TestID()),
		restApisFromSearch,
		true,
	)
	if err != nil {
		t.Fatalf("failed to get restApi ID from search: %v", err)
	}

	if restApiID != restApiIDFromSearch {
		t.Fatalf("expected restApi ID %s, got %s", restApiID, restApiIDFromSearch)
	}

	// Search resources
	resources, err := resourceApiSource.Search(ctx, scope, restApiID, true)
	if err != nil {
		t.Fatalf("failed to search APIGateway resources: %v", err)
	}

	if len(resources) == 0 {
		t.Fatalf("no resources found")
	}

	resourceUniqueAttribute := resources[0].GetUniqueAttribute()

	resourceUniqueAttrFromSearch, err := integration.GetUniqueAttributeValueBySignificantAttribute(
		resourceUniqueAttribute,
		"Path",
		"/test",
		resources,
		true,
	)
	if err != nil {
		t.Fatalf("failed to get resource ID: %v", err)
	}

	// Get resource
	resource, err := resourceApiSource.Get(ctx, scope, resourceUniqueAttrFromSearch, true)
	if err != nil {
		t.Fatalf("failed to get APIGateway resource: %v", err)
	}

	resourceUniqueAttrFromGet, err := integration.GetUniqueAttributeValueBySignificantAttribute(
		resourceUniqueAttribute,
		"Path",
		"/test",
		[]*sdp.Item{resource},
		true,
	)
	if err != nil {
		t.Fatalf("failed to get resource ID from get: %v", err)
	}

	if resourceUniqueAttrFromSearch != resourceUniqueAttrFromGet {
		t.Fatalf("expected resource ID %s, got %s", resourceUniqueAttrFromSearch, resourceUniqueAttrFromGet)
	}

	// Get method
	methodID := fmt.Sprintf("%s/GET", resourceUniqueAttrFromGet) // resourceUniqueAttribute contains the restApiID
	method, err := methodSource.Get(ctx, scope, methodID, true)
	if err != nil {
		t.Fatalf("failed to get APIGateway method: %v", err)
	}

	uniqueMethodAttr, err := method.GetAttributes().Get(method.GetUniqueAttribute())
	if err != nil {
		t.Fatalf("failed to get unique method attribute: %v", err)
	}

	if uniqueMethodAttr != methodID {
		t.Fatalf("expected method ID %s, got %s", methodID, uniqueMethodAttr)
	}

	// Get method response
	methodResponseID := fmt.Sprintf("%s/200", methodID)
	methodResponse, err := methodResponseSource.Get(ctx, scope, methodResponseID, true)
	if err != nil {
		t.Fatalf("failed to get APIGateway method response: %v", err)
	}

	uniqueMethodResponseAttr, err := methodResponse.GetAttributes().Get(methodResponse.GetUniqueAttribute())
	if err != nil {
		t.Fatalf("failed to get unique method response attribute: %v", err)
	}

	if uniqueMethodResponseAttr != methodResponseID {
		t.Fatalf("expected method response ID %s, got %s", methodResponseID, uniqueMethodResponseAttr)
	}

	// Get integration
	integrationID := fmt.Sprintf("%s/GET", resourceUniqueAttrFromGet) // resourceUniqueAttribute contains the restApiID
	itgr, err := integrationSource.Get(ctx, scope, integrationID, true)
	if err != nil {
		t.Fatalf("failed to get APIGateway itgr: %v", err)
	}

	uniqueIntegrationAttr, err := itgr.GetAttributes().Get(itgr.GetUniqueAttribute())
	if err != nil {
		t.Fatalf("failed to get unique itgr attribute: %v", err)
	}

	if uniqueIntegrationAttr != integrationID {
		t.Fatalf("expected integration ID %s, got %s", integrationID, uniqueIntegrationAttr)
	}

	// List API keys
	apiKeys, err := apiKeySource.List(ctx, scope, true)
	if err != nil {
		t.Fatalf("failed to list APIGateway API keys: %v", err)
	}

	if len(apiKeys) == 0 {
		t.Fatalf("no API keys found")
	}

	apiKeyUniqueAttribute := apiKeys[0].GetUniqueAttribute()

	apiKeyID, err := integration.GetUniqueAttributeValueByTags(
		apiKeyUniqueAttribute,
		apiKeys,
		integration.ResourceTags(integration.APIGateway, apiKeySrc),
		true,
	)
	if err != nil {
		t.Fatalf("failed to get API key ID: %v", err)
	}

	// Get API key
	apiKey, err := apiKeySource.Get(ctx, scope, apiKeyID, true)
	if err != nil {
		t.Fatalf("failed to get APIGateway API key: %v", err)
	}

	apiKeyIDFromGet, err := integration.GetUniqueAttributeValueByTags(
		apiKeyUniqueAttribute,
		[]*sdp.Item{apiKey},
		integration.ResourceTags(integration.APIGateway, apiKeySrc),
		true,
	)
	if err != nil {
		t.Fatalf("failed to get API key ID from get: %v", err)
	}

	if apiKeyID != apiKeyIDFromGet {
		t.Fatalf("expected API key ID %s, got %s", apiKeyID, apiKeyIDFromGet)
	}

	// Search API keys
	apiKeyName := integration.ResourceName(integration.APIGateway, apiKeySrc, integration.TestID())
	apiKeysFromSearch, err := apiKeySource.Search(ctx, scope, apiKeyName, true)
	if err != nil {
		t.Fatalf("failed to search APIGateway API keys: %v", err)
	}

	if len(apiKeysFromSearch) == 0 {
		t.Fatalf("no API keys found")
	}

	apiKeyIDFromSearch, err := integration.GetUniqueAttributeValueBySignificantAttribute(
		apiKeyUniqueAttribute,
		"Name",
		apiKeyName,
		apiKeysFromSearch,
		true,
	)
	if err != nil {
		t.Fatalf("failed to get API key ID from search: %v", err)
	}

	if apiKeyID != apiKeyIDFromSearch {
		t.Fatalf("expected API key ID %s, got %s", apiKeyID, apiKeyIDFromSearch)
	}

	t.Log("APIGateway integration test completed")
}
