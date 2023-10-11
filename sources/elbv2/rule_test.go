package elbv2

import (
	"context"
	"testing"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestRuleOutputMapper(t *testing.T) {
	output := elbv2.DescribeRulesOutput{
		Rules: []types.Rule{
			{
				RuleArn:  sources.PtrString("arn:aws:elasticloadbalancing:eu-west-2:944651592624:listener-rule/app/ingress/1bf10920c5bd199d/9d28f512be129134/0f73a74d21b008f7"),
				Priority: sources.PtrString("1"),
				Conditions: []types.RuleCondition{
					{
						Field: sources.PtrString("path-pattern"),
						Values: []string{
							"/api/gateway",
						},
						PathPatternConfig: &types.PathPatternConditionConfig{
							Values: []string{
								"/api/gateway",
							},
						},
						HostHeaderConfig: &types.HostHeaderConditionConfig{
							Values: []string{
								"foo",
							},
						},
						HttpHeaderConfig: &types.HttpHeaderConditionConfig{
							HttpHeaderName: sources.PtrString("SOMETHING"),
							Values: []string{
								"foo",
							},
						},
						HttpRequestMethodConfig: &types.HttpRequestMethodConditionConfig{
							Values: []string{
								"GET",
							},
						},
						QueryStringConfig: &types.QueryStringConditionConfig{
							Values: []types.QueryStringKeyValuePair{
								{
									Key:   sources.PtrString("foo"),
									Value: sources.PtrString("bar"),
								},
							},
						},
						SourceIpConfig: &types.SourceIpConditionConfig{
							Values: []string{
								"1.1.1.1/24",
							},
						},
					},
				},
				Actions: []types.Action{
					// Tested in actions.go
				},
				IsDefault: false,
			},
		},
	}

	items, err := ruleOutputMapper(context.Background(), mockElbClient{}, "foo", nil, &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if item.Tags["foo"] != "bar" {
			t.Errorf("expected tag foo to be bar, got %v", item.Tags["foo"])
		}

		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}
}
