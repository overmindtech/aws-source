package ecs

import (
	"context"
	"fmt"

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

func clusterGetFunc(ctx context.Context, client ECSClient, scope string, input *ecs.DescribeClustersInput) (*sdp.Item, error) {
	out, err := client.DescribeClusters(ctx, input)

	if err != nil {
		return nil, err
	}

	accountID, _, err := sources.ParseScope(scope)

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

	attributes, err := sources.ToAttributesWithExclude(cluster, "tags")

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "ecs-cluster",
		UniqueAttribute: "ClusterName",
		Scope:           scope,
		Attributes:      attributes,
		Tags:            tagsToMap(cluster.Tags),
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
				Query: &sdp.Query{
					// +overmind:link ecs-container-instance
					// Search for all container instances on this cluster
					Type:   "ecs-container-instance",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cluster.ClusterName,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Container instances can affect the cluster
					In: true,
					// The cluster will definitely affect the container
					// instances
					Out: true,
				},
			},
			{
				Query: &sdp.Query{
					// +overmind:link ecs-service
					Type:   "ecs-service",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cluster.ClusterName,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Services won't affect the cluster
					In: false,
					// The cluster will definitely affect the services
					Out: true,
				},
			},
			{
				Query: &sdp.Query{
					// +overmind:link ecs-task
					Type:   "ecs-task",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cluster.ClusterName,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Tasks won't affect the cluster
					In: false,
					// The cluster will definitely affect the tasks
					Out: true,
				},
			},
		},
	}

	if cluster.Status != nil {
		switch *cluster.Status {
		case "ACTIVE":
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case "PROVISIONING":
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case "DEPROVISIONING":
			item.Health = sdp.Health_HEALTH_WARNING.Enum()
		case "FAILED":
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		case "INACTIVE":
			// This means it's a deleted cluster
			item.Health = nil
		}
	}

	if cluster.Configuration != nil {
		if cluster.Configuration.ExecuteCommandConfiguration != nil {
			if cluster.Configuration.ExecuteCommandConfiguration.KmsKeyId != nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "kms-key",
						Method: sdp.QueryMethod_GET,
						Query:  *cluster.Configuration.ExecuteCommandConfiguration.KmsKeyId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the KMS key will probably affect the cluster
						In: true,
						// The cluster won't affect the KMS key though
						Out: false,
					},
				})
			}

			if cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration != nil {
				if cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.CloudWatchLogGroupName != nil {
					// +overmind:link logs-log-group
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "logs-log-group",
							Method: sdp.QueryMethod_GET,
							Query:  *cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.CloudWatchLogGroupName,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// These are tightly linked
							In:  true,
							Out: true,
						},
					})
				}

				if cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.S3BucketName != nil {
					// +overmind:link s3-bucket
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "s3-bucket",
							Method: sdp.QueryMethod_GET,
							Query:  *cluster.Configuration.ExecuteCommandConfiguration.LogConfiguration.S3BucketName,
							Scope:  sources.FormatScope(accountID, ""),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// These are tightly linked
							In:  true,
							Out: true,
						},
					})
				}
			}
		}
	}

	for _, provider := range cluster.CapacityProviders {
		// +overmind:link ecs-capacity-provider
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ecs-capacity-provider",
				Method: sdp.QueryMethod_GET,
				Query:  provider,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// These are tightly linked
				In:  true,
				Out: true,
			},
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ecs-cluster
// +overmind:descriptiveType ECS Cluster
// +overmind:get Get a cluster by name
// +overmind:list List all clusters
// +overmind:search Search for a cluster by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_ecs_cluster.arn
// +overmind:terraform:method SEARCH

func NewClusterSource(client ECSClient, accountID string, region string) *sources.AlwaysGetSource[*ecs.ListClustersInput, *ecs.ListClustersOutput, *ecs.DescribeClustersInput, *ecs.DescribeClustersOutput, ECSClient, *ecs.Options] {
	return &sources.AlwaysGetSource[*ecs.ListClustersInput, *ecs.ListClustersOutput, *ecs.DescribeClustersInput, *ecs.DescribeClustersOutput, ECSClient, *ecs.Options]{
		ItemType:  "ecs-cluster",
		Client:    client,
		AccountID: accountID,
		Region:    region,
		GetFunc:   clusterGetFunc,
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
