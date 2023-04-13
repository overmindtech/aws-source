package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func optionGroupOutputMapper(scope string, _ *rds.DescribeOptionGroupsInput, output *rds.DescribeOptionGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, group := range output.OptionGroupsList {
		attributes, err := sources.ToAttributesCase(group)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "rds-option-group",
			UniqueAttribute: "optionGroupName",
			Attributes:      attributes,
			Scope:           scope,
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type rds-option-group
// +overmind:descriptiveType RDS Option Group
// +overmind:get Get an option group by name
// +overmind:list List all RDS option groups
// +overmind:search Search for an option group by ARN
// +overmind:group AWS

func NewOptionGroupSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*rds.DescribeOptionGroupsInput, *rds.DescribeOptionGroupsOutput, *rds.Client, *rds.Options] {
	return &sources.DescribeOnlySource[*rds.DescribeOptionGroupsInput, *rds.DescribeOptionGroupsOutput, *rds.Client, *rds.Options]{
		ItemType:  "rds-option-group",
		Config:    config,
		AccountID: accountID,
		Client:    rds.NewFromConfig(config),
		PaginatorBuilder: func(client *rds.Client, params *rds.DescribeOptionGroupsInput) sources.Paginator[*rds.DescribeOptionGroupsOutput, *rds.Options] {
			return rds.NewDescribeOptionGroupsPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client *rds.Client, input *rds.DescribeOptionGroupsInput) (*rds.DescribeOptionGroupsOutput, error) {
			return client.DescribeOptionGroups(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*rds.DescribeOptionGroupsInput, error) {
			return &rds.DescribeOptionGroupsInput{
				OptionGroupName: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*rds.DescribeOptionGroupsInput, error) {
			return &rds.DescribeOptionGroupsInput{}, nil
		},
		OutputMapper: optionGroupOutputMapper,
	}
}
