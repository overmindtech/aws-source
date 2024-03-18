package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
	"strings"
)

func connectionOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, _ *networkmanager.GetConnectionsInput, output *networkmanager.GetConnectionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, s := range output.Connections {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(s, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		attrs.Set("globalNetworkIdConnectionId", idWithGlobalNetwork(*s.GlobalNetworkId, *s.ConnectionId))

		item := sdp.Item{
			Type:            "networkmanager-connection",
			UniqueAttribute: "globalNetworkIdConnectionId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(s.Tags),
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-global-network
						Type:   "networkmanager-global-network",
						Method: sdp.QueryMethod_GET,
						Query:  *s.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-link
						Type:   "networkmanager-link",
						Method: sdp.QueryMethod_GET,
						Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.LinkId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-link
						Type:   "networkmanager-link",
						Method: sdp.QueryMethod_GET,
						Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.ConnectedLinkId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-device
						Type:   "networkmanager-device",
						Method: sdp.QueryMethod_GET,
						Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.DeviceId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-device
						Type:   "networkmanager-device",
						Method: sdp.QueryMethod_GET,
						Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.ConnectedDeviceId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
			},
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-connection
// +overmind:descriptiveType Networkmanager Connection
// +overmind:get Get a Networkmanager Connection
// +overmind:list List all Networkmanager Connections
// +overmind:search Search for Networkmanager Connections by GlobalNetworkId
// +overmind:group AWS

func NewConnectionSource(client *networkmanager.Client, accountID, region string) *sources.DescribeOnlySource[*networkmanager.GetConnectionsInput, *networkmanager.GetConnectionsOutput, *networkmanager.Client, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetConnectionsInput, *networkmanager.GetConnectionsOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-connection",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetConnectionsInput) (*networkmanager.GetConnectionsOutput, error) {
			return client.GetConnections(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetConnectionsInput, error) {
			// We are using a custom id of {globalNetworkId}|{connectionId}
			sections := strings.Split(query, "|")

			if len(sections) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-connection get function",
				}
			}
			return &networkmanager.GetConnectionsInput{
				GlobalNetworkId: &sections[0],
				ConnectionIds: []string{
					sections[1],
				},
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetConnectionsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-connection, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetConnectionsInput) sources.Paginator[*networkmanager.GetConnectionsOutput, *networkmanager.Options] {
			return networkmanager.NewGetConnectionsPaginator(client, params)
		},
		OutputMapper: connectionOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetConnectionsInput, error) {
			// We may search by only globalNetworkId or by using a custom id of {globalNetworkId}|{deviceId}
			sections := strings.Split(query, "|")
			switch len(sections) {
			case 1:
				// globalNetworkId
				return &networkmanager.GetConnectionsInput{
					GlobalNetworkId: &sections[0],
				}, nil
			case 2:
				// {globalNetworkId}|{deviceId}
				return &networkmanager.GetConnectionsInput{
					GlobalNetworkId: &sections[0],
					DeviceId:        &sections[1],
				}, nil
			default:
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-connection get function",
				}
			}
		},
	}
}
