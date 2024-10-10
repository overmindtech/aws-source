package networkfirewall

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func (c testNetworkFirewallClient) DescribeRuleGroup(ctx context.Context, params *networkfirewall.DescribeRuleGroupInput, optFns ...func(*networkfirewall.Options)) (*networkfirewall.DescribeRuleGroupOutput, error) {
	now := time.Now()

	return &networkfirewall.DescribeRuleGroupOutput{
		RuleGroupResponse: &types.RuleGroupResponse{
			RuleGroupArn:  adapters.PtrString("arn:aws:network-firewall:us-east-1:123456789012:stateless-rulegroup/aws-network-firewall-DefaultStatelessRuleGroup-1J3Z3W2ZQXV3"),
			RuleGroupId:   adapters.PtrString("test"),
			RuleGroupName: adapters.PtrString("test"),
			AnalysisResults: []types.AnalysisResult{
				{
					AnalysisDetail: adapters.PtrString("test"),
					IdentifiedRuleIds: []string{
						"test",
					},
					IdentifiedType: types.IdentifiedTypeStatelessRuleContainsTcpFlags,
				},
			},
			Capacity:         adapters.PtrInt32(1),
			ConsumedCapacity: adapters.PtrInt32(1),
			Description:      adapters.PtrString("test"),
			EncryptionConfiguration: &types.EncryptionConfiguration{
				Type:  types.EncryptionTypeAwsOwnedKmsKey,
				KeyId: adapters.PtrString("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"), // link (this can be an ARN or ID)
			},
			LastModifiedTime:     &now,
			NumberOfAssociations: adapters.PtrInt32(1),
			RuleGroupStatus:      types.ResourceStatusActive,                                                                                           // health
			SnsTopic:             adapters.PtrString("arn:aws:sns:us-east-1:123456789012:aws-network-firewall-DefaultStatelessRuleGroup-1J3Z3W2ZQXV3"), // link
			SourceMetadata: &types.SourceMetadata{
				SourceArn:         adapters.PtrString("arn:aws:network-firewall:us-east-1:123456789012:firewall/aws-network-firewall-DefaultFirewall-1J3Z3W2ZQXV3"), // link
				SourceUpdateToken: adapters.PtrString("test"),
			},
			Tags: []types.Tag{
				{
					Key:   adapters.PtrString("test"),
					Value: adapters.PtrString("test"),
				},
			},
			Type: types.RuleGroupTypeStateless,
		},
		RuleGroup: &types.RuleGroup{
			RulesSource: &types.RulesSource{
				RulesSourceList: &types.RulesSourceList{
					GeneratedRulesType: types.GeneratedRulesTypeAllowlist,
					TargetTypes: []types.TargetType{
						types.TargetTypeHttpHost,
					},
					Targets: []string{
						"foo.bar.com", // link
					},
				},
				RulesString: adapters.PtrString("test"),
				StatefulRules: []types.StatefulRule{
					{
						Action: types.StatefulActionAlert,
						Header: &types.Header{
							Destination:     adapters.PtrString("1.1.1.1"),
							DestinationPort: adapters.PtrString("8080"),
							Direction:       types.StatefulRuleDirectionForward,
							Protocol:        types.StatefulRuleProtocolDcerpc,
							Source:          adapters.PtrString("test"),
							SourcePort:      adapters.PtrString("8080"),
						},
					},
				},
				StatelessRulesAndCustomActions: &types.StatelessRulesAndCustomActions{
					StatelessRules: []types.StatelessRule{
						{
							Priority: adapters.PtrInt32(1),
							RuleDefinition: &types.RuleDefinition{
								Actions: []string{},
								MatchAttributes: &types.MatchAttributes{
									DestinationPorts: []types.PortRange{
										{
											FromPort: 1,
											ToPort:   1,
										},
									},
									Destinations: []types.Address{
										{
											AddressDefinition: adapters.PtrString("1.1.1.1/1"),
										},
									},
									Protocols: []int32{1},
									SourcePorts: []types.PortRange{
										{
											FromPort: 1,
											ToPort:   1,
										},
									},
									Sources: []types.Address{},
									TCPFlags: []types.TCPFlagField{
										{
											Flags: []types.TCPFlag{
												types.TCPFlagAck,
											},
											Masks: []types.TCPFlag{
												types.TCPFlagEce,
											},
										},
									},
								},
							},
						},
					},
					CustomActions: []types.CustomAction{
						{
							ActionDefinition: &types.ActionDefinition{
								PublishMetricAction: &types.PublishMetricAction{
									Dimensions: []types.Dimension{
										{
											Value: adapters.PtrString("test"),
										},
									},
								},
							},
							ActionName: adapters.PtrString("test"),
						},
					},
				},
			},
		},
	}, nil
}

func (c testNetworkFirewallClient) ListRuleGroups(ctx context.Context, params *networkfirewall.ListRuleGroupsInput, optFns ...func(*networkfirewall.Options)) (*networkfirewall.ListRuleGroupsOutput, error) {
	return &networkfirewall.ListRuleGroupsOutput{
		RuleGroups: []types.RuleGroupMetadata{
			{
				Arn: adapters.PtrString("arn:aws:network-firewall:us-east-1:123456789012:stateless-rulegroup/aws-network-firewall-DefaultStatelessRuleGroup-1J3Z3W2ZQXV3"),
			},
		},
	}, nil
}

func TestRuleGroupGetFunc(t *testing.T) {
	item, err := ruleGroupGetFunc(context.Background(), testNetworkFirewallClient{}, "test", &networkfirewall.DescribeRuleGroupInput{})

	if err != nil {
		t.Fatal(err)
	}

	if err := item.Validate(); err != nil {
		t.Fatal(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "sns-topic",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:sns:us-east-1:123456789012:aws-network-firewall-DefaultStatelessRuleGroup-1J3Z3W2ZQXV3",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "network-firewall-rule-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:network-firewall:us-east-1:123456789012:firewall/aws-network-firewall-DefaultFirewall-1J3Z3W2ZQXV3",
			ExpectedScope:  "123456789012.us-east-1",
		},
	}

	tests.Execute(t, item)
}
