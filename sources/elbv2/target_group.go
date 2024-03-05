package elbv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func targetGroupOutputMapper(ctx context.Context, client elbClient, scope string, _ *elbv2.DescribeTargetGroupsInput, output *elbv2.DescribeTargetGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	tgArns := make([]string, 0)

	for _, tg := range output.TargetGroups {
		if tg.TargetGroupArn != nil {
			tgArns = append(tgArns, *tg.TargetGroupArn)
		}
	}

	tagsMap := getTagsMap(ctx, client, tgArns)

	for _, tg := range output.TargetGroups {
		attrs, err := sources.ToAttributesCase(tg)

		if err != nil {
			return nil, err
		}

		var tags map[string]string

		if tg.TargetGroupArn != nil {
			tags = tagsMap[*tg.TargetGroupArn]
		}

		item := sdp.Item{
			Type:            "elbv2-target-group",
			UniqueAttribute: "targetGroupName",
			Attributes:      attrs,
			Scope:           scope,
			Tags:            tags,
		}

		if tg.TargetGroupArn != nil {
			// +overmind:link elbv2-target-health
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "elbv2-target-health",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *tg.TargetGroupArn,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Target groups and their target health are tightly coupled
					In:  true,
					Out: true,
				},
			})
		}

		if tg.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *tg.VpcId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the VPC can affect the target group
					In: true,
					// The target group won't affect the VPC
					Out: false,
				},
			})
		}

		for _, lbArn := range tg.LoadBalancerArns {
			if a, err := sources.ParseARN(lbArn); err == nil {
				// +overmind:link elbv2-load-balancer
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-load-balancer",
						Method: sdp.QueryMethod_SEARCH,
						Query:  lbArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Load balancers and their target groups are tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type elbv2-target-group
// +overmind:descriptiveType Target Group
// +overmind:get Get a target group by name
// +overmind:list List all target groups
// +overmind:search Search for target groups by load balancer ARN or target group ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_alb_target_group.arn
// +overmind:terraform:queryMap aws_lb_target_group.arn
// +overmind:terraform:method SEARCH

func NewTargetGroupSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeTargetGroupsInput, *elbv2.DescribeTargetGroupsOutput, elbClient, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeTargetGroupsInput, *elbv2.DescribeTargetGroupsOutput, elbClient, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-target-group",
		DescribeFunc: func(ctx context.Context, client elbClient, input *elbv2.DescribeTargetGroupsInput) (*elbv2.DescribeTargetGroupsOutput, error) {
			return client.DescribeTargetGroups(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeTargetGroupsInput, error) {
			return &elbv2.DescribeTargetGroupsInput{
				Names: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeTargetGroupsInput, error) {
			return &elbv2.DescribeTargetGroupsInput{}, nil
		},
		InputMapperSearch: func(ctx context.Context, client elbClient, scope, query string) (*elbv2.DescribeTargetGroupsInput, error) {
			arn, err := sources.ParseARN(query)

			if err != nil {
				return nil, err
			}

			switch arn.Type() {
			case "targetgroup":
				// Search by target group
				return &elbv2.DescribeTargetGroupsInput{
					TargetGroupArns: []string{
						query,
					},
				}, nil
			case "loadbalancer":
				// Search by load balancer
				return &elbv2.DescribeTargetGroupsInput{
					LoadBalancerArn: &query,
				}, nil
			default:
				return nil, fmt.Errorf("unsupported resource type: %s", arn.Resource)
			}
		},
		PaginatorBuilder: func(client elbClient, params *elbv2.DescribeTargetGroupsInput) sources.Paginator[*elbv2.DescribeTargetGroupsOutput, *elbv2.Options] {
			return elbv2.NewDescribeTargetGroupsPaginator(client, params)
		},
		OutputMapper: targetGroupOutputMapper,
	}
}
