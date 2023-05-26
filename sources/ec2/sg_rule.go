package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func securityGroupRuleInputMapperGet(scope string, query string) (*ec2.DescribeSecurityGroupRulesInput, error) {
	return &ec2.DescribeSecurityGroupRulesInput{
		SecurityGroupRuleIds: []string{
			query,
		},
	}, nil
}

func securityGroupRuleInputMapperList(scope string) (*ec2.DescribeSecurityGroupRulesInput, error) {
	return &ec2.DescribeSecurityGroupRulesInput{}, nil
}

func securityGroupRuleOutputMapper(scope string, _ *ec2.DescribeSecurityGroupRulesInput, output *ec2.DescribeSecurityGroupRulesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, securityGroupRule := range output.SecurityGroupRules {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(securityGroupRule)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-security-group-rule",
			UniqueAttribute: "securityGroupRuleId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if securityGroupRule.GroupId != nil {
			// +overmind:link ec2-security-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  *securityGroupRule.GroupId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// These are tightly linked
					In:  true,
					Out: true,
				},
			})
		}

		if rg := securityGroupRule.ReferencedGroupInfo; rg != nil {
			if rg.GroupId != nil {
				// +overmind:link ec2-security-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-security-group",
						Method: sdp.QueryMethod_GET,
						Query:  *rg.GroupId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// These are tightly linked
						In:  true,
						Out: true,
					},
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-security-group-rule
// +overmind:descriptiveType Security Group Rule
// +overmind:get Get a security group rule by ID
// +overmind:list List all security group rules
// +overmind:search Search security group rules by ARN
// +overmind:group AWS

func NewSecurityGroupRuleSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeSecurityGroupRulesInput, *ec2.DescribeSecurityGroupRulesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeSecurityGroupRulesInput, *ec2.DescribeSecurityGroupRulesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-security-group-rule",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeSecurityGroupRulesInput) (*ec2.DescribeSecurityGroupRulesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeSecurityGroupRules(ctx, input)
		},
		InputMapperGet:  securityGroupRuleInputMapperGet,
		InputMapperList: securityGroupRuleInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeSecurityGroupRulesInput) sources.Paginator[*ec2.DescribeSecurityGroupRulesOutput, *ec2.Options] {
			return ec2.NewDescribeSecurityGroupRulesPaginator(client, params)
		},
		OutputMapper: securityGroupRuleOutputMapper,
	}
}
