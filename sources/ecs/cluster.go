package ecs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// ClusterIncludeFields Fields that we want included by default
var ClusterIncludeFields = []types.ClusterField{
	types.ClusterFieldAttachments,
	types.ClusterFieldConfigurations,
	types.ClusterFieldSettings,
	types.ClusterFieldStatistics,
	types.ClusterFieldTags,
}

func ClusterGetFunc(ctx context.Context, client ECSClient, scope string, input *ecs.DescribeClustersInput) (*sdp.Item, error) {
	out, err := client.DescribeClusters(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(out.Failures) != 0 {
		failure := out.Failures[0]

		if failure.Reason != nil && failure.Arn != nil {
			if *failure.Reason == "MISSING" {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: fmt.Sprintf("cluster with ARN %v not found", *failure.Arn),
				}
			}
		}

		return nil, fmt.Errorf("cluster get failure: %v", failure)
	}

	if len(out.Clusters) != 1 {
		return nil, fmt.Errorf("got %v clusters, expected 1", len(out.Clusters))
	}

	cluster := out.Clusters[0]

	attributes, err := sources.ToAttributesCase(cluster)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "ecs-cluster",
		UniqueAttribute: "clusterName",
		Scope:           scope,
		Attributes:      attributes,
		LinkedItemQueries: []*sdp.Query{
			{
				// Search for all container instances on this cluster
				Type:   "ecs-container-instance",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *cluster.ClusterName,
				Scope:  scope,
			},
			{
				Type:   "ecs-service",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *cluster.ClusterName,
				Scope:  scope,
			},
			{
				Type:   "ecs-task",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *cluster.ClusterName,
				Scope:  scope,
			},
		},
	}

	if cluster.Configuration != nil {
		if cluster.Configuration.ExecuteCommandConfiguration != nil {
			if cluster.Configuration.ExecuteCommandConfiguration.KmsKeyId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_GET,
					Query:  *cluster.Configuration.ExecuteCommandConfiguration.KmsKeyId,
					Scope:  scope,
				})
			}

			if cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration != nil {
				if cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.CloudWatchLogGroupName != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "logs-log-group",
						Method: sdp.QueryMethod_GET,
						Query:  *cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.CloudWatchLogGroupName,
						Scope:  scope,
					})
				}

				if cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.S3BucketName != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "s3-bucket",
						Method: sdp.QueryMethod_GET,
						Query:  *cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.S3BucketName,
						Scope:  scope,
					})
				}
			}
		}
	}

	for _, provider := range cluster.CapacityProviders {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "ecs-capacity-provider",
			Method: sdp.QueryMethod_GET,
			Query:  provider,
			Scope:  scope,
		})
	}

	return &item, nil
}

func NewClusterSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*ecs.ListClustersInput, *ecs.ListClustersOutput, *ecs.DescribeClustersInput, *ecs.DescribeClustersOutput, ECSClient, *ecs.Options] {
	return &sources.AlwaysGetSource[*ecs.ListClustersInput, *ecs.ListClustersOutput, *ecs.DescribeClustersInput, *ecs.DescribeClustersOutput, ECSClient, *ecs.Options]{
		ItemType:  "ecs-cluster",
		Client:    ecs.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		GetFunc:   ClusterGetFunc,
		GetInputMapper: func(scope, query string) *ecs.DescribeClustersInput {
			return &ecs.DescribeClustersInput{
				Clusters: []string{
					query,
				},
				Include: ClusterIncludeFields,
			}
		},
		ListInput: &ecs.ListClustersInput{},
		ListFuncPaginatorBuilder: func(client ECSClient, input *ecs.ListClustersInput) sources.Paginator[*ecs.ListClustersOutput, *ecs.Options] {
			return ecs.NewListClustersPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *ecs.ListClustersOutput, input *ecs.ListClustersInput) ([]*ecs.DescribeClustersInput, error) {
			inputs := make([]*ecs.DescribeClustersInput, 0)

			var a *sources.ARN
			var err error

			for _, arn := range output.ClusterArns {
				a, err = sources.ParseARN(arn)

				if err != nil {
					continue
				}

				inputs = append(inputs, &ecs.DescribeClustersInput{
					Clusters: []string{
						a.ResourceID(), // This will be the name of the cluster
					},
					Include: ClusterIncludeFields,
				})
			}

			return inputs, nil
		},
	}
}
