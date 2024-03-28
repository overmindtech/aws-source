package networkmanager

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func transitGatewayConnectPeerAssociationsOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, _ *networkmanager.GetTransitGatewayConnectPeerAssociationsInput, output *networkmanager.GetTransitGatewayConnectPeerAssociationsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, a := range output.TransitGatewayConnectPeerAssociations {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(a, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		attrs.Set("globalNetworkIdWithTransitGatewayConnectPeerArn", idWithGlobalNetwork(*a.GlobalNetworkId, *a.TransitGatewayConnectPeerArn))

		item := sdp.Item{
			Type:            "networkmanager-transit-gateway-connect-peer-association",
			UniqueAttribute: "globalNetworkIdWithTransitGatewayConnectPeerArn",
			Scope:           scope,
			Attributes:      attrs,
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-global-network
						Type:   "networkmanager-global-network",
						Method: sdp.QueryMethod_GET,
						Query:  *a.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
			},
		}

		if a.DeviceId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-device
					Type:   "networkmanager-device",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*a.GlobalNetworkId, *a.DeviceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		if a.LinkId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-link
					Type:   "networkmanager-link",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*a.GlobalNetworkId, *a.LinkId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		switch a.State {
		case types.TransitGatewayConnectPeerAssociationStatePending:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.TransitGatewayConnectPeerAssociationStateAvailable:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.TransitGatewayConnectPeerAssociationStateDeleting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.TransitGatewayConnectPeerAssociationStateDeleted:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-transit-gateway-connect-peer-association
// +overmind:descriptiveType Networkmanager Connect Peer Associations
// +overmind:get Get a Networkmanager Transit GatewayConnect Peer Associations
// +overmind:list List all Networkmanager Transit Gateway Connect Peer Associations
// +overmind:search Search for Networkmanager TransitGatewayConnectPeerAssociations by GlobalNetworkId
// +overmind:group AWS

func NewTransitGatewayConnectPeerAssociationSource(client *networkmanager.Client, accountID, region string) *sources.DescribeOnlySource[*networkmanager.GetTransitGatewayConnectPeerAssociationsInput, *networkmanager.GetTransitGatewayConnectPeerAssociationsOutput, *networkmanager.Client, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetTransitGatewayConnectPeerAssociationsInput, *networkmanager.GetTransitGatewayConnectPeerAssociationsOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-transit-gateway-connect-peer-association",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetTransitGatewayConnectPeerAssociationsInput) (*networkmanager.GetTransitGatewayConnectPeerAssociationsOutput, error) {
			return client.GetTransitGatewayConnectPeerAssociations(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetTransitGatewayConnectPeerAssociationsInput, error) {
			sections := strings.Split(query, "|")
			switch len(sections) {
			case 1:
				// only GlobalNetworkId
				return &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{
					GlobalNetworkId: &sections[0],
				}, nil
			case 2:
				// we are using a custom id of {globalNetworkId}|{networkmanager-connect-peer.ID}
				// e.g. searching from networkmanager-connect-peer
				return &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{
					GlobalNetworkId: &sections[0],
					TransitGatewayConnectPeerArns: []string{
						sections[1],
					},
				}, nil
			default:
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-transit-gateway-connect-peer-association get function",
				}
			}
		},
		InputMapperList: func(scope string) (*networkmanager.GetTransitGatewayConnectPeerAssociationsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-transit-gateway-connect-peer-association, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetTransitGatewayConnectPeerAssociationsInput) sources.Paginator[*networkmanager.GetTransitGatewayConnectPeerAssociationsOutput, *networkmanager.Options] {
			return networkmanager.NewGetTransitGatewayConnectPeerAssociationsPaginator(client, params)
		},
		OutputMapper: transitGatewayConnectPeerAssociationsOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetTransitGatewayConnectPeerAssociationsInput, error) {
			// Search by GlobalNetworkId
			return &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{
				GlobalNetworkId: &query,
			}, nil
		},
	}
}
