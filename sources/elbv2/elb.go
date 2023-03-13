package elbv2

import (
	"context"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func LoadBalancerOutputMapper(scope string, _ *elbv2.DescribeLoadBalancersInput, output *elbv2.DescribeLoadBalancersOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, lb := range output.LoadBalancers {
		attrs, err := sources.ToAttributesCase(lb)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "elbv2-load-balancer",
			UniqueAttribute: "loadBalancerName",
			Attributes:      attrs,
			Scope:           scope,
		}

		if lb.LoadBalancerArn != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "elbv2-target-group",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *lb.LoadBalancerArn,
				Scope:  scope,
			})

			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "elbv2-listener",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *lb.LoadBalancerArn,
				Scope:  scope,
			})
		}

		if lb.DNSName != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "dns",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *lb.DNSName,
				Scope:  "global",
			})
		}

		if lb.CanonicalHostedZoneId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "route53-hosted-zone",
				Method: sdp.QueryMethod_GET,
				Query:  *lb.CanonicalHostedZoneId,
				Scope:  scope,
			})
		}

		if lb.VpcId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-vpc",
				Method: sdp.QueryMethod_GET,
				Query:  *lb.VpcId,
				Scope:  scope,
			})
		}

		for _, az := range lb.AvailabilityZones {
			if az.ZoneName != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-availability-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *az.ZoneName,
					Scope:  scope,
				})
			}

			if az.SubnetId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  *az.SubnetId,
					Scope:  scope,
				})
			}

			for _, address := range az.LoadBalancerAddresses {
				if address.AllocationId != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ec2-address",
						Method: sdp.QueryMethod_GET,
						Query:  *address.AllocationId,
						Scope:  scope,
					})
				}

				if address.IPv6Address != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *address.IPv6Address,
						Scope:  "global",
					})
				}

				if address.IpAddress != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *address.IpAddress,
						Scope:  "global",
					})
				}

				if address.PrivateIPv4Address != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  *address.PrivateIPv4Address,
						Scope:  "global",
					})
				}
			}
		}

		for _, sg := range lb.SecurityGroups {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-security-group",
				Method: sdp.QueryMethod_GET,
				Query:  sg,
				Scope:  scope,
			})
		}

		if lb.CustomerOwnedIpv4Pool != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-coip-pool",
				Method: sdp.QueryMethod_GET,
				Query:  *lb.CustomerOwnedIpv4Pool,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewLoadBalancerSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeLoadBalancersInput, *elbv2.DescribeLoadBalancersOutput, *elbv2.Client, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeLoadBalancersInput, *elbv2.DescribeLoadBalancersOutput, *elbv2.Client, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-load-balancer",
		DescribeFunc: func(ctx context.Context, client *elbv2.Client, input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {
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
		PaginatorBuilder: func(client *elbv2.Client, params *elbv2.DescribeLoadBalancersInput) sources.Paginator[*elbv2.DescribeLoadBalancersOutput, *elbv2.Options] {
			return elbv2.NewDescribeLoadBalancersPaginator(client, params)
		},
		OutputMapper: LoadBalancerOutputMapper,
	}
}
