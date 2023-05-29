package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func dBClusterOutputMapper(scope string, _ *rds.DescribeDBClustersInput, output *rds.DescribeDBClustersOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, cluster := range output.DBClusters {
		attributes, err := sources.ToAttributesCase(cluster)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "rds-db-cluster",
			UniqueAttribute: "dBClusterIdentifier",
			Attributes:      attributes,
			Scope:           scope,
		}

		var a *sources.ARN

		if cluster.DBSubnetGroup != nil {
			// +overmind:link rds-db-subnet-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "rds-db-subnet-group",
					Method: sdp.QueryMethod_GET,
					Query:  *cluster.DBSubnetGroup,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Tightly coupled
					In:  true,
					Out: false,
				},
			})
		}

		for _, endpoint := range []*string{cluster.Endpoint, cluster.ReaderEndpoint} {
			if endpoint != nil {
				// +overmind:link dns
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "dns",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *endpoint,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// DNS always linked
						In:  true,
						Out: true,
					},
				})
			}
		}

		for _, replica := range cluster.ReadReplicaIdentifiers {
			if a, err = sources.ParseARN(replica); err == nil {
				// +overmind:link rds-db-cluster
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "rds-db-cluster",
						Method: sdp.QueryMethod_SEARCH,
						Query:  replica,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}

		for _, az := range cluster.AvailabilityZones {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-availability-zone",
					Method: sdp.QueryMethod_GET,
					Query:  az,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the AZ can affect the cluster
					In: true,
					// The cluster won't affect the AZ
					Out: false,
				},
			})
		}

		for _, member := range cluster.DBClusterMembers {
			if member.DBInstanceIdentifier != nil {
				// +overmind:link rds-db-instance
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "rds-db-instance",
						Method: sdp.QueryMethod_GET,
						Query:  *member.DBInstanceIdentifier,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}

		for _, sg := range cluster.VpcSecurityGroups {
			if sg.VpcSecurityGroupId != nil {
				// +overmind:link ec2-security-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-security-group",
						Method: sdp.QueryMethod_GET,
						Query:  *sg.VpcSecurityGroupId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the security group can affect the cluster
						In: true,
						// The cluster won't affect the security group
						Out: false,
					},
				})
			}
		}

		if cluster.HostedZoneId != nil {
			// +overmind:link route53-hosted-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "route53-hosted-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *cluster.HostedZoneId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the hosted zone can affect the cluster
					In: true,
					// The cluster won't affect the hosted zone
					Out: false,
				},
			})
		}

		if cluster.KmsKeyId != nil {
			if a, err = sources.ParseARN(*cluster.KmsKeyId); err == nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "kms-key",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cluster.KmsKeyId,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the KMS key can affect the cluster
						In: true,
						// The cluster won't affect the KMS key
					},
				})
			}
		}

		if cluster.ActivityStreamKinesisStreamName != nil {
			// +overmind:link kinesis-stream
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "kinesis-stream",
					Method: sdp.QueryMethod_GET,
					Query:  *cluster.ActivityStreamKinesisStreamName,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the Kinesis stream can affect the cluster
					In: true,
					// Changes to the cluster can affect the Kinesis stream
					Out: true,
				},
			})
		}

		for _, endpoint := range cluster.CustomEndpoints {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  endpoint,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// DNS always linked
					In:  true,
					Out: true,
				},
			})
		}

		for _, optionGroup := range cluster.DBClusterOptionGroupMemberships {
			if optionGroup.DBClusterOptionGroupName != nil {
				// +overmind:link rds-option-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "rds-option-group",
						Method: sdp.QueryMethod_GET,
						Query:  *optionGroup.DBClusterOptionGroupName,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly linked
						In:  true,
						Out: true,
					},
				})
			}
		}

		if cluster.MasterUserSecret != nil {
			if cluster.MasterUserSecret.KmsKeyId != nil {
				if a, err = sources.ParseARN(*cluster.MasterUserSecret.KmsKeyId); err == nil {
					// +overmind:link kms-key
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "kms-key",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *cluster.MasterUserSecret.KmsKeyId,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes to the KMS key can affect the cluster
							In: true,
							// The cluster won't affect the KMS key
							Out: false,
						},
					})
				}
			}

			if cluster.MasterUserSecret.SecretArn != nil {
				if a, err = sources.ParseARN(*cluster.MasterUserSecret.SecretArn); err == nil {
					// +overmind:link secretsmanager-secret
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "secretsmanager-secret",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *cluster.MasterUserSecret.SecretArn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes to the secret can affect the cluster
							In: true,
							// The cluster won't affect the secret
							Out: false,
						},
					})
				}
			}
		}

		if cluster.MonitoringRoleArn != nil {
			if a, err = sources.ParseARN(*cluster.MonitoringRoleArn); err == nil {
				// +overmind:link iam-role
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "iam-role",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cluster.MonitoringRoleArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the IAM role can affect the cluster
						In: true,
						// The cluster won't affect the IAM role
						Out: false,
					},
				})
			}
		}

		if cluster.PerformanceInsightsKMSKeyId != nil {
			// This is an ARN
			if a, err = sources.ParseARN(*cluster.PerformanceInsightsKMSKeyId); err == nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "kms-key",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cluster.PerformanceInsightsKMSKeyId,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the KMS key can affect the cluster
						In: true,
						// The cluster won't affect the KMS key
						Out: false,
					},
				})
			}
		}

		if cluster.ReplicationSourceIdentifier != nil {
			if a, err = sources.ParseARN(*cluster.ReplicationSourceIdentifier); err == nil {
				// +overmind:link rds-db-cluster
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "rds-db-cluster",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cluster.ReplicationSourceIdentifier,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly coupled
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
// +overmind:type rds-db-cluster
// +overmind:descriptiveType RDS Cluster
// +overmind:get Get a cluster by ID
// +overmind:list List all clusters
// +overmind:search Search for clusters by ARN
// +overmind:group AWS

func NewDBClusterSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*rds.DescribeDBClustersInput, *rds.DescribeDBClustersOutput, *rds.Client, *rds.Options] {
	return &sources.DescribeOnlySource[*rds.DescribeDBClustersInput, *rds.DescribeDBClustersOutput, *rds.Client, *rds.Options]{
		ItemType:  "rds-db-cluster",
		Config:    config,
		AccountID: accountID,
		Client:    rds.NewFromConfig(config),
		PaginatorBuilder: func(client *rds.Client, params *rds.DescribeDBClustersInput) sources.Paginator[*rds.DescribeDBClustersOutput, *rds.Options] {
			return rds.NewDescribeDBClustersPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client *rds.Client, input *rds.DescribeDBClustersInput) (*rds.DescribeDBClustersOutput, error) {
			return client.DescribeDBClusters(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*rds.DescribeDBClustersInput, error) {
			return &rds.DescribeDBClustersInput{
				DBClusterIdentifier: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*rds.DescribeDBClustersInput, error) {
			return &rds.DescribeDBClustersInput{}, nil
		},
		OutputMapper: dBClusterOutputMapper,
	}
}
