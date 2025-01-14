package adapters

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/overmindtech/aws-source/adapterhelpers"
)

func TestDeploymentOutputMapper(t *testing.T) {
	awsItem := &types.Deployment{
		Id:          aws.String("deployment-id"),
		CreatedDate: aws.Time(time.Now()),
		Description: aws.String("deployment-description"),
		ApiSummary:  map[string]map[string]types.MethodSnapshot{},
	}

	item, err := deploymentOutputMapper("scope", awsItem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewAPIGatewayDeploymentAdapter(t *testing.T) {
	config, account, region := adapterhelpers.GetAutoConfig(t)

	client := apigateway.NewFromConfig(config)

	adapter := NewAPIGatewayDeploymentAdapter(client, account, region)

	test := adapterhelpers.E2ETest{
		Adapter:  adapter,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}
