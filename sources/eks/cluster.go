package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func clusterGetFunc(ctx context.Context, client EKSClient, scope string, input *eks.DescribeClusterInput) (*sdp.Item, error) {
	output, err := client.DescribeCluster(ctx, input)

	if err != nil {
		return nil, err
	}

	if output.Cluster == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "cluster response was nil",
		}
	}

	cluster := output.Cluster

	attributes, err := sources.ToAttributesCase(cluster, "clientRequestToken")

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "eks-cluster",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
		LinkedItemQueries: []*sdp.LinkedItemQuery{
			{
				Query: &sdp.Query{
					// +overmind:link eks-addon
					Type:   "eks-addon",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cluster.Name,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// These are tightly linked
					In:  true,
					Out: true,
				},
			},
			{
				Query: &sdp.Query{
					// +overmind:link eks-fargate-profile
					Type:   "eks-fargate-profile",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cluster.Name,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// These are tightly linked
					In:  true,
					Out: true,
				},
			},
			{
				Query: &sdp.Query{
					// +overmind:link eks-nodegroup
					Type:   "eks-nodegroup",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cluster.Name,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// These are tightly linked
					In:  true,
					Out: true,
				},
			},
		},
	}

	switch cluster.Status {
	case types.ClusterStatusCreating:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.ClusterStatusActive:
		item.Health = sdp.Health_HEALTH_OK.Enum()
	case types.ClusterStatusDeleting:
		item.Health = sdp.Health_HEALTH_WARNING.Enum()
	case types.ClusterStatusFailed:
		item.Health = sdp.Health_HEALTH_ERROR.Enum()
	case types.ClusterStatusUpdating:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.ClusterStatusPending:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	}

	var a *sources.ARN

	if cluster.ConnectorConfig != nil {
		if cluster.ConnectorConfig.RoleArn != nil {
			if a, err = sources.ParseARN(*cluster.ConnectorConfig.RoleArn); err == nil {
				// +overmind:link iam-role
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "iam-role",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cluster.ConnectorConfig.RoleArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// The role can affect the cluster
						In: true,
						// The cluster can't affect the role
						Out: false,
					},
				})
			}
		}
	}

	for _, conf := range cluster.EncryptionConfig {
		if conf.Provider != nil {
			if conf.Provider.KeyArn != nil {
				if a, err = sources.ParseARN(*conf.Provider.KeyArn); err == nil {
					// +overmind:link kms-key
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "kms-key",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *conf.Provider.KeyArn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// The key can affect the cluster
							In: true,
							// The cluster can't affect the key
							Out: false,
						},
					})
				}
			}
		}
	}

	if cluster.Endpoint != nil {
		// +overmind:link http
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "http",
				Method: sdp.QueryMethod_GET,
				Query:  *cluster.Endpoint,
				Scope:  "global",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// HTTP should be linked bidirectionally
				In:  true,
				Out: true,
			},
		})
	}

	if cluster.ResourcesVpcConfig != nil {
		if cluster.ResourcesVpcConfig.ClusterSecurityGroupId != nil {
			// +overmind:link ec2-security-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  *cluster.ResourcesVpcConfig.ClusterSecurityGroupId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The SG can affect the cluster
					In: true,
					// The cluster can't affect the SG
					Out: false,
				},
			})
		}

		for _, id := range cluster.ResourcesVpcConfig.SecurityGroupIds {
			// +overmind:link ec2-security-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  id,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The SG can affect the cluster
					In: true,
					// The cluster can't affect the SG
					Out: false,
				},
			})
		}

		for _, id := range cluster.ResourcesVpcConfig.SubnetIds {
			// +overmind:link ec2-subnet
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  id,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The subnet can affect the cluster
					In: true,
					// The cluster can't affect the subnet
					Out: false,
				},
			})
		}

		if cluster.ResourcesVpcConfig.VpcId != nil {
			// +overmind:link ec2-vpc
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *cluster.ResourcesVpcConfig.VpcId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The VPC can affect the cluster
					In: true,
					// The cluster can't affect the VPC
					Out: false,
				},
			})
		}
	}

	if cluster.RoleArn != nil {
		if a, err = sources.ParseARN(*cluster.RoleArn); err == nil {
			// +overmind:link iam-role
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *cluster.RoleArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The role can affect the cluster
					In: true,
					// The cluster can't affect the role
					Out: false,
				},
			})
		}
	}

	return &item, nil

}

//go:generate docgen ../../docs-data
// +overmind:type eks-cluster
// +overmind:descriptiveType EKS Cluster
// +overmind:get Get a cluster by name
// +overmind:list List all clusters
// +overmind:search Search for clusters by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_eks_cluster.name

func NewClusterSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*eks.ListClustersInput, *eks.ListClustersOutput, *eks.DescribeClusterInput, *eks.DescribeClusterOutput, EKSClient, *eks.Options] {
	return &sources.AlwaysGetSource[*eks.ListClustersInput, *eks.ListClustersOutput, *eks.DescribeClusterInput, *eks.DescribeClusterOutput, EKSClient, *eks.Options]{
		ItemType:  "eks-cluster",
		Client:    eks.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &eks.ListClustersInput{},
		GetInputMapper: func(scope, query string) *eks.DescribeClusterInput {
			return &eks.DescribeClusterInput{
				Name: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client EKSClient, input *eks.ListClustersInput) sources.Paginator[*eks.ListClustersOutput, *eks.Options] {
			return eks.NewListClustersPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *eks.ListClustersOutput, _ *eks.ListClustersInput) ([]*eks.DescribeClusterInput, error) {
			inputs := make([]*eks.DescribeClusterInput, len(output.Clusters))

			for i := range output.Clusters {
				inputs[i] = &eks.DescribeClusterInput{
					Name: &output.Clusters[i],
				}
			}

			return inputs, nil
		},
		GetFunc: clusterGetFunc,
	}
}
