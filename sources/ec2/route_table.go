package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func RouteTableInputMapperGet(scope string, query string) (*ec2.DescribeRouteTablesInput, error) {
	return &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{
			query,
		},
	}, nil
}

func RouteTableInputMapperList(scope string) (*ec2.DescribeRouteTablesInput, error) {
	return &ec2.DescribeRouteTablesInput{}, nil
}

func RouteTableOutputMapper(scope string, output *ec2.DescribeRouteTablesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, rt := range output.RouteTables {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(rt)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
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
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-subnet",
					Method: sdp.RequestMethod_GET,
					Query:  *assoc.SubnetId,
					Scope:  scope,
				})
			}

			if assoc.GatewayId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-internet-gateway",
					Method: sdp.RequestMethod_GET,
					Query:  *assoc.GatewayId,
					Scope:  scope,
				})
			}
		}

		for _, route := range rt.Routes {
			if route.CarrierGatewayId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-carrier-gateway",
					Method: sdp.RequestMethod_GET,
					Query:  *route.CarrierGatewayId,
					Scope:  scope,
				})
			}
			if route.EgressOnlyInternetGatewayId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-egress-only-internet-gateway",
					Method: sdp.RequestMethod_GET,
					Query:  *route.EgressOnlyInternetGatewayId,
					Scope:  scope,
				})
			}
			if route.InstanceId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-instance",
					Method: sdp.RequestMethod_GET,
					Query:  *route.InstanceId,
					Scope:  scope,
				})
			}
			if route.LocalGatewayId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-local-gateway",
					Method: sdp.RequestMethod_GET,
					Query:  *route.LocalGatewayId,
					Scope:  scope,
				})
			}
			if route.NatGatewayId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-nat-gateway",
					Method: sdp.RequestMethod_GET,
					Query:  *route.NatGatewayId,
					Scope:  scope,
				})
			}
			if route.NetworkInterfaceId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-network-interface",
					Method: sdp.RequestMethod_GET,
					Query:  *route.NetworkInterfaceId,
					Scope:  scope,
				})
			}
			if route.TransitGatewayId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-transit-gateway",
					Method: sdp.RequestMethod_GET,
					Query:  *route.TransitGatewayId,
					Scope:  scope,
				})
			}
			if route.VpcPeeringConnectionId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-vpc-peering-connection",
					Method: sdp.RequestMethod_GET,
					Query:  *route.VpcPeeringConnectionId,
					Scope:  scope,
				})
			}
		}

		if rt.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *rt.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewRouteTableSource(config aws.Config, accountID string) *sources.AWSSource[*ec2.DescribeRouteTablesInput, *ec2.DescribeRouteTablesOutput, *ec2.Client, *ec2.Options] {
	return &sources.AWSSource[*ec2.DescribeRouteTablesInput, *ec2.DescribeRouteTablesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-route-table",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
			return client.DescribeRouteTables(ctx, input)
		},
		InputMapperGet:  RouteTableInputMapperGet,
		InputMapperList: RouteTableInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeRouteTablesInput) sources.Paginator[*ec2.DescribeRouteTablesOutput, *ec2.Options] {
			return ec2.NewDescribeRouteTablesPaginator(client, params)
		},
		OutputMapper: RouteTableOutputMapper,
	}
}
