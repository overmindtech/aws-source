package efs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func AccessPointOutputMapper(_ context.Context, _ *efs.Client, scope string, input *efs.DescribeAccessPointsInput, output *efs.DescribeAccessPointsOutput) ([]*sdp.Item, error) {
	if output == nil {
		return nil, errors.New("nil output from AWS")
	}

	items := make([]*sdp.Item, 0)

	for _, ap := range output.AccessPoints {
		attrs, err := sources.ToAttributesCase(ap)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "efs-access-point",
			UniqueAttribute: "accessPointId",
			Scope:           scope,
			Attributes:      attrs,
			Health:          lifeCycleStateToHealth(ap.LifeCycleState),
			Tags:            tagsToMap(ap.Tags),
		}

		if ap.FileSystemId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "efs-file-system",
					Method: sdp.QueryMethod_GET,
					Query:  *ap.FileSystemId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Access points are tightly coupled with filesystems
					In:  true,
					Out: true,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type efs-access-point
// +overmind:descriptiveType EFS Access Point
// +overmind:get Get an access point by ID
// +overmind:list List all access points
// +overmind:search Search for an access point by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_efs_access_point.id

func NewAccessPointSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*efs.DescribeAccessPointsInput, *efs.DescribeAccessPointsOutput, *efs.Client, *efs.Options] {
	return &sources.DescribeOnlySource[*efs.DescribeAccessPointsInput, *efs.DescribeAccessPointsOutput, *efs.Client, *efs.Options]{
		ItemType:  "efs-access-point",
		Config:    config,
		Client:    efs.NewFromConfig(config),
		AccountID: accountID,
		DescribeFunc: func(ctx context.Context, client *efs.Client, input *efs.DescribeAccessPointsInput) (*efs.DescribeAccessPointsOutput, error) {
			// Wait for rate limiting
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeAccessPoints(ctx, input)
		},
		PaginatorBuilder: func(client *efs.Client, params *efs.DescribeAccessPointsInput) sources.Paginator[*efs.DescribeAccessPointsOutput, *efs.Options] {
			return efs.NewDescribeAccessPointsPaginator(client, params)
		},
		InputMapperGet: func(scope, query string) (*efs.DescribeAccessPointsInput, error) {
			return &efs.DescribeAccessPointsInput{
				AccessPointId: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*efs.DescribeAccessPointsInput, error) {
			return &efs.DescribeAccessPointsInput{}, nil
		},
		OutputMapper: AccessPointOutputMapper,
	}
}
