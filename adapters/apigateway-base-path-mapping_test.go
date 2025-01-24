package adapters

import (
	"github.com/overmindtech/sdp-go"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/overmindtech/aws-source/adapterhelpers"
)

func TestBasePathMappingOutputMapper(t *testing.T) {
	awsItem := &types.BasePathMapping{
		BasePath:  aws.String("base-path"),
		RestApiId: aws.String("rest-api-id"),
		Stage:     aws.String("stage"),
	}

	item, err := basePathMappingOutputMapper("domain-name", "scope", awsItem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapterhelpers.QueryTests{
		{
			ExpectedType:   "apigateway-domain-name",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "domain-name",
			ExpectedScope:  "scope",
		},
		{
			ExpectedType:   "apigateway-rest-api",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "rest-api-id",
			ExpectedScope:  "scope",
		},
	}

	tests.Execute(t, item)
}

func TestNewAPIGatewayBasePathMappingAdapter(t *testing.T) {
	config, account, region := adapterhelpers.GetAutoConfig(t)

	client := apigateway.NewFromConfig(config)

	adapter := NewAPIGatewayBasePathMappingAdapter(client, account, region)

	test := adapterhelpers.E2ETest{
		Adapter:  adapter,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}
