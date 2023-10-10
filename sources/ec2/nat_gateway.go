package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func natGatewayInputMapperGet(scope string, query string) (*ec2.DescribeNatGatewaysInput, error) {
	return &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{
			query,
		},
	}, nil
}

func natGatewayInputMapperList(scope string) (*ec2.DescribeNatGatewaysInput, error) {
	return &ec2.DescribeNatGatewaysInput{}, nil
}

func natGatewayOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeNatGatewaysInput, output *ec2.DescribeNatGatewaysOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ng := range output.NatGateways {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ng)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
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
				// +overmind:link ec2-network-interface
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-network-interface",
						Method: sdp.QueryMethod_GET,
						Query:  *address.NetworkInterfaceId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// The nat gateway and it's interfaces will affect each
						// other
						In:  true,
						Out: true,
					},
				})
			}

			if address.PrivateIp != nil {
				// +overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *address.PrivateIp,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// IPs always link
						In:  true,
						Out: true,
					},
				})
			}

			if address.PublicIp != nil {
				// +overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *address.PublicIp,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// IPs always link
						In:  true,
						Out: true,
					},
				})
			}
		}

		if ng.SubnetId != nil {
			// +overmind:link ec2-subnet
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  *ng.SubnetId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the subnet won't affect the gateway
					In: false,
					// Changing the gateway will affect the subnet since this
					// will be gateway that subnet uses to access the internet
					Out: true,
				},
			})
		}

		if ng.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *ng.VpcId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the VPC could affect the gateway
					In: true,
					// Changing the gateway won't affect the VPC
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-nat-gateway
// +overmind:descriptiveType NAT Gateway
// +overmind:get Get a NAT Gateway by ID
// +overmind:list List all NAT gateways
// +overmind:search Search for NAT gateways by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_nat_gateway.id

func NewNatGatewaySource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeNatGatewaysInput, *ec2.DescribeNatGatewaysOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeNatGatewaysInput, *ec2.DescribeNatGatewaysOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-nat-gateway",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeNatGatewaysInput) (*ec2.DescribeNatGatewaysOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeNatGateways(ctx, input)
		},
		InputMapperGet:  natGatewayInputMapperGet,
		InputMapperList: natGatewayInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeNatGatewaysInput) sources.Paginator[*ec2.DescribeNatGatewaysOutput, *ec2.Options] {
			return ec2.NewDescribeNatGatewaysPaginator(client, params)
		},
		OutputMapper: natGatewayOutputMapper,
	}
}
