package elbv2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
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
								"foo.bar.com", // link
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
				IsDefault: sources.PtrBool(false),
			},
		},
	}

	items, err := ruleOutputMapper(context.Background(), mockElbClient{}, "foo", nil, &output)

	if err != nil {
		t.Error(err)
	}

	if len(items) != 1 {
		t.Error("expected 1 item")
	}

	item := items[0]

	tests := sources.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "foo.bar.com",
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewRuleSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)
	client := elasticloadbalancingv2.NewFromConfig(config)

	lbSource := NewLoadBalancerSource(client, account, region)
	listenerSource := NewListenerSource(client, account, region)
	ruleSource := NewRuleSource(client, account, region)

	lbs, err := lbSource.List(context.Background(), lbSource.Scopes()[0], false)
	if err != nil {
		t.Fatal(err)
	}

	if len(lbs) == 0 {
		t.Skip("no load balancers found")
	}

	lbARN, err := lbs[0].GetAttributes().Get("LoadBalancerArn")
	if err != nil {
		t.Fatal(err)
	}

	listeners, err := listenerSource.Search(context.Background(), listenerSource.Scopes()[0], fmt.Sprint(lbARN), false)
	if err != nil {
		t.Fatal(err)
	}

	if len(listeners) == 0 {
		t.Skip("no listeners found")
	}

	listenerARN, err := listeners[0].GetAttributes().Get("ListenerArn")
	if err != nil {
		t.Fatal(err)
	}

	goodSearch := fmt.Sprint(listenerARN)

	test := sources.E2ETest{
		Adapter:         ruleSource,
		Timeout:         10 * time.Second,
		GoodSearchQuery: &goodSearch,
		SkipList:        true,
	}

	test.Run(t)
}
