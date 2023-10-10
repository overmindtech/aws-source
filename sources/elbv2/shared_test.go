package elbv2

import (
	"context"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
)

type mockElbClient struct{}

func (m mockElbClient) DescribeTags(ctx context.Context, params *elbv2.DescribeTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTagsOutput, error) {
	tagDescriptions := make([]types.TagDescription, 0)

	for _, arn := range params.ResourceArns {
		tagDescriptions = append(tagDescriptions, types.TagDescription{
			ResourceArn: &arn,
			Tags: []types.Tag{
				{
					Key:   sources.PtrString("foo"),
					Value: sources.PtrString("bar"),
				},
			},
		})
	}

	return &elbv2.DescribeTagsOutput{
		TagDescriptions: tagDescriptions,
	}, nil
}

func (m mockElbClient) DescribeLoadBalancers(ctx context.Context, params *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error) {
	return nil, nil
}

func (m mockElbClient) DescribeListeners(ctx context.Context, params *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error) {
	return nil, nil
}

func (m mockElbClient) DescribeRules(ctx context.Context, params *elbv2.DescribeRulesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeRulesOutput, error) {
	return nil, nil
}

func (m mockElbClient) DescribeTargetGroups(ctx context.Context, params *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error) {
	return nil, nil
}
