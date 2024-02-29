package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func globalNetworkOutputMapper(_ context.Context, _ NetworkmanagerClient, scope string, _ *networkmanager.DescribeGlobalNetworksInput, output *networkmanager.DescribeGlobalNetworksOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, gn := range output.GlobalNetworks {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(gn, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "networkmanager-global-network",
			UniqueAttribute: "globalNetworkId",
			Scope:           scope,
			Attributes:      attrs,
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
		items = append(items, &item)
	}

	return items, nil
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

func NewGlobalNetworkSource(config aws.Config, accountID string, region string) *sources.DescribeOnlySource[*networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, NetworkmanagerClient, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, NetworkmanagerClient, *networkmanager.Options]{
		ItemType:  "networkmanager-global-network",
		Client:    networkmanager.NewFromConfig(config),
		AccountID: accountID,
		DescribeFunc: func(ctx context.Context, client NetworkmanagerClient, input *networkmanager.DescribeGlobalNetworksInput) (*networkmanager.DescribeGlobalNetworksOutput, error) {
			return client.DescribeGlobalNetworks(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.DescribeGlobalNetworksInput, error) {
			return &networkmanager.DescribeGlobalNetworksInput{
				GlobalNetworkIds: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.DescribeGlobalNetworksInput, error) {
			return &networkmanager.DescribeGlobalNetworksInput{}, nil
		},
		PaginatorBuilder: func(client NetworkmanagerClient, params *networkmanager.DescribeGlobalNetworksInput) sources.Paginator[*networkmanager.DescribeGlobalNetworksOutput, *networkmanager.Options] {
			return networkmanager.NewDescribeGlobalNetworksPaginator(client, params)
		},
		OutputMapper: globalNetworkOutputMapper,
	}
}
