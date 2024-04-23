package networkmanager

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func connectPeerAssociationsOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, _ *networkmanager.GetConnectPeerAssociationsInput, output *networkmanager.GetConnectPeerAssociationsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, a := range output.ConnectPeerAssociations {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(a)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		if a.GlobalNetworkId == nil || a.ConnectPeerId == nil {
			return nil, sdp.NewQueryError(errors.New("globalNetworkId or connectPeerId is nil for connect peer association"))
		}

		attrs.Set("globalNetworkIdConnectPeerId", idWithGlobalNetwork(*a.GlobalNetworkId, *a.ConnectPeerId))

		item := sdp.Item{
			Type:            "networkmanager-connect-peer-association",
			UniqueAttribute: "globalNetworkIdConnectPeerId",
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
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-connect-peer
						Type:   "networkmanager-connect-peer",
						Method: sdp.QueryMethod_GET,
						Query:  *a.ConnectPeerId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
			},
		}

		switch a.State {
		case types.ConnectPeerAssociationStatePending:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.ConnectPeerAssociationStateAvailable:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.ConnectPeerAssociationStateDeleting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.ConnectPeerAssociationStateDeleted:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		}

		if a.DeviceId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-device
					Type:   "networkmanager-device",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*a.GlobalNetworkId, *a.DeviceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}

		if a.LinkId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-link
					Type:   "networkmanager-link",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*a.GlobalNetworkId, *a.LinkId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-connect-peer-association
// +overmind:descriptiveType Networkmanager Connect Peer Associations
// +overmind:get Get a Networkmanager Connect Peer Associations
// +overmind:list List all Networkmanager Connect Peer Associations
// +overmind:search Search for Networkmanager ConnectPeerAssociations by GlobalNetworkId
// +overmind:group AWS

func NewConnectPeerAssociationSource(client *networkmanager.Client, accountID string, region string) *sources.DescribeOnlySource[*networkmanager.GetConnectPeerAssociationsInput, *networkmanager.GetConnectPeerAssociationsOutput, *networkmanager.Client, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetConnectPeerAssociationsInput, *networkmanager.GetConnectPeerAssociationsOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-connect-peer-association",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetConnectPeerAssociationsInput) (*networkmanager.GetConnectPeerAssociationsOutput, error) {
			return client.GetConnectPeerAssociations(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetConnectPeerAssociationsInput, error) {
			// We are using a custom id of {globalNetworkId}|{connectPeerId}
			sections := strings.Split(query, "|")

			if len(sections) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-connect-peer-association get function",
				}
			}
			return &networkmanager.GetConnectPeerAssociationsInput{
				GlobalNetworkId: &sections[0],
				ConnectPeerIds: []string{
					sections[1],
				},
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetConnectPeerAssociationsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-connect-peer-association, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetConnectPeerAssociationsInput) sources.Paginator[*networkmanager.GetConnectPeerAssociationsOutput, *networkmanager.Options] {
			return networkmanager.NewGetConnectPeerAssociationsPaginator(client, params)
		},
		OutputMapper: connectPeerAssociationsOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetConnectPeerAssociationsInput, error) {
			// Search by GlobalNetworkId
			return &networkmanager.GetConnectPeerAssociationsInput{
				GlobalNetworkId: &query,
			}, nil
		},
	}
}
