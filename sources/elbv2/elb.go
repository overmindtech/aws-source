package elbv2

import (
	"context"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func loadBalancerOutputMapper(ctx context.Context, client elbClient, scope string, _ *elbv2.DescribeLoadBalancersInput, output *elbv2.DescribeLoadBalancersOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	// Get the ARNs so that we can get the tags
	arns := make([]string, 0)

	for _, lb := range output.LoadBalancers {
		if lb.LoadBalancerArn != nil {
			arns = append(arns, *lb.LoadBalancerArn)
		}
	}

	tagsMap, err := getTagsMap(ctx, client, arns)

	if err != nil {
		return nil, err
	}

	for _, lb := range output.LoadBalancers {
		attrs, err := sources.ToAttributesCase(lb)

		if err != nil {
			return nil, err
		}

		var tags map[string]string

		if lb.LoadBalancerArn != nil {
			tags = tagsMap[*lb.LoadBalancerArn]
		}

		item := sdp.Item{
			Type:            "elbv2-load-balancer",
			UniqueAttribute: "loadBalancerName",
			Attributes:      attrs,
			Scope:           scope,
			Tags:            tags,
		}

		if lb.LoadBalancerArn != nil {
			// +overmind:link elbv2-target-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "elbv2-target-group",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *lb.LoadBalancerArn,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Load balancers and their target groups are tightly coupled
					In:  true,
					Out: true,
				},
			})

			// +overmind:link elbv2-listener
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "elbv2-listener",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *lb.LoadBalancerArn,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Load balancers and their listeners are tightly coupled
					In:  true,
					Out: true,
				},
			})
		}

		if lb.DNSName != nil {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *lb.DNSName,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// DNS always links
					In:  true,
					Out: true,
				},
			})
		}

		if lb.CanonicalHostedZoneId != nil {
			// +overmind:link route53-hosted-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "route53-hosted-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *lb.CanonicalHostedZoneId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the hosted zone could affect the LB
					In: true,
					// The LB won't affect the hosted zone
					Out: false,
				},
			})
		}

		if lb.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *lb.VpcId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the VPC could affect the LB
					In: true,
					// The LB won't affect the VPC
					Out: false,
				},
			})
		}

		for _, az := range lb.AvailabilityZones {
			if az.ZoneName != nil {
				// +overmind:link ec2-availability-zone
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-availability-zone",
						Method: sdp.QueryMethod_GET,
						Query:  *az.ZoneName,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the availability zone could affect the LB
						In: true,
						// The LB won't affect the availability zone
						Out: false,
					},
				})
			}

			if az.SubnetId != nil {
				// +overmind:link ec2-subnet
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-subnet",
						Method: sdp.QueryMethod_GET,
						Query:  *az.SubnetId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the subnet could affect the LB
						In: true,
						// The LB won't affect the subnet
						Out: false,
					},
				})
			}

			for _, address := range az.LoadBalancerAddresses {
				// +overmind:link ec2-address
				if address.AllocationId != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-address",
							Method: sdp.QueryMethod_GET,
							Query:  *address.AllocationId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the address could affect the LB
							In: true,
							// The LB can also affect the address
							Out: true,
						},
					})
				}

				if address.IPv6Address != nil {
					// +overmind:link ip
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *address.IPv6Address,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// IPs always link
							In:  true,
							Out: true,
						},
					})
				}

				if address.IpAddress != nil {
					// +overmind:link ip
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *address.IpAddress,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// IPs always link
							In:  true,
							Out: true,
						},
					})
				}

				if address.PrivateIPv4Address != nil {
					// +overmind:link ip
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ip",
							Method: sdp.QueryMethod_GET,
							Query:  *address.PrivateIPv4Address,
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
		}

		for _, sg := range lb.SecurityGroups {
			// +overmind:link ec2-security-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  sg,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the security group could affect the LB
					In: true,
					// The LB won't affect the security group
					Out: false,
				},
			})
		}

		if lb.CustomerOwnedIpv4Pool != nil {
			// +overmind:link ec2-coip-pool
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-coip-pool",
					Method: sdp.QueryMethod_GET,
					Query:  *lb.CustomerOwnedIpv4Pool,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the COIP pool could affect the LB
					In: true,
					// The LB won't affect the COIP pool
					Out: false,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type elbv2-load-balancer
// +overmind:descriptiveType Elastic Load Balancer
// +overmind:get Get an ELB by name
// +overmind:list List all ELBs
// +overmind:search Search for ELBs by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_lb.arn
// +overmind:terraform:method SEARCH

func NewLoadBalancerSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeLoadBalancersInput, *elbv2.DescribeLoadBalancersOutput, elbClient, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeLoadBalancersInput, *elbv2.DescribeLoadBalancersOutput, elbClient, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-load-balancer",
		DescribeFunc: func(ctx context.Context, client elbClient, input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
			return client.DescribeLoadBalancers(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeLoadBalancersInput, error) {
			return &elbv2.DescribeLoadBalancersInput{
				Names: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeLoadBalancersInput, error) {
			return &elbv2.DescribeLoadBalancersInput{}, nil
		},
		PaginatorBuilder: func(client elbClient, params *elbv2.DescribeLoadBalancersInput) sources.Paginator[*elbv2.DescribeLoadBalancersOutput, *elbv2.Options] {
			return elbv2.NewDescribeLoadBalancersPaginator(client, params)
		},
		OutputMapper: loadBalancerOutputMapper,
	}
}
