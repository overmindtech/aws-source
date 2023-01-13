package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func DBParameterGroupOutputMapper(scope string, output *rds.DescribeDBParameterGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, pg := range output.DBParameterGroups {
		attributes, err := sources.ToAttributesCase(pg)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "rds-db-parameter-group",
			UniqueAttribute: "dBParameterGroupName",
			Attributes:      attributes,
			Scope:           scope,
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewDBParameterGroupSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*rds.DescribeDBParameterGroupsInput, *rds.DescribeDBParameterGroupsOutput, *rds.Client, *rds.Options] {
	return &sources.DescribeOnlySource[*rds.DescribeDBParameterGroupsInput, *rds.DescribeDBParameterGroupsOutput, *rds.Client, *rds.Options]{
		ItemType:  "rds-db-parameter-group",
		Config:    config,
		AccountID: accountID,
		Client:    rds.NewFromConfig(config),
		PaginatorBuilder: func(client *rds.Client, params *rds.DescribeDBParameterGroupsInput) sources.Paginator[*rds.DescribeDBParameterGroupsOutput, *rds.Options] {
			return rds.NewDescribeDBParameterGroupsPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client *rds.Client, input *rds.DescribeDBParameterGroupsInput) (*rds.DescribeDBParameterGroupsOutput, error) {
			return client.DescribeDBParameterGroups(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*rds.DescribeDBParameterGroupsInput, error) {
			return &rds.DescribeDBParameterGroupsInput{
				DBParameterGroupName: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*rds.DescribeDBParameterGroupsInput, error) {
			return &rds.DescribeDBParameterGroupsInput{}, nil
		},
		OutputMapper: DBParameterGroupOutputMapper,
	}
}
