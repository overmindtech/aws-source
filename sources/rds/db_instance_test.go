package rds

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestDBInstanceOutputMapper(t *testing.T) {
	output := &rds.DescribeDBInstancesOutput{
		DBInstances: []types.DBInstance{
			{
				DBInstanceIdentifier: sources.PtrString("database-1-instance-1"),
				DBInstanceClass:      sources.PtrString("db.r6g.large"),
				Engine:               sources.PtrString("aurora-mysql"),
				DBInstanceStatus:     sources.PtrString("available"),
				MasterUsername:       sources.PtrString("admin"),
				Endpoint: &types.Endpoint{
					Address:      sources.PtrString("database-1-instance-1.camcztjohmlj.eu-west-2.rds.amazonaws.com"), // link
					Port:         3306,                                                                                // link
					HostedZoneId: sources.PtrString("Z1TTGA775OQIYO"),                                                 // link
				},
				AllocatedStorage:      1,
				InstanceCreateTime:    sources.PtrTime(time.Now()),
				PreferredBackupWindow: sources.PtrString("00:05-00:35"),
				BackupRetentionPeriod: 1,
				DBSecurityGroups: []types.DBSecurityGroupMembership{
					{
						DBSecurityGroupName: sources.PtrString("name"), // This is EC2Classic only so we're skipping this
					},
				},
				VpcSecurityGroups: []types.VpcSecurityGroupMembership{
					{
						VpcSecurityGroupId: sources.PtrString("sg-094e151c9fc5da181"), // link
						Status:             sources.PtrString("active"),
					},
				},
				DBParameterGroups: []types.DBParameterGroupStatus{
					{
						DBParameterGroupName: sources.PtrString("default.aurora-mysql8.0"), // link
						ParameterApplyStatus: sources.PtrString("in-sync"),
					},
				},
				AvailabilityZone: sources.PtrString("eu-west-2a"), // link
				DBSubnetGroup: &types.DBSubnetGroup{
					DBSubnetGroupName:        sources.PtrString("default-vpc-0d7892e00e573e701"), // link
					DBSubnetGroupDescription: sources.PtrString("Created from the RDS Management Console"),
					VpcId:                    sources.PtrString("vpc-0d7892e00e573e701"), // link
					SubnetGroupStatus:        sources.PtrString("Complete"),
					Subnets: []types.Subnet{
						{
							SubnetIdentifier: sources.PtrString("subnet-0d8ae4b4e07647efa"), // lnk
							SubnetAvailabilityZone: &types.AvailabilityZone{
								Name: sources.PtrString("eu-west-2b"),
							},
							SubnetOutpost: &types.Outpost{
								Arn: sources.PtrString("arn:aws:service:region:account:type/id"), // link
							},
							SubnetStatus: sources.PtrString("Active"),
						},
					},
				},
				PreferredMaintenanceWindow: sources.PtrString("fri:04:49-fri:05:19"),
				PendingModifiedValues:      &types.PendingModifiedValues{},
				MultiAZ:                    false,
				EngineVersion:              sources.PtrString("8.0.mysql_aurora.3.02.0"),
				AutoMinorVersionUpgrade:    true,
				ReadReplicaDBInstanceIdentifiers: []string{
					"read",
				},
				LicenseModel: sources.PtrString("general-public-license"),
				OptionGroupMemberships: []types.OptionGroupMembership{
					{
						OptionGroupName: sources.PtrString("default:aurora-mysql-8-0"),
						Status:          sources.PtrString("in-sync"),
					},
				},
				PubliclyAccessible:      false,
				StorageType:             sources.PtrString("aurora"),
				DbInstancePort:          0,
				DBClusterIdentifier:     sources.PtrString("database-1"), // link
				StorageEncrypted:        true,
				KmsKeyId:                sources.PtrString("arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933"), // link
				DbiResourceId:           sources.PtrString("db-ET7CE5D5TQTK7MXNJGJNFQD52E"),
				CACertificateIdentifier: sources.PtrString("rds-ca-2019"),
				DomainMemberships: []types.DomainMembership{
					{
						Domain:      sources.PtrString("domain"),
						FQDN:        sources.PtrString("fqdn"),
						IAMRoleName: sources.PtrString("role"),
						Status:      sources.PtrString("enrolled"),
					},
				},
				CopyTagsToSnapshot:                 false,
				MonitoringInterval:                 sources.PtrInt32(60),
				EnhancedMonitoringResourceArn:      sources.PtrString("arn:aws:logs:eu-west-2:052392120703:log-group:RDSOSMetrics:log-stream:db-ET7CE5D5TQTK7MXNJGJNFQD52E"), // link
				MonitoringRoleArn:                  sources.PtrString("arn:aws:iam::052392120703:role/rds-monitoring-role"),                                                  // link
				PromotionTier:                      sources.PtrInt32(1),
				DBInstanceArn:                      sources.PtrString("arn:aws:rds:eu-west-2:052392120703:db:database-1-instance-1"),
				IAMDatabaseAuthenticationEnabled:   false,
				PerformanceInsightsEnabled:         sources.PtrBool(true),
				PerformanceInsightsKMSKeyId:        sources.PtrString("arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933"), // link
				PerformanceInsightsRetentionPeriod: sources.PtrInt32(7),
				DeletionProtection:                 false,
				AssociatedRoles: []types.DBInstanceRole{
					{
						FeatureName: sources.PtrString("something"),
						RoleArn:     sources.PtrString("arn:aws:service:region:account:type/id"), // link
						Status:      sources.PtrString("associated"),
					},
				},
				TagList:                []types.Tag{},
				CustomerOwnedIpEnabled: sources.PtrBool(false),
				BackupTarget:           sources.PtrString("region"),
				NetworkType:            sources.PtrString("IPV4"),
				StorageThroughput:      sources.PtrInt32(0),
				ActivityStreamEngineNativeAuditFieldsIncluded: sources.PtrBool(true),
				ActivityStreamKinesisStreamName:               sources.PtrString("aws-rds-das-db-AB1CDEFG23GHIJK4LMNOPQRST"), // link
				ActivityStreamKmsKeyId:                        sources.PtrString("ab12345e-1111-2bc3-12a3-ab1cd12345e"),      // Not linking at the moment because there are too many possible formats. If you want to change this, submit a PR
				ActivityStreamMode:                            types.ActivityStreamModeAsync,
				ActivityStreamPolicyStatus:                    types.ActivityStreamPolicyStatusLocked,
				ActivityStreamStatus:                          types.ActivityStreamStatusStarted,
				AutomaticRestartTime:                          sources.PtrTime(time.Now()),
				AutomationMode:                                types.AutomationModeAllPaused,
				AwsBackupRecoveryPointArn:                     sources.PtrString("arn:aws:service:region:account:type/id"), // link
				CertificateDetails: &types.CertificateDetails{
					CAIdentifier: sources.PtrString("id"),
					ValidTill:    sources.PtrTime(time.Now()),
				},
				CharacterSetName:         sources.PtrString("something"),
				CustomIamInstanceProfile: sources.PtrString("arn:aws:service:region:account:type/id"), // link?
				DBInstanceAutomatedBackupsReplications: []types.DBInstanceAutomatedBackupsReplication{
					{
						DBInstanceAutomatedBackupsArn: sources.PtrString("arn:aws:service:region:account:type/id"), // link
					},
				},
				DBName:                       sources.PtrString("name"),
				DBSystemId:                   sources.PtrString("id"),
				EnabledCloudwatchLogsExports: []string{},
				Iops:                         sources.PtrInt32(10),
				LatestRestorableTime:         sources.PtrTime(time.Now()),
				ListenerEndpoint: &types.Endpoint{
					Address:      sources.PtrString("foo.bar.com"), // link
					HostedZoneId: sources.PtrString("id"),          // link
					Port:         5432,                             // link
				},
				MasterUserSecret: &types.MasterUserSecret{
					KmsKeyId:     sources.PtrString("id"),                                     // link
					SecretArn:    sources.PtrString("arn:aws:service:region:account:type/id"), // link
					SecretStatus: sources.PtrString("okay"),
				},
				MaxAllocatedStorage:                   sources.PtrInt32(10),
				NcharCharacterSetName:                 sources.PtrString("english"),
				ProcessorFeatures:                     []types.ProcessorFeature{},
				ReadReplicaDBClusterIdentifiers:       []string{},
				ReadReplicaSourceDBInstanceIdentifier: sources.PtrString("id"),
				ReplicaMode:                           types.ReplicaModeMounted,
				ResumeFullAutomationModeTime:          sources.PtrTime(time.Now()),
				SecondaryAvailabilityZone:             sources.PtrString("eu-west-1"), // link
				StatusInfos:                           []types.DBInstanceStatusInfo{},
				TdeCredentialArn:                      sources.PtrString("arn:aws:service:region:account:type/id"), // I don't have a good example for this so skipping for now. PR if required
				Timezone:                              sources.PtrString("GB"),
			},
		},
	}

	items, err := dBInstanceOutputMapper(context.Background(), nil, "foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("got %v items, expected 1", len(items))
	}

	item := items[0]

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "database-1-instance-1.camcztjohmlj.eu-west-2.rds.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "route53-hosted-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "Z1TTGA775OQIYO",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-094e151c9fc5da181",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "rds-db-parameter-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default.aurora-mysql8.0",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-availability-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "eu-west-2a",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "rds-db-subnet-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default-vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "rds-db-cluster",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "database-1",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933",
			ExpectedScope:  "052392120703.eu-west-2",
		},
		{
			ExpectedType:   "logs-log-stream",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:logs:eu-west-2:052392120703:log-group:RDSOSMetrics:log-stream:db-ET7CE5D5TQTK7MXNJGJNFQD52E",
			ExpectedScope:  "052392120703.eu-west-2",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam::052392120703:role/rds-monitoring-role",
			ExpectedScope:  "052392120703",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933",
			ExpectedScope:  "052392120703.eu-west-2",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "kinesis-stream",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "aws-rds-das-db-AB1CDEFG23GHIJK4LMNOPQRST",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "backup-recovery-point",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "iam-instance-profile",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "rds-db-instance-automated-backup",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "foo.bar.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "route53-hosted-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "secretsmanager-secret",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
	}

	tests.Execute(t, item)
}

func TestNewDBInstanceSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewDBInstanceSource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
