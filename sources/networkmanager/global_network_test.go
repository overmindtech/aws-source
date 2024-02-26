package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
	"testing"
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

func TestLoadBalancerOutputMapper(t *testing.T) {
	output := networkmanager.DescribeGlobalNetworksOutput{
		GlobalNetworks: []types.GlobalNetwork{
			{
				GlobalNetworkArn: sources.PtrString("arn:aws:networkmanager:eu-west-2:052392120703:networkmanager/global-network/default"),
				GlobalNetworkId:  sources.PtrString("default"),
			},
		},
	}

	items, err := globalNetworkOutputMapper(context.Background(), &TestClient{}, "foo", nil, &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-site",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "default",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}
