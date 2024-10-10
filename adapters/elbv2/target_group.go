package elbv2

import (
	"context"
	"fmt"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/overmindtech/aws-source/adapters"
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
		attrs, err := adapters.ToAttributesWithExclude(tg)

		if err != nil {
			return nil, err
		}

		var tags map[string]string

		if tg.TargetGroupArn != nil {
			tags = tagsMap[*tg.TargetGroupArn]
		}

		item := sdp.Item{
			Type:            "elbv2-target-group",
			UniqueAttribute: "TargetGroupName",
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
			if a, err := adapters.ParseARN(lbArn); err == nil {
				// +overmind:link elbv2-load-balancer
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-load-balancer",
						Method: sdp.QueryMethod_SEARCH,
						Query:  lbArn,
						Scope:  adapters.FormatScope(a.AccountID, a.Region),
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

func NewTargetGroupAdapter(client elbClient, accountID string, region string) *adapters.DescribeOnlyAdapter[*elbv2.DescribeTargetGroupsInput, *elbv2.DescribeTargetGroupsOutput, elbClient, *elbv2.Options] {
	return &adapters.DescribeOnlyAdapter[*elbv2.DescribeTargetGroupsInput, *elbv2.DescribeTargetGroupsOutput, elbClient, *elbv2.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "elbv2-target-group",
		AdapterMetadata: TargetGroupMetadata(),
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
			arn, err := adapters.ParseARN(query)

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
		PaginatorBuilder: func(client elbClient, params *elbv2.DescribeTargetGroupsInput) adapters.Paginator[*elbv2.DescribeTargetGroupsOutput, *elbv2.Options] {
			return elbv2.NewDescribeTargetGroupsPaginator(client, params)
		},
		OutputMapper: targetGroupOutputMapper,
	}
}

func TargetGroupMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "elbv2-target-group",
		DescriptiveName: "Target Group",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a target group by name",
			ListDescription:   "List all target groups",
			SearchDescription: "Search for target groups by load balancer ARN or target group ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{
				TerraformQueryMap: "aws_alb_target_group.arn",
				TerraformMethod:   sdp.QueryMethod_SEARCH,
			},
			{
				TerraformQueryMap: "aws_lb_target_group.arn",
				TerraformMethod:   sdp.QueryMethod_SEARCH,
			},
		},
		PotentialLinks: []string{"ec2-vpc", "elbv2-load-balancer", "elbv2-target-health"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_OBSERVABILITY,
	}
}
