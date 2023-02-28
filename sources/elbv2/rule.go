package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func RuleOutputMapper(scope string, _ *elbv2.DescribeRulesInput, output *elbv2.DescribeRulesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, rule := range output.Rules {
		attrs, err := sources.ToAttributesCase(rule)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "elbv2-rule",
			UniqueAttribute: "ruleArn",
			Attributes:      attrs,
			Scope:           scope,
		}

		var requests []*sdp.ItemRequest

		for _, action := range rule.Actions {
			requests = ActionToRequests(action)
			item.LinkedItemRequests = append(item.LinkedItemRequests, requests...)
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewRuleSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeRulesInput, *elbv2.DescribeRulesOutput, *elbv2.Client, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeRulesInput, *elbv2.DescribeRulesOutput, *elbv2.Client, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-rule",
		DescribeFunc: func(ctx context.Context, client *elbv2.Client, input *elbv2.DescribeRulesInput) (*elbv2.DescribeRulesOutput, error) {
			return client.DescribeRules(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeRulesInput, error) {
			return &elbv2.DescribeRulesInput{
				RuleArns: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeRulesInput, error) {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: "list not supported for elbv2-rule, use search",
			}
		},
		InputMapperSearch: func(ctx context.Context, client *elbv2.Client, scope, query string) (*elbv2.DescribeRulesInput, error) {
			// Search by listener ARN
			return &elbv2.DescribeRulesInput{
				ListenerArn: &query,
			}, nil
		},
		OutputMapper: RuleOutputMapper,
	}
}
