package networkmanager

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func deviceOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, _ *networkmanager.GetDevicesInput, output *networkmanager.GetDevicesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, s := range output.Devices {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = adapters.ToAttributesWithExclude(s, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		if s.GlobalNetworkId == nil || s.DeviceId == nil {
			return nil, sdp.NewQueryError(errors.New("globalNetworkId or deviceId is nil for device"))
		}

		attrs.Set("GlobalNetworkIdDeviceId", idWithGlobalNetwork(*s.GlobalNetworkId, *s.DeviceId))

		item := sdp.Item{
			Type:            "networkmanager-device",
			UniqueAttribute: "GlobalNetworkIdDeviceId",
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
						// +overmind:link networkmanager-link-association
						Type:   "networkmanager-link-association",
						Method: sdp.QueryMethod_SEARCH,
						Query:  idWithTypeAndGlobalNetwork(*s.GlobalNetworkId, "device", *s.DeviceId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-connection
						Type:   "networkmanager-connection",
						Method: sdp.QueryMethod_SEARCH,
						Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.DeviceId),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
			},
		}

		if s.SiteId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-site
					Type:   "networkmanager-site",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.SiteId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}

		if s.DeviceArn != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-network-resource-relationship
					Type:   "networkmanager-network-resource-relationship",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.DeviceArn),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		switch s.State {
		case types.DeviceStatePending:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.DeviceStateAvailable:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.DeviceStateDeleting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.DeviceStateUpdating:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-device
// +overmind:descriptiveType Networkmanager Device
// +overmind:get Get a Networkmanager Device
// +overmind:search Search for Networkmanager Devices by GlobalNetworkId, or by GlobalNetworkId with SiteId
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_device.arn
// +overmind:terraform:method SEARCH

func NewDeviceSource(client *networkmanager.Client, accountID string) *adapters.DescribeOnlySource[*networkmanager.GetDevicesInput, *networkmanager.GetDevicesOutput, *networkmanager.Client, *networkmanager.Options] {
	return &adapters.DescribeOnlySource[*networkmanager.GetDevicesInput, *networkmanager.GetDevicesOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		ItemType:  "networkmanager-device",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetDevicesInput) (*networkmanager.GetDevicesOutput, error) {
			return client.GetDevices(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetDevicesInput, error) {
			// We are using a custom id of {globalNetworkId}|{deviceId}
			sections := strings.Split(query, "|")

			if len(sections) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-device get function",
				}
			}
			return &networkmanager.GetDevicesInput{
				GlobalNetworkId: &sections[0],
				DeviceIds: []string{
					sections[1],
				},
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetDevicesInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-device, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetDevicesInput) adapters.Paginator[*networkmanager.GetDevicesOutput, *networkmanager.Options] {
			return networkmanager.NewGetDevicesPaginator(client, params)
		},
		OutputMapper: deviceOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetDevicesInput, error) {
			// We may search by only globalNetworkId or by using a custom id of {globalNetworkId}|{siteId}
			sections := strings.Split(query, "|")
			switch len(sections) {
			case 1:
				// globalNetworkId
				return &networkmanager.GetDevicesInput{
					GlobalNetworkId: &sections[0],
				}, nil
			case 2:
				// {globalNetworkId}|{siteId}
				return &networkmanager.GetDevicesInput{
					GlobalNetworkId: &sections[0],
					SiteId:          &sections[1],
				}, nil
			default:
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-device get function",
				}
			}

		},
	}
}
