package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

type ParameterGroup struct {
	types.DBParameterGroup

	Parameters []types.Parameter
}

func dBParameterGroupItemMapper(_, scope string, awsItem *ParameterGroup) (*sdp.Item, error) {
	attributes, err := adapters.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "rds-db-parameter-group",
		UniqueAttribute: "DBParameterGroupName",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type rds-db-parameter-group
// +overmind:descriptiveType RDS Parameter Group
// +overmind:get Get a parameter group by name
// +overmind:list List all parameter groups
// +overmind:search Search for a parameter group by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_db_parameter_group.arn
// +overmind:terraform:method SEARCH

func NewDBParameterGroupAdapter(client rdsClient, accountID string, region string) *adapters.GetListAdapter[*ParameterGroup, rdsClient, *rds.Options] {
	return &adapters.GetListAdapter[*ParameterGroup, rdsClient, *rds.Options]{
		ItemType:        "rds-db-parameter-group",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		AdapterMetadata: DBParameterGroupMetadata(),
		GetFunc: func(ctx context.Context, client rdsClient, scope, query string) (*ParameterGroup, error) {
			out, err := client.DescribeDBParameterGroups(ctx, &rds.DescribeDBParameterGroupsInput{
				DBParameterGroupName: &query,
			})
			if err != nil {
				return nil, err
			}

			if len(out.DBParameterGroups) != 1 {
				return nil, fmt.Errorf("expected 1 group, got %v", len(out.DBParameterGroups))
			}

			paramsOut, err := client.DescribeDBParameters(ctx, &rds.DescribeDBParametersInput{
				DBParameterGroupName: out.DBParameterGroups[0].DBParameterGroupName,
			})
			if err != nil {
				return nil, err
			}

			return &ParameterGroup{
				Parameters:       paramsOut.Parameters,
				DBParameterGroup: out.DBParameterGroups[0],
			}, nil
		},
		ListFunc: func(ctx context.Context, client rdsClient, scope string) ([]*ParameterGroup, error) {
			out, err := client.DescribeDBParameterGroups(ctx, &rds.DescribeDBParameterGroupsInput{})
			if err != nil {
				return nil, err
			}

			groups := make([]*ParameterGroup, 0)

			for _, group := range out.DBParameterGroups {
				paramsOut, err := client.DescribeDBParameters(ctx, &rds.DescribeDBParametersInput{
					DBParameterGroupName: group.DBParameterGroupName,
				})
				if err != nil {
					return nil, err
				}

				groups = append(groups, &ParameterGroup{
					Parameters:       paramsOut.Parameters,
					DBParameterGroup: group,
				})
			}

			return groups, nil
		},
		ListTagsFunc: func(ctx context.Context, pg *ParameterGroup, c rdsClient) (map[string]string, error) {
			out, err := c.ListTagsForResource(ctx, &rds.ListTagsForResourceInput{
				ResourceName: pg.DBParameterGroupArn,
			})
			if err != nil {
				return nil, err
			}

			return tagsToMap(out.TagList), nil
		},
		ItemMapper: dBParameterGroupItemMapper,
	}
}

func DBParameterGroupMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "rds-db-parameter-group",
		DescriptiveName: "RDS Parameter Group",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a parameter group by name",
			ListDescription:   "List all parameter groups",
			SearchDescription: "Search for a parameter group by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{
				TerraformMethod:   sdp.QueryMethod_SEARCH,
				TerraformQueryMap: "aws_db_parameter_group.arn",
			},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_DATABASE,
	}
}
