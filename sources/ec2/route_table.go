package ec2

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func routeTableInputMapperGet(scope string, query string) (*ec2.DescribeRouteTablesInput, error) {
	return &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{
			query,
		},
	}, nil
}

func routeTableInputMapperList(scope string) (*ec2.DescribeRouteTablesInput, error) {
	return &ec2.DescribeRouteTablesInput{}, nil
}

func routeTableOutputMapper(scope string, _ *ec2.DescribeRouteTablesInput, output *ec2.DescribeRouteTablesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, rt := range output.RouteTables {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(rt)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-route-table",
			UniqueAttribute: "routeTableId",
			Scope:           scope,
			Attributes:      attrs,
		}

		for _, assoc := range rt.Associations {
			if assoc.SubnetId != nil {
				// +overmind:link ec2-subnet
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-subnet",
						Method: sdp.QueryMethod_GET,
						Query:  *assoc.SubnetId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// All things in a route table could affect each other
						// since changing the target could affect the
						// traffic that is routed to it. And changing the route
						// table could affect the target
						In:  true,
						Out: true,
					},
				})
			}

			if assoc.GatewayId != nil {
				// +overmind:link ec2-internet-gateway
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-internet-gateway",
						Method: sdp.QueryMethod_GET,
						Query:  *assoc.GatewayId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
		}

		for _, route := range rt.Routes {
			if route.GatewayId != nil {
				if strings.HasPrefix(*route.GatewayId, "igw") {
					// +overmind:link ec2-internet-gateway
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-internet-gateway",
							Method: sdp.QueryMethod_GET,
							Query:  *route.GatewayId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							In:  true,
							Out: true,
						},
					})
				}
				if strings.HasPrefix(*route.GatewayId, "vpce") {
					// +overmind:link ec2-vpc-endpoint
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-vpc-endpoint",
							Method: sdp.QueryMethod_GET,
							Query:  *route.GatewayId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							In:  true,
							Out: true,
						},
					})
				}
			}
			if route.CarrierGatewayId != nil {
				// +overmind:link ec2-carrier-gateway
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-carrier-gateway",
						Method: sdp.QueryMethod_GET,
						Query:  *route.CarrierGatewayId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
			if route.EgressOnlyInternetGatewayId != nil {
				// +overmind:link ec2-egress-only-internet-gateway
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-egress-only-internet-gateway",
						Method: sdp.QueryMethod_GET,
						Query:  *route.EgressOnlyInternetGatewayId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
			if route.InstanceId != nil {
				// +overmind:link ec2-instance
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-instance",
						Method: sdp.QueryMethod_GET,
						Query:  *route.InstanceId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
			if route.LocalGatewayId != nil {
				// +overmind:link ec2-local-gateway
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-local-gateway",
						Method: sdp.QueryMethod_GET,
						Query:  *route.LocalGatewayId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
			if route.NatGatewayId != nil {
				// +overmind:link ec2-nat-gateway
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-nat-gateway",
						Method: sdp.QueryMethod_GET,
						Query:  *route.NatGatewayId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
			if route.NetworkInterfaceId != nil {
				// +overmind:link ec2-network-interface
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-network-interface",
						Method: sdp.QueryMethod_GET,
						Query:  *route.NetworkInterfaceId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
			if route.TransitGatewayId != nil {
				// +overmind:link ec2-transit-gateway
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-transit-gateway",
						Method: sdp.QueryMethod_GET,
						Query:  *route.TransitGatewayId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
			if route.VpcPeeringConnectionId != nil {
				// +overmind:link ec2-vpc-peering-connection
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-vpc-peering-connection",
						Method: sdp.QueryMethod_GET,
						Query:  *route.VpcPeeringConnectionId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: true,
					},
				})
			}
		}

		if rt.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *rt.VpcId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-route-table
// +overmind:descriptiveType Route Table
// +overmind:get Get a route table by ID
// +overmind:list List all route tables
// +overmind:search Search route tables by ARN
// +overmind:group AWS

func NewRouteTableSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeRouteTablesInput, *ec2.DescribeRouteTablesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeRouteTablesInput, *ec2.DescribeRouteTablesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-route-table",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeRouteTables(ctx, input)
		},
		InputMapperGet:  routeTableInputMapperGet,
		InputMapperList: routeTableInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeRouteTablesInput) sources.Paginator[*ec2.DescribeRouteTablesOutput, *ec2.Options] {
			return ec2.NewDescribeRouteTablesPaginator(client, params)
		},
		OutputMapper: routeTableOutputMapper,
	}
}
