package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func statusToHealth(status string) *sdp.Health {
	switch status {
	case "Available":
		return sdp.Health_HEALTH_OK.Enum()
	case "Backing-up":
		return sdp.Health_HEALTH_OK.Enum()
	case "Configuring-enhanced-monitoring":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Configuring-iam-database-auth":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Configuring-log-exports":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Converting-to-vpc":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Creating":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Deleting":
		return sdp.Health_HEALTH_WARNING.Enum()
	case "Failed":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Inaccessible-encryption-credentials":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Inaccessible-encryption-credentials-recoverable":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Incompatible-network":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Incompatible-option-group":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Incompatible-parameters":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Incompatible-restore":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Maintenance":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Modifying":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Moving-to-vpc":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Rebooting":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Resetting-master-credentials":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Renaming":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Restore-error":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Starting":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Stopped":
		return nil
	case "Stopping":
		return sdp.Health_HEALTH_PENDING.Enum()
	case "Storage-full":
		return sdp.Health_HEALTH_ERROR.Enum()
	case "Storage-optimization":
		return sdp.Health_HEALTH_OK.Enum()
	case "Upgrading":
		return sdp.Health_HEALTH_PENDING.Enum()
	}

	return nil
}

func dBInstanceOutputMapper(scope string, _ *rds.DescribeDBInstancesInput, output *rds.DescribeDBInstancesOutput) ([]*sdp.Item, error) {
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

		if instance.DBInstanceStatus != nil {
			item.Health = statusToHealth(*instance.DBInstanceStatus)
		}

		var a *sources.ARN

		if instance.Endpoint != nil {
			if instance.Endpoint.Address != nil {
				// +overmind:link dns
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.Endpoint.Address,
					Scope:  "global",
				})

				if instance.Endpoint.Port != 0 {
					// +overmind:link networksocket
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "networksocket",
						Method: sdp.QueryMethod_SEARCH,
						Query:  fmt.Sprintf("%v:%v", *instance.Endpoint.Address, instance.Endpoint.Port),
						Scope:  "global",
					})
				}
			}

			if instance.Endpoint.HostedZoneId != nil {
				// +overmind:link route53-hosted-zone
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "route53-hosted-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.Endpoint.HostedZoneId,
					Scope:  scope,
				})
			}
		}

		for _, sg := range instance.VpcSecurityGroups {
			if sg.VpcSecurityGroupId != nil {
				// +overmind:link ec2-security-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  *sg.VpcSecurityGroupId,
					Scope:  scope,
				})
			}
		}

		for _, paramGroup := range instance.DBParameterGroups {
			if paramGroup.DBParameterGroupName != nil {
				// +overmind:link rds-db-parameter-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "rds-db-parameter-group",
					Method: sdp.QueryMethod_GET,
					Query:  *paramGroup.DBParameterGroupName,
					Scope:  scope,
				})
			}
		}

		if instance.AvailabilityZone != nil {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.QueryMethod_GET,
				Query:  *instance.AvailabilityZone,
				Scope:  scope,
			})
		}

		if dbSubnetGroup != nil {
			// +overmind:link rds-db-subnet-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "rds-db-subnet-group",
				Method: sdp.QueryMethod_GET,
				Query:  *dbSubnetGroup,
				Scope:  scope,
			})
		}

		if instance.DBClusterIdentifier != nil {
			// +overmind:link rds-db-cluster
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "rds-db-cluster",
				Method: sdp.QueryMethod_GET,
				Query:  *instance.DBClusterIdentifier,
				Scope:  scope,
			})
		}

		if instance.KmsKeyId != nil {
			// This actually uses the ARN not the id
			if a, err = sources.ParseARN(*instance.KmsKeyId); err == nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.KmsKeyId,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.EnhancedMonitoringResourceArn != nil {
			if a, err = sources.ParseARN(*instance.EnhancedMonitoringResourceArn); err == nil {
				// +overmind:link logs-log-stream
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "logs-log-stream",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.EnhancedMonitoringResourceArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.MonitoringRoleArn != nil {
			if a, err = sources.ParseARN(*instance.MonitoringRoleArn); err == nil {
				// +overmind:link iam-role
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.MonitoringRoleArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.PerformanceInsightsKMSKeyId != nil {
			// This is an ARN
			if a, err = sources.ParseARN(*instance.PerformanceInsightsKMSKeyId); err == nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.PerformanceInsightsKMSKeyId,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, role := range instance.AssociatedRoles {
			if role.RoleArn != nil {
				if a, err = sources.ParseARN(*role.RoleArn); err == nil {
					// +overmind:link iam-role
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "iam-role",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *role.RoleArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if instance.ActivityStreamKinesisStreamName != nil {
			// +overmind:link kinesis-stream
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "kinesis-stream",
				Method: sdp.QueryMethod_GET,
				Query:  *instance.ActivityStreamKinesisStreamName,
				Scope:  scope,
			})
		}

		if instance.AwsBackupRecoveryPointArn != nil {
			if a, err = sources.ParseARN(*instance.AwsBackupRecoveryPointArn); err == nil {
				// +overmind:link backup-recovery-point
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "backup-recovery-point",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.AwsBackupRecoveryPointArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if instance.CustomIamInstanceProfile != nil {
			// This is almost certainly an ARN since IAM basically always is
			if a, err = sources.ParseARN(*instance.CustomIamInstanceProfile); err == nil {
				// +overmind:link iam-instance-profile
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "iam-instance-profile",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.CustomIamInstanceProfile,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, replication := range instance.DBInstanceAutomatedBackupsReplications {
			if replication.DBInstanceAutomatedBackupsArn != nil {
				if a, err = sources.ParseARN(*replication.DBInstanceAutomatedBackupsArn); err == nil {
					// +overmind:link rds-db-instance-automated-backup
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "rds-db-instance-automated-backup",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *replication.DBInstanceAutomatedBackupsArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if instance.ListenerEndpoint != nil {
			if instance.ListenerEndpoint.Address != nil {
				// +overmind:link dns
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *instance.ListenerEndpoint.Address,
					Scope:  "global",
				})

				if instance.ListenerEndpoint.Port != 0 {
					// +overmind:link networksocket
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "networksocket",
						Method: sdp.QueryMethod_SEARCH,
						Query:  fmt.Sprintf("%v:%v", *instance.ListenerEndpoint.Address, instance.ListenerEndpoint.Port),
						Scope:  "global",
					})
				}
			}

			if instance.ListenerEndpoint.HostedZoneId != nil {
				// +overmind:link route53-hosted-zone
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "route53-hosted-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.ListenerEndpoint.HostedZoneId,
					Scope:  scope,
				})
			}
		}

		if instance.MasterUserSecret != nil {
			if instance.MasterUserSecret.KmsKeyId != nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.MasterUserSecret.KmsKeyId,
					Scope:  scope,
				})
			}

			if instance.MasterUserSecret.SecretArn != nil {
				if a, err = sources.ParseARN(*instance.MasterUserSecret.SecretArn); err == nil {
					// +overmind:link secretsmanager-secret
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "secretsmanager-secret",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *instance.MasterUserSecret.SecretArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if instance.SecondaryAvailabilityZone != nil {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-availability-zone",
				Method: sdp.QueryMethod_GET,
				Query:  *instance.SecondaryAvailabilityZone,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type rds-db-instance
// +overmind:descriptiveType RDS Instance
// +overmind:get Get an instance by ID
// +overmind:list List all instances
// +overmind:search Search for instances by ARN
// +overmind:group AWS

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
		OutputMapper: dBInstanceOutputMapper,
	}
}
