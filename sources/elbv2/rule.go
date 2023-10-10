package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func ruleOutputMapper(ctx context.Context, client elbClient, scope string, _ *elbv2.DescribeRulesInput, output *elbv2.DescribeRulesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	ruleArns := make([]string, 0)

	for _, rule := range output.Rules {
		if rule.RuleArn != nil {
			ruleArns = append(ruleArns, *rule.RuleArn)
		}
	}

	tagsMap, err := getTagsMap(ctx, client, ruleArns)

	if err != nil {
		return nil, err
	}

	for _, rule := range output.Rules {
		attrs, err := sources.ToAttributesCase(rule)

		if err != nil {
			return nil, err
		}

		var tags map[string]string

		if rule.RuleArn != nil {
			tags = tagsMap[*rule.RuleArn]
		}

		item := sdp.Item{
			Type:            "elbv2-rule",
			UniqueAttribute: "ruleArn",
			Attributes:      attrs,
			Scope:           scope,
			Tags:            tags,
		}

		var requests []*sdp.LinkedItemQuery

		for _, action := range rule.Actions {
			requests = ActionToRequests(action)
			item.LinkedItemQueries = append(item.LinkedItemQueries, requests...)
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type elbv2-rule
// +overmind:descriptiveType ELB Rule
// +overmind:get Get a rule by ARN
// +overmind:search Search for rules by listener ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_alb_listener_rule.arn
// +overmind:terraform:queryMap aws_lb_listener_rule.arn
// +overmind:terraform:method SEARCH

func NewRuleSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeRulesInput, *elbv2.DescribeRulesOutput, elbClient, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeRulesInput, *elbv2.DescribeRulesOutput, elbClient, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-rule",
		DescribeFunc: func(ctx context.Context, client elbClient, input *elbv2.DescribeRulesInput) (*elbv2.DescribeRulesOutput, error) {
			return client.DescribeRules(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeRulesInput, error) {
			return &elbv2.DescribeRulesInput{
				RuleArns: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeRulesInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for elbv2-rule, use search",
			}
		},
		InputMapperSearch: func(ctx context.Context, client elbClient, scope, query string) (*elbv2.DescribeRulesInput, error) {
			// Search by listener ARN
			return &elbv2.DescribeRulesInput{
				ListenerArn: &query,
			}, nil
		},
		OutputMapper: ruleOutputMapper,
	}
}
