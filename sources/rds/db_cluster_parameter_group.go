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

type ClusterParameterGroup struct {
	types.DBClusterParameterGroup

	Parameters []types.Parameter
}

func DBClusterParameterGroupItemMapper(scope string, awsItem *ClusterParameterGroup) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "rds-db-cluster-parameter-group",
		UniqueAttribute: "dBClusterParameterGroupName",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

func NewDBClusterParameterGroupSource(config aws.Config, accountID string, region string) *sources.GetListSource[*ClusterParameterGroup, *rds.Client, *rds.Options] {
	return &sources.GetListSource[*ClusterParameterGroup, *rds.Client, *rds.Options]{
		ItemType:  "rds-db-cluster-parameter-group",
		Client:    rds.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		GetFunc: func(ctx context.Context, client *rds.Client, scope, query string) (*ClusterParameterGroup, error) {
			out, err := client.DescribeDBClusterParameterGroups(ctx, &rds.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: &query,
			})

			if err != nil {
				return nil, err
			}

			if len(out.DBClusterParameterGroups) != 1 {
				return nil, fmt.Errorf("expected 1 group, got %v", len(out.DBClusterParameterGroups))
			}

			paramsOut, err := client.DescribeDBClusterParameters(ctx, &rds.DescribeDBClusterParametersInput{
				DBClusterParameterGroupName: out.DBClusterParameterGroups[0].DBClusterParameterGroupName,
			})

			if err != nil {
				return nil, err
			}

			return &ClusterParameterGroup{
				Parameters:              paramsOut.Parameters,
				DBClusterParameterGroup: out.DBClusterParameterGroups[0],
			}, nil
		},
		ListFunc: func(ctx context.Context, client *rds.Client, scope string) ([]*ClusterParameterGroup, error) {
			out, err := client.DescribeDBClusterParameterGroups(ctx, &rds.DescribeDBClusterParameterGroupsInput{})

			if err != nil {
				return nil, err
			}

			groups := make([]*ClusterParameterGroup, 0)

			for _, group := range out.DBClusterParameterGroups {
				paramsOut, err := client.DescribeDBClusterParameters(ctx, &rds.DescribeDBClusterParametersInput{
					DBClusterParameterGroupName: group.DBClusterParameterGroupName,
				})

				if err != nil {
					return nil, err
				}

				groups = append(groups, &ClusterParameterGroup{
					Parameters:              paramsOut.Parameters,
					DBClusterParameterGroup: group,
				})
			}

			return groups, nil
		},
		ItemMapper: DBClusterParameterGroupItemMapper,
	}
}
