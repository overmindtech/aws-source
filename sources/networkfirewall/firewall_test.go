package networkfirewall

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func (c testNetworkFirewallClient) DescribeFirewall(ctx context.Context, params *networkfirewall.DescribeFirewallInput, optFns ...func(*networkfirewall.Options)) (*networkfirewall.DescribeFirewallOutput, error) {
	return &networkfirewall.DescribeFirewallOutput{
		Firewall: &types.Firewall{
			FirewallId:        sources.PtrString("test"),
			FirewallPolicyArn: sources.PtrString("arn:aws:network-firewall:us-east-1:123456789012:stateless-rulegroup/aws-network-firewall-DefaultStatelessRuleGroup-1J3Z3W2ZQXV3"), // link
			SubnetMappings: []types.SubnetMapping{
				{
					SubnetId:      sources.PtrString("subnet-12345678901234567"), // link
					IPAddressType: types.IPAddressTypeIpv4,
				},
			},
			VpcId:            sources.PtrString("vpc-12345678901234567"), // link
			DeleteProtection: false,
			Description:      sources.PtrString("test"),
			EncryptionConfiguration: &types.EncryptionConfiguration{
				Type:  types.EncryptionTypeAwsOwnedKmsKey,
				KeyId: sources.PtrString("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"), // link (this can be an ARN or ID)
			},
			FirewallArn:                    sources.PtrString("arn:aws:network-firewall:us-east-1:123456789012:firewall/aws-network-firewall-DefaultFirewall-1J3Z3W2ZQXV3"),
			FirewallName:                   sources.PtrString("test"),
			FirewallPolicyChangeProtection: false,
			SubnetChangeProtection:         false,
			Tags: []types.Tag{
				{
					Key:   sources.PtrString("test"),
					Value: sources.PtrString("test"),
				},
			},
		},
		FirewallStatus: &types.FirewallStatus{
			ConfigurationSyncStateSummary: types.ConfigurationSyncStateInSync,
			Status:                        types.FirewallStatusValueDeleting,
			CapacityUsageSummary: &types.CapacityUsageSummary{
				CIDRs: &types.CIDRSummary{
					AvailableCIDRCount: sources.PtrInt32(1),
					IPSetReferences: map[string]types.IPSetMetadata{
						"test": {
							ResolvedCIDRCount: sources.PtrInt32(1),
						},
					},
					UtilizedCIDRCount: sources.PtrInt32(1),
				},
			},
			SyncStates: map[string]types.SyncState{
				"test": {
					Attachment: &types.Attachment{
						EndpointId:    sources.PtrString("test"),
						Status:        types.AttachmentStatusCreating,
						StatusMessage: sources.PtrString("test"),
						SubnetId:      sources.PtrString("test"), // link,
					},
				},
			},
		},
	}, nil
}

func (c testNetworkFirewallClient) ListFirewalls(context.Context, *networkfirewall.ListFirewallsInput, ...func(*networkfirewall.Options)) (*networkfirewall.ListFirewallsOutput, error) {
	return &networkfirewall.ListFirewallsOutput{
		Firewalls: []types.FirewallMetadata{
			{
				FirewallArn: sources.PtrString("arn:aws:network-firewall:us-east-1:123456789012:firewall/aws-network-firewall-DefaultFirewall-1J3Z3W2ZQXV3"),
			},
		},
	}, nil
}

func TestFirewallGetFunc(t *testing.T) {
	item, err := firewallGetFunc(context.Background(), testNetworkFirewallClient{}, "test", &networkfirewall.DescribeFirewallInput{})

	if err != nil {
		t.Fatal(err)
	}

	if err := item.Validate(); err != nil {
		t.Fatal(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-12345678901234567",
			ExpectedScope:  "test",
		},
		{
			ExpectedType:   "network-firewall-firewall-policy",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:network-firewall:us-east-1:123456789012:stateless-rulegroup/aws-network-firewall-DefaultStatelessRuleGroup-1J3Z3W2ZQXV3",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-12345678901234567",
			ExpectedScope:  "test",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test",
			ExpectedScope:  "test",
		},
	}

	tests.Execute(t, item)
}
