package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func DBClusterOutputMapper(scope string, output *rds.DescribeDBClustersOutput) ([]*sdp.Item, error) {
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
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "rds-db-subnet-group",
				Method: sdp.RequestMethod_GET,
				Query:  *cluster.DBSubnetGroup,
				Scope:  scope,
			})
		}

		for _, endpoint := range []*string{cluster.Endpoint, cluster.ReaderEndpoint} {
			if endpoint != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "dns",
					Method: sdp.RequestMethod_GET,
					Query:  *endpoint,
					Scope:  "global",
				})

				if cluster.Port != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "networksocket",
						Method: sdp.RequestMethod_SEARCH,
						Query:  fmt.Sprintf("%v:%v", *endpoint, *cluster.Port),
						Scope:  "global",
					})
				}
			}
		}

		for _, replica := range cluster.ReadReplicaIdentifiers {
			if a, err = sources.ParseARN(replica); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "rds-db-cluster",
					Method: sdp.RequestMethod_SEARCH,
					Query:  replica,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, member := range cluster.DBClusterMembers {
			if member.DBInstanceIdentifier != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "rds-db-instance",
					Method: sdp.RequestMethod_GET,
					Query:  *member.DBInstanceIdentifier,
					Scope:  scope,
				})
			}
		}

		for _, sg := range cluster.VpcSecurityGroups {
			if sg.VpcSecurityGroupId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  *sg.VpcSecurityGroupId,
					Scope:  scope,
				})
			}
		}

		if cluster.HostedZoneId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "route53-hosted-zone",
				Method: sdp.RequestMethod_GET,
				Query:  *cluster.HostedZoneId,
				Scope:  scope,
			})
		}

		if cluster.KmsKeyId != nil {
			if a, err = sources.ParseARN(*cluster.KmsKeyId); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "kms-key",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *cluster.KmsKeyId,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if cluster.ActivityStreamKinesisStreamName != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "kinesis-stream",
				Method: sdp.RequestMethod_GET,
				Query:  *cluster.ActivityStreamKinesisStreamName,
				Scope:  scope,
			})
		}

		for _, endpoint := range cluster.CustomEndpoints {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "dns",
				Method: sdp.RequestMethod_GET,
				Query:  endpoint,
				Scope:  "global",
			})
		}

		for _, optionGroup := range cluster.DBClusterOptionGroupMemberships {
			if optionGroup.DBClusterOptionGroupName != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "rds-option-group",
					Method: sdp.RequestMethod_GET,
					Query:  *optionGroup.DBClusterOptionGroupName,
					Scope:  scope,
				})
			}
		}

		if cluster.MasterUserSecret != nil {
			if cluster.MasterUserSecret.KmsKeyId != nil {
				if a, err = sources.ParseARN(*cluster.MasterUserSecret.KmsKeyId); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "kms-key",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *cluster.MasterUserSecret.KmsKeyId,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}

			if cluster.MasterUserSecret.SecretArn != nil {
				if a, err = sources.ParseARN(*cluster.MasterUserSecret.SecretArn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "secretsmanager-secret",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *cluster.MasterUserSecret.SecretArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if cluster.MonitoringRoleArn != nil {
			if a, err = sources.ParseARN(*cluster.MonitoringRoleArn); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "iam-role",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *cluster.MonitoringRoleArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if cluster.PerformanceInsightsKMSKeyId != nil {
			// This is an ARN
			if a, err = sources.ParseARN(*cluster.PerformanceInsightsKMSKeyId); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "kms-key",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *cluster.PerformanceInsightsKMSKeyId,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if cluster.ReplicationSourceIdentifier != nil {
			if a, err = sources.ParseARN(*cluster.ReplicationSourceIdentifier); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "rds-db-cluster",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *cluster.ReplicationSourceIdentifier,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

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
		OutputMapper: DBClusterOutputMapper,
	}
}
