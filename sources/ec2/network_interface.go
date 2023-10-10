package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func networkInterfaceInputMapperGet(scope string, query string) (*ec2.DescribeNetworkInterfacesInput, error) {
	return &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{
			query,
		},
	}, nil
}

func networkInterfaceInputMapperList(scope string) (*ec2.DescribeNetworkInterfacesInput, error) {
	return &ec2.DescribeNetworkInterfacesInput{}, nil
}

func networkInterfaceOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeNetworkInterfacesInput, output *ec2.DescribeNetworkInterfacesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ni := range output.NetworkInterfaces {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ni)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-network-interface",
			UniqueAttribute: "networkInterfaceId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if ni.Attachment != nil {
			if ni.Attachment.InstanceId != nil {
				// +overmind:link ec2-instance
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-instance",
						Method: sdp.QueryMethod_GET,
						Query:  *ni.Attachment.InstanceId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// The instance and the interface are closely linked
						// and affect each other
						In:  true,
						Out: true,
					},
				})
			}
		}

		if ni.AvailabilityZone != nil {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-availability-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *ni.AvailabilityZone,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// AZs don't change
					In:  false,
					Out: false,
				},
			})
		}

		for _, sg := range ni.Groups {
			if sg.GroupId != nil {
				// +overmind:link ec2-security-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-security-group",
						Method: sdp.QueryMethod_GET,
						Query:  *sg.GroupId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// A security group will affect an interface
						In: true,
						// An interface won't affect a security group
						Out: false,
					},
				})
			}
		}

		for _, ip := range ni.Ipv6Addresses {
			if ip.Ipv6Address != nil {
				// +overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *ip.Ipv6Address,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// IPs are always linked
						In:  true,
						Out: true,
					},
				})
			}
		}

		for _, ip := range ni.PrivateIpAddresses {
			if assoc := ip.Association; assoc != nil {
				if assoc.PublicDnsName != nil {
					// +overmind:link dns
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "dns",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *assoc.PublicDnsName,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// DNS names are always linked
							In:  true,
							Out: true,
						},
					})
				}

				if assoc.PublicIp != nil {
					// +overmind:link ip
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *assoc.PublicIp,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// IPs are always linked
							In:  true,
							Out: true,
						},
					})
				}

				if assoc.CarrierIp != nil {
					// +overmind:link ip
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *assoc.CarrierIp,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// IPs are always linked
							In:  true,
							Out: true,
						},
					})
				}

				if assoc.CustomerOwnedIp != nil {
					// +overmind:link ip
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *assoc.CustomerOwnedIp,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// IPs are always linked
							In:  true,
							Out: true,
						},
					})
				}
			}

			if ip.PrivateDnsName != nil {
				// +overmind:link dns
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "dns",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *ip.PrivateDnsName,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// DNS names are always linked
						In:  true,
						Out: true,
					},
				})
			}

			if ip.PrivateIpAddress != nil {
				// +overmind:link ip
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *ip.PrivateIpAddress,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// IPs are always linked
						In:  true,
						Out: true,
					},
				})
			}
		}

		if ni.SubnetId != nil {
			// +overmind:link ec2-subnet
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  *ni.SubnetId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the subnet will affect interfaces within that
					// subnet
					In: true,
					// Changing the interface won't affect the subnet
					Out: false,
				},
			})
		}

		if ni.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *ni.VpcId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the VPC will affect interfaces within that VPC
					In: true,
					// Changing the interface won't affect the VPC
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-network-interface
// +overmind:descriptiveType EC2 Network Interface
// +overmind:get Get a network interface by ID
// +overmind:list List all network interfaces
// +overmind:search Search network interfaces by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_network_interface.id

func NewNetworkInterfaceSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeNetworkInterfacesInput, *ec2.DescribeNetworkInterfacesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeNetworkInterfacesInput, *ec2.DescribeNetworkInterfacesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-network-interface",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeNetworkInterfaces(ctx, input)
		},
		InputMapperGet:  networkInterfaceInputMapperGet,
		InputMapperList: networkInterfaceInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeNetworkInterfacesInput) sources.Paginator[*ec2.DescribeNetworkInterfacesOutput, *ec2.Options] {
			return ec2.NewDescribeNetworkInterfacesPaginator(client, params)
		},
		OutputMapper: networkInterfaceOutputMapper,
	}
}
