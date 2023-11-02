package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type unifiedFirewall struct {
	Name   string
	Config *types.Firewall
	Status *types.FirewallStatus
}

func firewallGetFunc(ctx context.Context, client networkFirewallClient, scope string, input *networkfirewall.DescribeFirewallInput) (*sdp.Item, error) {
	response, err := client.DescribeFirewall(ctx, input)

	if err != nil {
		return nil, err
	}

	if response == nil || response.Firewall == nil || response.Firewall.FirewallName == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "Firewall was nil",
			Scope:       scope,
		}
	}

	uf := unifiedFirewall{
		Name:   *response.Firewall.FirewallName,
		Config: response.Firewall,
		Status: response.FirewallStatus,
	}

	attributes, err := sources.ToAttributesCase(uf)

	if err != nil {
		return nil, err
	}

	var health *sdp.Health

	if response.FirewallStatus != nil {
		switch response.FirewallStatus.Status {
		case types.FirewallStatusValueDeleting:
			health = sdp.Health_HEALTH_PENDING.Enum()
		case types.FirewallStatusValueProvisioning:
			health = sdp.Health_HEALTH_PENDING.Enum()
		case types.FirewallStatusValueReady:
			health = sdp.Health_HEALTH_OK.Enum()
		}
	}

	tags := make(map[string]string)

	for _, tag := range response.Firewall.Tags {
		tags[*tag.Key] = *tag.Value
	}

	item := sdp.Item{
		Type:            "network-firewall-firewall",
		UniqueAttribute: "name",
		Scope:           scope,
		Attributes:      attributes,
		Health:          health,
		Tags:            tags,
	}

	config := response.Firewall

	if config.FirewallPolicyArn != nil {
		if a, err := sources.ParseARN(*config.FirewallPolicyArn); err == nil {
			//+overmind:link network-firewall-firewall-policy
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "network-firewall-firewall-policy",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *config.FirewallPolicyArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Policy will affect the firewall but not the other way around
					In:  true,
					Out: false,
				},
			})
		}
	}

	for _, mapping := range config.SubnetMappings {
		if mapping.SubnetId != nil {
			//+overmind:link ec2-subnet
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  *mapping.SubnetId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to public subnets could affect the firewall
					In:  true,
					Out: false,
				},
			})
		}
	}

	if config.VpcId != nil {
		//+overmind:link ec2-vpc
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ec2-vpc",
				Method: sdp.QueryMethod_GET,
				Query:  *config.VpcId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the VPC could affect the firewall
				In:  true,
				Out: false,
			},
		})
	}

	if config.EncryptionConfiguration != nil && config.EncryptionConfiguration.KeyId != nil {
		// This can be an ARN or an ID if it's in the same account
		if a, err := sources.ParseARN(*config.EncryptionConfiguration.KeyId); err == nil {
			//+overmind:link kms-key
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *config.EncryptionConfiguration.KeyId,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		} else {
			//+overmind:link kms-key
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_GET,
					Query:  *config.EncryptionConfiguration.KeyId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}
	}

	for _, state := range response.FirewallStatus.SyncStates {
		if state.Attachment != nil && state.Attachment.SubnetId != nil {
			//+overmind:link ec2-subnet
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  *state.Attachment.SubnetId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to public subnets could affect the firewall
					In:  true,
					Out: false,
				},
			})
		}
	}

	return &item, nil
}

func NewFirewallSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*networkfirewall.ListFirewallsInput, *networkfirewall.ListFirewallsOutput, *networkfirewall.DescribeFirewallInput, *networkfirewall.DescribeFirewallOutput, networkFirewallClient, *networkfirewall.Options] {
	return &sources.AlwaysGetSource[*networkfirewall.ListFirewallsInput, *networkfirewall.ListFirewallsOutput, *networkfirewall.DescribeFirewallInput, *networkfirewall.DescribeFirewallOutput, networkFirewallClient, *networkfirewall.Options]{
		ItemType:  "network-firewall-firewall",
		Client:    networkfirewall.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &networkfirewall.ListFirewallsInput{},
		GetInputMapper: func(scope, query string) *networkfirewall.DescribeFirewallInput {
			return &networkfirewall.DescribeFirewallInput{
				FirewallName: aws.String(query),
			}
		},
		SearchGetInputMapper: func(scope, query string) (*networkfirewall.DescribeFirewallInput, error) {
			return &networkfirewall.DescribeFirewallInput{
				FirewallArn: &query,
			}, nil
		},
		ListFuncPaginatorBuilder: func(client networkFirewallClient, input *networkfirewall.ListFirewallsInput) sources.Paginator[*networkfirewall.ListFirewallsOutput, *networkfirewall.Options] {
			return networkfirewall.NewListFirewallsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *networkfirewall.ListFirewallsOutput, input *networkfirewall.ListFirewallsInput) ([]*networkfirewall.DescribeFirewallInput, error) {
			var inputs []*networkfirewall.DescribeFirewallInput

			for _, firewall := range output.Firewalls {
				inputs = append(inputs, &networkfirewall.DescribeFirewallInput{
					FirewallArn: firewall.FirewallArn,
				})
			}
			return inputs, nil
		},
		GetFunc: func(ctx context.Context, client networkFirewallClient, scope string, input *networkfirewall.DescribeFirewallInput) (*sdp.Item, error) {
			return firewallGetFunc(ctx, client, scope, input)
		},
	}
}
