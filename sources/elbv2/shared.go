package elbv2

import (
	"context"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type elbClient interface {
	DescribeTags(ctx context.Context, params *elbv2.DescribeTagsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTagsOutput, error)
	DescribeLoadBalancers(ctx context.Context, params *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeListeners(ctx context.Context, params *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error)
	DescribeRules(ctx context.Context, params *elbv2.DescribeRulesInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeRulesOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error)
}

func tagsToMap(tags []types.Tag) map[string]string {
	m := make(map[string]string)

	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			m[*tag.Key] = *tag.Value
		}
	}

	return m
}

// Gets a map of ARN to tags (in map[string]string format) for the given ARNs
func getTagsMap(ctx context.Context, client elbClient, arns []string) (map[string]map[string]string, error) {
	tagsMap := make(map[string]map[string]string)

	tagsOut, err := client.DescribeTags(ctx, &elbv2.DescribeTagsInput{
		ResourceArns: arns,
	})

	if err != nil {
		return nil, err
	}

	for _, tagDescription := range tagsOut.TagDescriptions {
		if tagDescription.ResourceArn != nil {
			tagsMap[*tagDescription.ResourceArn] = tagsToMap(tagDescription.Tags)
		}
	}

	return tagsMap, nil
}
