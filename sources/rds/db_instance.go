package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func DBInstanceOutputMapper(scope string, _ *rds.DescribeDBInstancesInput, output *rds.DescribeDBInstancesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, instance := range output.DBInstances {
		var dbSubnetGroup *string

		if instance.DBSubnetGroup != nil && instance.DBSubnetGroup.DBSubnetGroupName != nil {
			// Extract the subnet group so we can create a link
			dbSubnetGroup = instance.DBSubnetGroup.DBSubnetGroupName

			// Remove the data since this will come from a separate item
			instance.DBSubnetGroup = nil
		}

		attributes, err := sources.ToAttributesCase(instance)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "rds-db-instance",
			UniqueAttribute: "dBInstanceIdentifier",
			Attributes:      attributes,
			Scope:           scope,
		}

		var a *sources.ARN

		if instance.Endpoint != nil {
			if instance.Endpoint.Address != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "dns",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.Endpoint.Address,
					Scope:  "global",
				})

				if instance.Endpoint.Port != 0 {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "networksocket",
						Method: sdp.RequestMethod_SEARCH,
						Query:  fmt.Sprintf("%v:%v", *instance.Endpoint.Address, instance.Endpoint.Port),
						Scope:  "global",
					})
				}
			}

			if instance.Endpoint.HostedZoneId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "route53-hosted-zone",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.Endpoint.HostedZoneId,
					Scope:  scope,
				})
			}
		}

		for _, sg := range instance.VpcSecurityGroups {
			if sg.VpcSecurityGroupId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  *sg.VpcSecurityGroupId,
					Scope:  scope,
				})
			}
		}

		for _, paramGroup := range instance.DBParameterGroups {
			if paramGroup.DBParameterGroupName != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "rds-db-parameter-group",
					Method: sdp.RequestMethod_GET,
					Query:  *paramGroup.DBParameterGroupName,
					Scope:  scope,
				})
			}
		}

		if instance.AvailabilityZone != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.RequestMethod_GET,
				Query:  *instance.AvailabilityZone,
				Scope:  scope,
			})
		}

		if dbSubnetGroup != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "rds-db-subnet-group",
				Method: sdp.RequestMethod_GET,
				Query:  *dbSubnetGroup,
				Scope:  scope,
			})
		}

		if instance.DBClusterIdentifier != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "rds-db-cluster",
				Method: sdp.RequestMethod_GET,
				Query:  *instance.DBClusterIdentifier,
				Scope:  scope,
			})
		}

		if instance.KmsKeyId != nil {
			// This actually uses the ARN not the id
			if a, err = sources.ParseARN(*instance.KmsKeyId); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "kms-key",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *instance.KmsKeyId,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.EnhancedMonitoringResourceArn != nil {
			if a, err = sources.ParseARN(*instance.EnhancedMonitoringResourceArn); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "logs-log-stream",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *instance.EnhancedMonitoringResourceArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.MonitoringRoleArn != nil {
			if a, err = sources.ParseARN(*instance.MonitoringRoleArn); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "iam-role",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *instance.MonitoringRoleArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.PerformanceInsightsKMSKeyId != nil {
			// This is an ARN
			if a, err = sources.ParseARN(*instance.PerformanceInsightsKMSKeyId); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "kms-key",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *instance.PerformanceInsightsKMSKeyId,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, role := range instance.AssociatedRoles {
			if role.RoleArn != nil {
				if a, err = sources.ParseARN(*role.RoleArn); err == nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "iam-role",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *role.RoleArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if instance.ActivityStreamKinesisStreamName != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "kinesis-stream",
				Method: sdp.RequestMethod_GET,
				Query:  *instance.ActivityStreamKinesisStreamName,
				Scope:  scope,
			})
		}

		if instance.AwsBackupRecoveryPointArn != nil {
			if a, err = sources.ParseARN(*instance.AwsBackupRecoveryPointArn); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "backup-recovery-point",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *instance.AwsBackupRecoveryPointArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.CustomIamInstanceProfile != nil {
			// This is almost certainly an ARN since IAM basically always is
			if a, err = sources.ParseARN(*instance.CustomIamInstanceProfile); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "iam-instance-profile",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *instance.CustomIamInstanceProfile,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, replication := range instance.DBInstanceAutomatedBackupsReplications {
			if replication.DBInstanceAutomatedBackupsArn != nil {
				if a, err = sources.ParseARN(*replication.DBInstanceAutomatedBackupsArn); err == nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "rds-db-instance-automated-backup",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *replication.DBInstanceAutomatedBackupsArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if instance.ListenerEndpoint != nil {
			if instance.ListenerEndpoint.Address != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "dns",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.ListenerEndpoint.Address,
					Scope:  "global",
				})

				if instance.ListenerEndpoint.Port != 0 {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "networksocket",
						Method: sdp.RequestMethod_SEARCH,
						Query:  fmt.Sprintf("%v:%v", *instance.ListenerEndpoint.Address, instance.ListenerEndpoint.Port),
						Scope:  "global",
					})
				}
			}

			if instance.ListenerEndpoint.HostedZoneId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "route53-hosted-zone",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.ListenerEndpoint.HostedZoneId,
					Scope:  scope,
				})
			}
		}

		if instance.MasterUserSecret != nil {
			if instance.MasterUserSecret.KmsKeyId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "kms-key",
					Method: sdp.RequestMethod_GET,
					Query:  *instance.MasterUserSecret.KmsKeyId,
					Scope:  scope,
				})
			}

			if instance.MasterUserSecret.SecretArn != nil {
				if a, err = sources.ParseARN(*instance.MasterUserSecret.SecretArn); err == nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "secretsmanager-secret",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *instance.MasterUserSecret.SecretArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if instance.SecondaryAvailabilityZone != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.RequestMethod_GET,
				Query:  *instance.SecondaryAvailabilityZone,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewDBInstanceSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*rds.DescribeDBInstancesInput, *rds.DescribeDBInstancesOutput, *rds.Client, *rds.Options] {
	return &sources.DescribeOnlySource[*rds.DescribeDBInstancesInput, *rds.DescribeDBInstancesOutput, *rds.Client, *rds.Options]{
		ItemType:  "rds-db-instance",
		Config:    config,
		AccountID: accountID,
		Client:    rds.NewFromConfig(config),
		PaginatorBuilder: func(client *rds.Client, params *rds.DescribeDBInstancesInput) sources.Paginator[*rds.DescribeDBInstancesOutput, *rds.Options] {
			return rds.NewDescribeDBInstancesPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client *rds.Client, input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
			return client.DescribeDBInstances(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*rds.DescribeDBInstancesInput, error) {
			return &rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*rds.DescribeDBInstancesInput, error) {
			return &rds.DescribeDBInstancesInput{}, nil
		},
		OutputMapper: DBInstanceOutputMapper,
	}
}
