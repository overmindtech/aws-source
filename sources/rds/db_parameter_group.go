package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type ParameterGroup struct {
	types.DBParameterGroup

	Parameters []types.Parameter
}

func dBParameterGroupItemMapper(scope string, awsItem *ParameterGroup) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "rds-db-parameter-group",
		UniqueAttribute: "dBParameterGroupName",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

func NewDBParameterGroupSource(config aws.Config, accountID string, region string) *sources.GetListSource[*ParameterGroup, *rds.Client, *rds.Options] {
	return &sources.GetListSource[*ParameterGroup, *rds.Client, *rds.Options]{
		ItemType:  "rds-db-parameter-group",
		Client:    rds.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		GetFunc: func(ctx context.Context, client *rds.Client, scope, query string) (*ParameterGroup, error) {
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
		ListFunc: func(ctx context.Context, client *rds.Client, scope string) ([]*ParameterGroup, error) {
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
		ItemMapper: dBParameterGroupItemMapper,
	}
}
