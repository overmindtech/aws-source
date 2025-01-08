package adapters

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

func TestVpcLinkOutputMapper(t *testing.T) {
	awsItem := &types.VpcLink{
		Id:          aws.String("vpc-link-id"),
		Name:        aws.String("vpc-link-name"),
		Description: aws.String("vpc-link-description"),
		TargetArns:  []string{"arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"},
		Status:      types.VpcLinkStatusAvailable,
		Tags:        map[string]string{"key": "value"},
	}

	item, err := vpcLinkOutputMapper("scope", awsItem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapterhelpers.QueryTests{
		{
			ExpectedType:   "elbv2-load-balancer",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188",
			ExpectedScope:  "scope",
		},
	}

	tests.Execute(t, item)
}

func TestNewAPIGatewayVpcLinkAdapter(t *testing.T) {
	config, account, region := adapterhelpers.GetAutoConfig(t)

	client := apigateway.NewFromConfig(config)

	adapter := NewAPIGatewayVpcLinkAdapter(client, account, region)

	test := adapterhelpers.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
