package networkmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func globalNetworkOutputMapper(_ context.Context, client *networkmanager.Client, scope string, _ *networkmanager.DescribeGlobalNetworksInput, output *networkmanager.DescribeGlobalNetworksOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, gn := range output.GlobalNetworks {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = adapters.ToAttributesWithExclude(gn, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "networkmanager-global-network",
			UniqueAttribute: "GlobalNetworkId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(gn.Tags),
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-site
						Type:   "networkmanager-site",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-transit-gateway-registration
						Type:   "networkmanager-transit-gateway-registration",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-connect-peer-association
						Type:   "networkmanager-connect-peer-association",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-transit-gateway-connect-peer-association
						Type:   "networkmanager-transit-gateway-connect-peer-association",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-network-resource
						Type:   "networkmanager-network-resource",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-network-resource-relationship
						Type:   "networkmanager-network-resource-relationship",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-link
						Type:   "networkmanager-link",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-device
						Type:   "networkmanager-device",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-connection
						Type:   "networkmanager-connection",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *gn.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
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
// +overmind:descriptiveType Network Manager Global Network
// +overmind:get Get a global network by id
// +overmind:list List all global networks
// +overmind:search Search for a global network by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_global_network.arn
// +overmind:terraform:method SEARCH

func NewGlobalNetworkSource(client *networkmanager.Client, accountID string) *adapters.DescribeOnlySource[*networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, *networkmanager.Client, *networkmanager.Options] {
	return &adapters.DescribeOnlySource[*networkmanager.DescribeGlobalNetworksInput, *networkmanager.DescribeGlobalNetworksOutput, *networkmanager.Client, *networkmanager.Options]{
		ItemType:  "networkmanager-global-network",
		Client:    client,
		AccountID: accountID,
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.DescribeGlobalNetworksInput) (*networkmanager.DescribeGlobalNetworksOutput, error) {
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
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.DescribeGlobalNetworksInput) adapters.Paginator[*networkmanager.DescribeGlobalNetworksOutput, *networkmanager.Options] {
			return networkmanager.NewDescribeGlobalNetworksPaginator(client, params)
		},
		OutputMapper: globalNetworkOutputMapper,
	}
}

// idWithGlobalNetwork makes custom ID of given entity with global network ID and this entity ID/ARN
func idWithGlobalNetwork(gn, idOrArn string) string {
	return fmt.Sprintf("%s|%s", gn, idOrArn)
}

// idWithTypeAndGlobalNetwork makes custom ID of given entity with global network ID and this entity type and ID/ARN
func idWithTypeAndGlobalNetwork(gb, rType, idOrArn string) string {
	return fmt.Sprintf("%s|%s|%s", gb, rType, idOrArn)
}
