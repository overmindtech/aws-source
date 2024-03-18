package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func coreNetworkGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.CoreNetwork, error) {
	out, err := client.GetCoreNetwork(ctx, &networkmanager.GetCoreNetworkInput{
		CoreNetworkId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.CoreNetwork, nil
}

func coreNetworkItemMapper(scope string, cn *types.CoreNetwork) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(cn)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "networkmanager-core-network",
		UniqueAttribute: "coreNetworkId",
		Attributes:      attributes,
		Scope:           scope,
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
				Query: &sdp.Query{
					// +overmind:link networkmanager-global-network
					Type:   "networkmanager-global-network",
					Method: sdp.QueryMethod_GET,
					Query:  *cn.GlobalNetworkId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			},
			{
				Query: &sdp.Query{
					// +overmind:link networkmanager-core-network-policy
					Type:   "networkmanager-core-network-policy",
					Method: sdp.QueryMethod_GET,
					Query:  *cn.CoreNetworkId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  false,
					Out: true,
				},
			},
			{
				Query: &sdp.Query{
					// +overmind:link networkmanager-connect-peer
					Type:   "networkmanager-connect-peer",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cn.CoreNetworkId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  false,
					Out: true,
				},
			},
		},
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-core-network
// +overmind:descriptiveType Networkmanager Core Network
// +overmind:get Get a Networkmanager Core Network by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_core_network.id

func NewCoreNetworkSource(client *networkmanager.Client, accountID, region string) *sources.GetListSource[*types.CoreNetwork, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.CoreNetwork, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-core-network",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.CoreNetwork, error) {
			return coreNetworkGetFunc(ctx, client, scope, query)
		},
		ItemMapper: coreNetworkItemMapper,

		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.CoreNetwork, error) {
			out, err := client.ListCoreNetworks(ctx, &networkmanager.ListCoreNetworksInput{})

			if err != nil {
				return nil, err
			}

			coreNetworks := make([]*types.CoreNetwork, len(out.CoreNetworks))

			for i, _ := range out.CoreNetworks {
				coreNetworks[i] = &types.CoreNetwork{
					CoreNetworkArn:  out.CoreNetworks[i].CoreNetworkArn,
					CoreNetworkId:   out.CoreNetworks[i].CoreNetworkId,
					Description:     out.CoreNetworks[i].Description,
					GlobalNetworkId: out.CoreNetworks[i].GlobalNetworkId,
					State:           out.CoreNetworks[i].State,
					Tags:            out.CoreNetworks[i].Tags,
				}
			}

			return coreNetworks, nil
		},
	}
}
