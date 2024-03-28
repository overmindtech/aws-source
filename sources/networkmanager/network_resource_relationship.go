package networkmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func networkResourceRelationshipOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, input *networkmanager.GetNetworkResourceRelationshipsInput, output *networkmanager.GetNetworkResourceRelationshipsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)
	// Connecting networkmanager-global-network with internal or external resources happening in
	// networkmanager-network-resource source
	// No point to double-link same resources to networkmanager-global-network here again
	// Instead here we will create connections between these resources itself

	for _, relationship := range output.Relationships {
		if relationship.From == nil || relationship.To == nil {
			continue
		}

		// Define item FROM
		arnFrom, err := sources.ParseARN(*relationship.From)
		if err != nil {
			return nil, err
		}
		fromResourceType := fmt.Sprintf("%s-%s", arnFrom.Service, arnFrom.Type())
		// For each item we have to set correct UniqueAttribute
		uniqueAttrName, uniqueAttrVal := "", ""
		switch fromResourceType {
		case "networkmanager-connection":
			uniqueAttrName = "globalNetworkIdConnectionId"
			uniqueAttrVal = idWithGlobalNetwork(*input.GlobalNetworkId, arnFrom.ResourceID())
		case "networkmanager-device":
			uniqueAttrName = "globalNetworkIdDeviceId"
			uniqueAttrVal = idWithGlobalNetwork(*input.GlobalNetworkId, arnFrom.ResourceID())
		case "networkmanager-link":
			uniqueAttrName = "globalNetworkIdLinkId"
			uniqueAttrVal = idWithGlobalNetwork(*input.GlobalNetworkId, arnFrom.ResourceID())
		case "networkmanager-site":
			uniqueAttrName = "globalNetworkIdSiteId"
			uniqueAttrVal = idWithGlobalNetwork(*input.GlobalNetworkId, arnFrom.ResourceID())
		case "directconnect-connection":
			uniqueAttrName = "connectionId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "directconnect-direct-connect-gateway":
			uniqueAttrName = "directConnectGatewayId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "directconnect-virtual-interface":
			uniqueAttrName = "virtualInterfaceId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "ec2-customer-gateway":
			// TODO: add support for ec2-customer-gateway
			uniqueAttrName = "customerGatewayId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "ec2-transit-gateway":
			// TODO: add support for ec2-transit-gateway
			uniqueAttrName = "transitGatewayId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "ec2-transit-gateway-attachment":
			// TODO: add support for ec2-transit-gateway-attachment
			uniqueAttrName = "transitGatewayAttachmentId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "ec2-transit-gateway-connect-peer":
			// TODO: add support for ec2-transit-gateway-connect-peer
			uniqueAttrName = "transitGatewayConnectPeerId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "ec2-transit-gateway-route-table":
			// TODO: add support for ec2-transit-gateway-route-table
			uniqueAttrName = "transitGatewayRouteTableId"
			uniqueAttrVal = arnFrom.ResourceID()
		case "ec2-vpn-connection":
			// TODO: add support for ec2-vpn-connection
			uniqueAttrName = "vpnConnectionId"
			uniqueAttrVal = arnFrom.ResourceID()
		default:
			// skip unknown item types
			continue
		}
		attrs, err := sdp.ToAttributes(map[string]interface{}{
			uniqueAttrName: uniqueAttrVal,
		})
		if err != nil {
			return nil, err
		}
		item := sdp.Item{
			Type:              fmt.Sprintf("%s-%s", arnFrom.Service, arnFrom.Resource),
			UniqueAttribute:   uniqueAttrName,
			Scope:             scope,
			Attributes:        attrs,
			LinkedItemQueries: []*sdp.LinkedItemQuery{},
		}

		// Define item TO
		arnTo, err := sources.ParseARN(*relationship.To)
		if err != nil {
			return nil, err
		}
		toResourceType := fmt.Sprintf("%s-%s", arnTo.Service, arnTo.Type())
		// For each linked item we must define +overmind:link comment section
		switch toResourceType {
		case "networkmanager-connection":
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-connection
					Type:   "networkmanager-connection",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, arnTo.ResourceID()),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "networkmanager-device":
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-device
					Type:   "networkmanager-device",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, arnTo.ResourceID()),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "networkmanager-link":
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-link
					Type:   "networkmanager-link",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, arnTo.ResourceID()),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "networkmanager-site":
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-site
					Type:   "networkmanager-site",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, arnTo.ResourceID()),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "directconnect-connection":
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link directconnect-connection
					Type:   "directconnect-connection",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "directconnect-direct-connect-gateway":
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link directconnect-direct-connect-gateway
					Type:   "directconnect-direct-connect-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "directconnect-virtual-interface":
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link directconnect-virtual-interface
					Type:   "directconnect-virtual-interface",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "ec2-customer-gateway":
			// TODO: add support for ec2-customer-gateway
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-customer-gateway
					Type:   "ec2-customer-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "ec2-transit-gateway":
			// TODO: add support for ec2-transit-gateway
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway
					Type:   "ec2-transit-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "ec2-transit-gateway-attachment":
			// TODO: add support for ec2-transit-gateway-attachment
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway-attachment
					Type:   "ec2-transit-gateway-attachment",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "ec2-transit-gateway-connect-peer":
			// TODO: add support for ec2-transit-gateway-connect-peer
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway-connect-peer
					Type:   "ec2-transit-gateway-connect-peer",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "ec2-transit-gateway-route-table":
			// TODO: add support for ec2-transit-gateway-route-table
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway-route-table
					Type:   "ec2-transit-gateway-route-table",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case "ec2-vpn-connection":
			// TODO: add support for ec2-vpn-connection
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-vpn-connection
					Type:   "ec2-vpn-connection",
					Method: sdp.QueryMethod_GET,
					Query:  arnTo.ResourceID(),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		default:
			// skip unknown item types
			continue
		}
		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-network-resource-relationship
// +overmind:descriptiveType Networkmanager Network Resource Relationships
// +overmind:get Get a Networkmanager Network Resource Relationship by GlobalNetworkId and ResourceARN
// +overmind:list List all Networkmanager NetworkResourceRelationships
// +overmind:search Search for Networkmanager NetworkResourceRelationships by GlobalNetworkId
// +overmind:group AWS

func NewNetworkResourceRelationshipsSource(client *networkmanager.Client, accountID, region string) *sources.DescribeOnlySource[*networkmanager.GetNetworkResourceRelationshipsInput, *networkmanager.GetNetworkResourceRelationshipsOutput, *networkmanager.Client, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetNetworkResourceRelationshipsInput, *networkmanager.GetNetworkResourceRelationshipsOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-network-resource-relationship",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetNetworkResourceRelationshipsInput) (*networkmanager.GetNetworkResourceRelationshipsOutput, error) {
			return client.GetNetworkResourceRelationships(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetNetworkResourceRelationshipsInput, error) {
			// We are using a custom id of {globalNetworkId}|{resourceARN}
			sections := strings.Split(query, "|")

			if len(sections) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-network-resource-relationship get function",
				}
			}
			return &networkmanager.GetNetworkResourceRelationshipsInput{
				GlobalNetworkId: &sections[0],
				ResourceArn:     &sections[1],
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetNetworkResourceRelationshipsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-network-resource-relationship, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetNetworkResourceRelationshipsInput) sources.Paginator[*networkmanager.GetNetworkResourceRelationshipsOutput, *networkmanager.Options] {
			return networkmanager.NewGetNetworkResourceRelationshipsPaginator(client, params)
		},
		OutputMapper: networkResourceRelationshipOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetNetworkResourceRelationshipsInput, error) {
			// Search by GlobalNetworkId
			return &networkmanager.GetNetworkResourceRelationshipsInput{
				GlobalNetworkId: &query,
			}, nil
		},
	}
}
