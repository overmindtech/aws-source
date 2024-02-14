package networkmanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func globalNetworkGetFunc(ctx context.Context, client NetworkmanagerClient, scope string, input *networkmanager.DescribeGlobalNetworksInput) (*sdp.Item, error) {
	out, err := client.DescribeGlobalNetworks(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(out.GlobalNetworks) == 0 {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: fmt.Sprintf("global network with ID %s not found", input.GlobalNetworkIds[0]),
		}
	}
	if len(out.GlobalNetworks) != 1 {
		return nil, fmt.Errorf("got %d global networks, expected 1", len(out.GlobalNetworks))
	}

	gn := out.GlobalNetworks[0]

	attributes, err := sources.ToAttributesCase(gn, "tags")
	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "networkmanager-global-network",
		UniqueAttribute: "globalNetworkId",
		Scope:           scope,
		Attributes:      attributes,
		Tags:            tagsToMap(gn.Tags),
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
				Query: &sdp.Query{
					// +overmind:link networkmanager-site
					// Search for all sites with this global network
					Type:   "networkmanager-site",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *gn.GlobalNetworkId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// ?? Sites can affect the global network
					In: true,
					// The global network will definitely affect the site
					Out: true,
				},
			},
		},
	}

	switch gn.State {
	case types.GlobalNetworkStatePending:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.GlobalNetworkStateAvailable:
		item.Health = sdp.Health_HEALTH_OK.Enum()
	case types.GlobalNetworkStateUpdating:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.GlobalNetworkStateDeleting:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-global-network
// +overmind:descriptiveType Netwotkmanager Global Network
// +overmind:get Get a global network by id
// +overmind:list List all global networks
// +overmind:search Search for a global network by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_global_network.arn
// +overmind:terraform:method SEARCH

func NewGlobalNetworkSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, *networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, NetworkmanagerClient, *networkmanager.Options] {
	return &sources.AlwaysGetSource[*networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, *networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, NetworkmanagerClient, *networkmanager.Options]{
		ItemType:  "networkmanager-global-network",
		Client:    networkmanager.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		GetFunc:   globalNetworkGetFunc,
		GetInputMapper: func(scope, query string) *networkmanager.DescribeGlobalNetworksInput {
			return &networkmanager.DescribeGlobalNetworksInput{
				GlobalNetworkIds: []string{
					query,
				},
			}
		},
		ListInput: &networkmanager.DescribeGlobalNetworksInput{},
		ListFuncPaginatorBuilder: func(client NetworkmanagerClient, input *networkmanager.DescribeGlobalNetworksInput) sources.Paginator[*networkmanager.DescribeGlobalNetworksOutput, *networkmanager.Options] {
			return networkmanager.NewDescribeGlobalNetworksPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *networkmanager.DescribeGlobalNetworksOutput, input *networkmanager.DescribeGlobalNetworksInput) ([]*networkmanager.DescribeGlobalNetworksInput, error) {
			inputs := make([]*networkmanager.DescribeGlobalNetworksInput, 0)

			for _, gn := range output.GlobalNetworks {
				inputs = append(inputs, &networkmanager.DescribeGlobalNetworksInput{
					GlobalNetworkIds: []string{
						*gn.GlobalNetworkId, // This will be the id of the global network
					},
				})
			}

			return inputs, nil
		},
	}
}
