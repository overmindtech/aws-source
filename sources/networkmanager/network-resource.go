package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

const (
	// networkmanager resources
	resourceTypeConnection = "connection"
	resourceTypeDevice     = "device"
	resourceTypeLink       = "link"
	resourceTypeSite       = "site"
	// directconnect resources
	resourceTypeDxCon     = "dxcon"
	resourceTypeDxGateway = "dx-gateway"
	resourceTypeDxVif     = "dx-vif"
	// ec2 VPC resources
	resourceTypeVPCCustomerGateway          = "customer-gateway"
	resourceTypeVPCTransitGateway           = "transit-gateway"
	resourceTypeVPCTransitGatewayAttachment = "transit-gateway-attachment"
	resourceTypeVPCTransitGatewayPeer       = "transit-gateway-connect-peer"
	resourceTypeVPCTransitGatewayRouteTable = "transit-gateway-route-table"
	resourceTypeVPCVPNConnection            = "vpn-connection"
)

// networkResourcesOutputMapper return a list of all connections, devices, links, sites in this global network
func networkResourcesOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, input *networkmanager.GetNetworkResourcesInput, output *networkmanager.GetNetworkResourcesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, r := range output.NetworkResources {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(r, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		attrs.Set("globalNetworkIdNetworkResourceArn", idWithGlobalNetwork(*input.GlobalNetworkId, *r.ResourceArn))

		item := sdp.Item{
			Type:            "networkmanager-network-resource",
			UniqueAttribute: "globalNetworkIdNetworkResourceArn",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(r.Tags),
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-global-network
						Type:   "networkmanager-global-network",
						Method: sdp.QueryMethod_GET,
						Query:  *input.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-network-resource-relationship
						Type:   "networkmanager-network-resource-relationship",
						Method: sdp.QueryMethod_GET,
						Query:  idWithGlobalNetwork(*input.GlobalNetworkId, *r.ResourceArn),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				},
			},
		}

		// This endpoint return collection of different resources
		switch *r.ResourceType {
		case resourceTypeConnection:
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-connection
					Type:   "networkmanager-connection",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, *r.ResourceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeDevice:
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-device
					Type:   "networkmanager-device",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, *r.ResourceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeLink:
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-link
					Type:   "networkmanager-link",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, *r.ResourceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeSite:
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-site
					Type:   "networkmanager-site",
					Method: sdp.QueryMethod_SEARCH,
					Query:  idWithGlobalNetwork(*input.GlobalNetworkId, *r.ResourceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeDxCon:
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link directconnect-connection
					Type:   "directconnect-connection",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeDxGateway:
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link directconnect-direct-connect-gateway
					Type:   "directconnect-direct-connect-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeDxVif:
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link directconnect-virtual-interface
					Type:   "directconnect-virtual-interface",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeVPCCustomerGateway:
			// TODO: add support for ec2-customer-gateway
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-customer-gateway
					Type:   "ec2-customer-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeVPCTransitGateway:
			// TODO: add support for ec2-transit-gateway
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway
					Type:   "ec2-transit-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeVPCTransitGatewayAttachment:
			// TODO: add support for ec2-transit-gateway-attachment
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway-attachment
					Type:   "ec2-transit-gateway-attachment",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeVPCTransitGatewayPeer:
			// TODO: add support for ec2-transit-gateway-connect-peer
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway-connect-peer
					Type:   "ec2-transit-gateway-connect-peer",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeVPCTransitGatewayRouteTable:
			// TODO: add support for ec2-transit-gateway-route-table
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-transit-gateway-route-table
					Type:   "ec2-transit-gateway-route-table",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		case resourceTypeVPCVPNConnection:
			// TODO: add support for ec2-vpn-connection
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link ec2-vpn-connection
					Type:   "ec2-vpn-connection",
					Method: sdp.QueryMethod_GET,
					Query:  *r.ResourceId,
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
// +overmind:type networkmanager-network-resource
// +overmind:descriptiveType Networkmanager Network Resources
// +overmind:get Get a Networkmanager Network Resources
// +overmind:list List all Networkmanager Network Resources
// +overmind:search Search for Networkmanager Network Resources by GlobalNetworkId
// +overmind:group AWS

func NewNetworkResourceSource(client *networkmanager.Client, accountID, region string) *sources.DescribeOnlySource[*networkmanager.GetNetworkResourcesInput, *networkmanager.GetNetworkResourcesOutput, *networkmanager.Client, *networkmanager.Options] {
	return &sources.DescribeOnlySource[*networkmanager.GetNetworkResourcesInput, *networkmanager.GetNetworkResourcesOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ItemType:  "networkmanager-network-resource",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetNetworkResourcesInput) (*networkmanager.GetNetworkResourcesOutput, error) {
			return client.GetNetworkResources(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*networkmanager.GetNetworkResourcesInput, error) {
			return &networkmanager.GetNetworkResourcesInput{
				GlobalNetworkId: sources.PtrString(query), // GlobalNetworkId, required
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetNetworkResourcesInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-network-resource, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetNetworkResourcesInput) sources.Paginator[*networkmanager.GetNetworkResourcesOutput, *networkmanager.Options] {
			return networkmanager.NewGetNetworkResourcesPaginator(client, params)
		},
		OutputMapper: networkResourcesOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetNetworkResourcesInput, error) {
			// Search by GlobalNetworkId
			return &networkmanager.GetNetworkResourcesInput{
				GlobalNetworkId: &query,
			}, nil
		},
	}
}
