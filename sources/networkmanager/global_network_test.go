package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
	"testing"
	"time"
)

func (t *TestClient) DescribeGlobalNetworks(ctx context.Context, params *networkmanager.DescribeGlobalNetworksInput, optFns ...func(*networkmanager.Options)) (*networkmanager.DescribeGlobalNetworksOutput, error) {
	return &networkmanager.DescribeGlobalNetworksOutput{
		GlobalNetworks: []types.GlobalNetwork{
			{
				Tags:             []types.Tag{},
				GlobalNetworkArn: sources.PtrString("arn:aws:networkmanager:eu-west-2:052392120703:global-network/default"),
				GlobalNetworkId:  sources.PtrString("default"),
			},
		},
	}, nil
}

func TestGlobalNetworkGetFunc(t *testing.T) {
	scope := "123456789012.eu-west-2"
	item, err := globalNetworkGetFunc(context.Background(), &TestClient{}, scope, &networkmanager.DescribeGlobalNetworksInput{})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-site",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "default",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}

func TestNewGlobalNetworkSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewGlobalNetworkSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
