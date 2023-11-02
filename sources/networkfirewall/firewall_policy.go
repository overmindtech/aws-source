package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type unifiedFirewallPolicy struct {
	types.FirewallPolicyResponse
	FirewallPolicy *types.FirewallPolicy
}

func firewallPolicyGetFunc(ctx context.Context, client networkFirewallClient, scope string, input *networkfirewall.DescribeFirewallPolicyInput) (*sdp.Item, error) {
	resp, err := client.DescribeFirewallPolicy(ctx, input)

	if err != nil {
		return nil, err
	}

	ufp := unifiedFirewallPolicy{
		FirewallPolicyResponse: *resp.FirewallPolicyResponse,
		FirewallPolicy:         resp.FirewallPolicy,
	}

	attributes, err := sources.ToAttributesCase(ufp)

	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)

	for _, tag := range resp.FirewallPolicyResponse.Tags {
		tags[*tag.Key] = *tag.Value
	}

	var health *sdp.Health

	if resp.FirewallPolicyResponse != nil {
		switch resp.FirewallPolicyResponse.FirewallPolicyStatus {
		case types.ResourceStatusActive:
			health = sdp.Health_HEALTH_OK.Enum()
		case types.ResourceStatusDeleting:
			health = sdp.Health_HEALTH_PENDING.Enum()
		case types.ResourceStatusError:
			health = sdp.Health_HEALTH_ERROR.Enum()
		}
	}

	item := sdp.Item{
		Type:            "network-firewall-firewall-policy",
		UniqueAttribute: "firewallPolicyName",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            tags,
		Health:          health,
	}

	//+overmind:link kms-key
	item.LinkedItemQueries = append(item.LinkedItemQueries, encryptionConfigurationLink(ufp.EncryptionConfiguration, scope))

	ruleGroupArns := make([]string, 0)

	for _, ruleGroup := range resp.FirewallPolicy.StatelessRuleGroupReferences {
		if ruleGroup.ResourceArn != nil {
			ruleGroupArns = append(ruleGroupArns, *ruleGroup.ResourceArn)
		}
	}

	for _, ruleGroup := range resp.FirewallPolicy.StatefulRuleGroupReferences {
		if ruleGroup.ResourceArn != nil {
			ruleGroupArns = append(ruleGroupArns, *ruleGroup.ResourceArn)
		}
	}

	for _, arn := range ruleGroupArns {
		if a, err := sources.ParseARN(arn); err == nil {
			//+overmind:link network-firewall-rule-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "network-firewall-rule-group",
					Query:  arn,
					Method: sdp.QueryMethod_SEARCH,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}
	}

	if resp.FirewallPolicy.TLSInspectionConfigurationArn != nil {
		if a, err := sources.ParseARN(*resp.FirewallPolicy.TLSInspectionConfigurationArn); err == nil {
			//+overmind:link network-firewall-tls-inspection-configuration
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "network-firewall-tls-inspection-configuration",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *resp.FirewallPolicy.TLSInspectionConfigurationArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type network-firewall-firewall-policy
// +overmind:descriptiveType Network Firewall Policy
// +overmind:get Get a Network Firewall Policy by name
// +overmind:list List Network Firewall Policies
// +overmind:search Search for Network Firewall Policies by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkfirewall_firewall_policy.name

func NewFirewallPolicySource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*networkfirewall.ListFirewallPoliciesInput, *networkfirewall.ListFirewallPoliciesOutput, *networkfirewall.DescribeFirewallPolicyInput, *networkfirewall.DescribeFirewallPolicyOutput, networkFirewallClient, *networkfirewall.Options] {
	return &sources.AlwaysGetSource[*networkfirewall.ListFirewallPoliciesInput, *networkfirewall.ListFirewallPoliciesOutput, *networkfirewall.DescribeFirewallPolicyInput, *networkfirewall.DescribeFirewallPolicyOutput, networkFirewallClient, *networkfirewall.Options]{
		ItemType:  "network-firewall-firewall-policy",
		Client:    networkfirewall.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &networkfirewall.ListFirewallPoliciesInput{},
		GetInputMapper: func(scope, query string) *networkfirewall.DescribeFirewallPolicyInput {
			return &networkfirewall.DescribeFirewallPolicyInput{
				FirewallPolicyName: &query,
			}
		},
		SearchGetInputMapper: func(scope, query string) (*networkfirewall.DescribeFirewallPolicyInput, error) {
			return &networkfirewall.DescribeFirewallPolicyInput{
				FirewallPolicyArn: &query,
			}, nil
		},
		ListFuncPaginatorBuilder: func(client networkFirewallClient, input *networkfirewall.ListFirewallPoliciesInput) sources.Paginator[*networkfirewall.ListFirewallPoliciesOutput, *networkfirewall.Options] {
			return networkfirewall.NewListFirewallPoliciesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *networkfirewall.ListFirewallPoliciesOutput, input *networkfirewall.ListFirewallPoliciesInput) ([]*networkfirewall.DescribeFirewallPolicyInput, error) {
			var inputs []*networkfirewall.DescribeFirewallPolicyInput

			for _, firewall := range output.FirewallPolicies {
				inputs = append(inputs, &networkfirewall.DescribeFirewallPolicyInput{
					FirewallPolicyArn: firewall.Arn,
				})
			}
			return inputs, nil
		},
		GetFunc: func(ctx context.Context, client networkFirewallClient, scope string, input *networkfirewall.DescribeFirewallPolicyInput) (*sdp.Item, error) {
			return firewallPolicyGetFunc(ctx, client, scope, input)
		},
	}
}
