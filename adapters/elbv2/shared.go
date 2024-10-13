package elbv2

import (
	"context"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/adapterhelpers"
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
func getTagsMap(ctx context.Context, client elbClient, arns []string) map[string]map[string]string {
	tagsMap := make(map[string]map[string]string)

	if len(arns) > 0 {
		tagsOut, err := client.DescribeTags(ctx, &elbv2.DescribeTagsInput{
			ResourceArns: arns,
		})
		if err != nil {
			tags := adapterhelpers.HandleTagsError(ctx, err)

			// Set these tags for all ARNs
			for _, arn := range arns {
				tagsMap[arn] = tags
			}

			return tagsMap
		}

		for _, tagDescription := range tagsOut.TagDescriptions {
			if tagDescription.ResourceArn != nil {
				tagsMap[*tagDescription.ResourceArn] = tagsToMap(tagDescription.Tags)
			}
		}
	}

	return tagsMap
}
