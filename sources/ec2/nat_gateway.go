package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func NatGatewayInputMapperGet(scope string, query string) (*ec2.DescribeNatGatewaysInput, error) {
	return &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{
			query,
		},
	}, nil
}

func NatGatewayInputMapperList(scope string) (*ec2.DescribeNatGatewaysInput, error) {
	return &ec2.DescribeNatGatewaysInput{}, nil
}

func NatGatewayOutputMapper(scope string, output *ec2.DescribeNatGatewaysOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ng := range output.NatGateways {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ng)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-nat-gateway",
			UniqueAttribute: "natGatewayId",
			Scope:           scope,
			Attributes:      attrs,
		}

		for _, address := range ng.NatGatewayAddresses {
			if address.NetworkInterfaceId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-network-interface",
					Method: sdp.RequestMethod_GET,
					Query:  *address.NetworkInterfaceId,
					Scope:  scope,
				})
			}

			if address.PrivateIp != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *address.PrivateIp,
					Scope:  "global",
				})
			}

			if address.PublicIp != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ip",
					Method: sdp.RequestMethod_GET,
					Query:  *address.PublicIp,
					Scope:  "global",
				})
			}
		}

		if ng.SubnetId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-subnet",
				Method: sdp.RequestMethod_GET,
				Query:  *ng.SubnetId,
				Scope:  scope,
			})
		}

		if ng.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *ng.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewNatGatewaySource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeNatGatewaysInput, *ec2.DescribeNatGatewaysOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeNatGatewaysInput, *ec2.DescribeNatGatewaysOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-nat-gateway",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeNatGatewaysInput) (*ec2.DescribeNatGatewaysOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeNatGateways(ctx, input)
		},
		InputMapperGet:  NatGatewayInputMapperGet,
		InputMapperList: NatGatewayInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeNatGatewaysInput) sources.Paginator[*ec2.DescribeNatGatewaysOutput, *ec2.Options] {
			return ec2.NewDescribeNatGatewaysPaginator(client, params)
		},
		OutputMapper: NatGatewayOutputMapper,
	}
}
