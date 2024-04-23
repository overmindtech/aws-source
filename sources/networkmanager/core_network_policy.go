package networkmanager

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func coreNetworkPolicyGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.CoreNetworkPolicy, error) {
	out, err := client.GetCoreNetworkPolicy(ctx, &networkmanager.GetCoreNetworkPolicyInput{
		CoreNetworkId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.CoreNetworkPolicy, nil
}

func coreNetworkPolicyItemMapper(scope string, cn *types.CoreNetworkPolicy) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(cn)
	if err != nil {
		return nil, err
	}

	if cn.CoreNetworkId == nil {
		return nil, sdp.NewQueryError(errors.New("coreNetworkId is nil for core network policy"))
	}

	item := sdp.Item{
		Type:            "networkmanager-core-network-policy",
		UniqueAttribute: "coreNetworkId",
		Attributes:      attributes,
		Scope:           scope,
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
				Query: &sdp.Query{
					// +overmind:link networkmanager-core-network
					Type:   "networkmanager-core-network",
					Method: sdp.QueryMethod_GET,
					Query:  *cn.CoreNetworkId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			},
		},
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-core-network-policy
// +overmind:descriptiveType Networkmanager Core Network Policy
// +overmind:get Get a Networkmanager Core Network Policy by Core Network id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_core_network_policy.core_network_id

func NewCoreNetworkPolicySource(client *networkmanager.Client, accountID, region string) *sources.GetListSource[*types.CoreNetworkPolicy, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.CoreNetworkPolicy, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-core-network-policy",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.CoreNetworkPolicy, error) {
			return coreNetworkPolicyGetFunc(ctx, client, scope, query)
		},
		ItemMapper: coreNetworkPolicyItemMapper,
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.CoreNetworkPolicy, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-core-network-policy, use get",
			}
		},
	}
}
