package elb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func loadBalancerOutputMapper(scope string, _ *elb.DescribeLoadBalancersInput, output *elb.DescribeLoadBalancersOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, desc := range output.LoadBalancerDescriptions {
		attrs, err := sources.ToAttributesCase(desc)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "elb-load-balancer",
			UniqueAttribute: "loadBalancerName",
			Attributes:      attrs,
			Scope:           scope,
		}

		if desc.DNSName != nil {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "dns",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *desc.DNSName,
				Scope:  "global",
			})
		}

		if desc.CanonicalHostedZoneName != nil {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "dns",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *desc.CanonicalHostedZoneName,
				Scope:  "global",
			})
		}

		if desc.CanonicalHostedZoneNameID != nil {
			// +overmind:link route53-hosted-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "route53-hosted-zone",
				Method: sdp.QueryMethod_GET,
				Query:  *desc.CanonicalHostedZoneNameID,
				Scope:  scope,
			})
		}

		for _, az := range desc.AvailabilityZones {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.QueryMethod_GET,
				Query:  az,
				Scope:  scope,
			})
		}

		for _, subnet := range desc.Subnets {
			// +overmind:link ec2-subnet
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-subnet",
				Method: sdp.QueryMethod_GET,
				Query:  subnet,
				Scope:  scope,
			})
		}

		if desc.VPCId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-vpc",
				Method: sdp.QueryMethod_GET,
				Query:  *desc.VPCId,
				Scope:  scope,
			})
		}

		for _, instance := range desc.Instances {
			if instance.InstanceId != nil {
				// +overmind:link ec2-instance
				// The EC2 instance itself
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.InstanceId,
					Scope:  scope,
				})

				if desc.LoadBalancerName != nil {
					name := InstanceHealthName{
						LoadBalancerName: *desc.LoadBalancerName,
						InstanceId:       *instance.InstanceId,
					}

					// +overmind:link elb-instance-health
					// The health for that instance
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "elb-instance-health",
						Method: sdp.QueryMethod_GET,
						Query:  name.String(),
						Scope:  scope,
					})
				}
			}
		}

		if desc.SourceSecurityGroup != nil {
			if desc.SourceSecurityGroup.GroupName != nil {
				// +overmind:link ec2-security-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *desc.SourceSecurityGroup.GroupName,
					Scope:  scope,
				})
			}
		}

		for _, sg := range desc.SecurityGroups {
			// +overmind:link ec2-security-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-security-group",
				Method: sdp.QueryMethod_GET,
				Query:  sg,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type elb-load-balancer
// +overmind:descriptiveType Classic Load Balancer
// +overmind:get Get a classic load balancer by name
// +overmind:list List all classic load balancers
// +overmind:search Search for classic load balancers by ARN
// +overmind:group AWS

func NewLoadBalancerSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elb.DescribeLoadBalancersInput, *elb.DescribeLoadBalancersOutput, *elb.Client, *elb.Options] {
	return &sources.DescribeOnlySource[*elb.DescribeLoadBalancersInput, *elb.DescribeLoadBalancersOutput, *elb.Client, *elb.Options]{
		Config:    config,
		Client:    elb.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elb-load-balancer",
		DescribeFunc: func(ctx context.Context, client *elb.Client, input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
			return client.DescribeLoadBalancers(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elb.DescribeLoadBalancersInput, error) {
			return &elb.DescribeLoadBalancersInput{
				LoadBalancerNames: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elb.DescribeLoadBalancersInput, error) {
			return &elb.DescribeLoadBalancersInput{}, nil
		},
		PaginatorBuilder: func(client *elb.Client, params *elb.DescribeLoadBalancersInput) sources.Paginator[*elb.DescribeLoadBalancersOutput, *elb.Options] {
			return elb.NewDescribeLoadBalancersPaginator(client, params)
		},
		OutputMapper: loadBalancerOutputMapper,
	}
}
