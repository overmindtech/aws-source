package rds

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestDBInstanceOutputMapper(t *testing.T) {
	output := &rds.DescribeDBInstancesOutput{
		DBInstances: []types.DBInstance{
			{
				DBInstanceIdentifier: adapters.PtrString("database-1-instance-1"),
				DBInstanceClass:      adapters.PtrString("db.r6g.large"),
				Engine:               adapters.PtrString("aurora-mysql"),
				DBInstanceStatus:     adapters.PtrString("available"),
				MasterUsername:       adapters.PtrString("admin"),
				Endpoint: &types.Endpoint{
					Address:      adapters.PtrString("database-1-instance-1.camcztjohmlj.eu-west-2.rds.amazonaws.com"), // link
					Port:         adapters.PtrInt32(3306),                                                              // link
					HostedZoneId: adapters.PtrString("Z1TTGA775OQIYO"),                                                 // link
				},
				AllocatedStorage:      adapters.PtrInt32(1),
				InstanceCreateTime:    adapters.PtrTime(time.Now()),
				PreferredBackupWindow: adapters.PtrString("00:05-00:35"),
				BackupRetentionPeriod: adapters.PtrInt32(1),
				DBSecurityGroups: []types.DBSecurityGroupMembership{
					{
						DBSecurityGroupName: adapters.PtrString("name"), // This is EC2Classic only so we're skipping this
					},
				},
				VpcSecurityGroups: []types.VpcSecurityGroupMembership{
					{
						VpcSecurityGroupId: adapters.PtrString("sg-094e151c9fc5da181"), // link
						Status:             adapters.PtrString("active"),
					},
				},
				DBParameterGroups: []types.DBParameterGroupStatus{
					{
						DBParameterGroupName: adapters.PtrString("default.aurora-mysql8.0"), // link
						ParameterApplyStatus: adapters.PtrString("in-sync"),
					},
				},
				AvailabilityZone: adapters.PtrString("eu-west-2a"), // link
				DBSubnetGroup: &types.DBSubnetGroup{
					DBSubnetGroupName:        adapters.PtrString("default-vpc-0d7892e00e573e701"), // link
					DBSubnetGroupDescription: adapters.PtrString("Created from the RDS Management Console"),
					VpcId:                    adapters.PtrString("vpc-0d7892e00e573e701"), // link
					SubnetGroupStatus:        adapters.PtrString("Complete"),
					Subnets: []types.Subnet{
						{
							SubnetIdentifier: adapters.PtrString("subnet-0d8ae4b4e07647efa"), // lnk
							SubnetAvailabilityZone: &types.AvailabilityZone{
								Name: adapters.PtrString("eu-west-2b"),
							},
							SubnetOutpost: &types.Outpost{
								Arn: adapters.PtrString("arn:aws:service:region:account:type/id"), // link
							},
							SubnetStatus: adapters.PtrString("Active"),
						},
					},
				},
				PreferredMaintenanceWindow: adapters.PtrString("fri:04:49-fri:05:19"),
				PendingModifiedValues:      &types.PendingModifiedValues{},
				MultiAZ:                    adapters.PtrBool(false),
				EngineVersion:              adapters.PtrString("8.0.mysql_aurora.3.02.0"),
				AutoMinorVersionUpgrade:    adapters.PtrBool(true),
				ReadReplicaDBInstanceIdentifiers: []string{
					"read",
				},
				LicenseModel: adapters.PtrString("general-public-license"),
				OptionGroupMemberships: []types.OptionGroupMembership{
					{
						OptionGroupName: adapters.PtrString("default:aurora-mysql-8-0"),
						Status:          adapters.PtrString("in-sync"),
					},
				},
				PubliclyAccessible:      adapters.PtrBool(false),
				StorageType:             adapters.PtrString("aurora"),
				DbInstancePort:          adapters.PtrInt32(0),
				DBClusterIdentifier:     adapters.PtrString("database-1"), // link
				StorageEncrypted:        adapters.PtrBool(true),
				KmsKeyId:                adapters.PtrString("arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933"), // link
				DbiResourceId:           adapters.PtrString("db-ET7CE5D5TQTK7MXNJGJNFQD52E"),
				CACertificateIdentifier: adapters.PtrString("rds-ca-2019"),
				DomainMemberships: []types.DomainMembership{
					{
						Domain:      adapters.PtrString("domain"),
						FQDN:        adapters.PtrString("fqdn"),
						IAMRoleName: adapters.PtrString("role"),
						Status:      adapters.PtrString("enrolled"),
					},
				},
				CopyTagsToSnapshot:                 adapters.PtrBool(false),
				MonitoringInterval:                 adapters.PtrInt32(60),
				EnhancedMonitoringResourceArn:      adapters.PtrString("arn:aws:logs:eu-west-2:052392120703:log-group:RDSOSMetrics:log-stream:db-ET7CE5D5TQTK7MXNJGJNFQD52E"), // link
				MonitoringRoleArn:                  adapters.PtrString("arn:aws:iam::052392120703:role/rds-monitoring-role"),                                                  // link
				PromotionTier:                      adapters.PtrInt32(1),
				DBInstanceArn:                      adapters.PtrString("arn:aws:rds:eu-west-2:052392120703:db:database-1-instance-1"),
				IAMDatabaseAuthenticationEnabled:   adapters.PtrBool(false),
				PerformanceInsightsEnabled:         adapters.PtrBool(true),
				PerformanceInsightsKMSKeyId:        adapters.PtrString("arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933"), // link
				PerformanceInsightsRetentionPeriod: adapters.PtrInt32(7),
				DeletionProtection:                 adapters.PtrBool(false),
				AssociatedRoles: []types.DBInstanceRole{
					{
						FeatureName: adapters.PtrString("something"),
						RoleArn:     adapters.PtrString("arn:aws:service:region:account:type/id"), // link
						Status:      adapters.PtrString("associated"),
					},
				},
				TagList:                []types.Tag{},
				CustomerOwnedIpEnabled: adapters.PtrBool(false),
				BackupTarget:           adapters.PtrString("region"),
				NetworkType:            adapters.PtrString("IPV4"),
				StorageThroughput:      adapters.PtrInt32(0),
				ActivityStreamEngineNativeAuditFieldsIncluded: adapters.PtrBool(true),
				ActivityStreamKinesisStreamName:               adapters.PtrString("aws-rds-das-db-AB1CDEFG23GHIJK4LMNOPQRST"), // link
				ActivityStreamKmsKeyId:                        adapters.PtrString("ab12345e-1111-2bc3-12a3-ab1cd12345e"),      // Not linking at the moment because there are too many possible formats. If you want to change this, submit a PR
				ActivityStreamMode:                            types.ActivityStreamModeAsync,
				ActivityStreamPolicyStatus:                    types.ActivityStreamPolicyStatusLocked,
				ActivityStreamStatus:                          types.ActivityStreamStatusStarted,
				AutomaticRestartTime:                          adapters.PtrTime(time.Now()),
				AutomationMode:                                types.AutomationModeAllPaused,
				AwsBackupRecoveryPointArn:                     adapters.PtrString("arn:aws:service:region:account:type/id"), // link
				CertificateDetails: &types.CertificateDetails{
					CAIdentifier: adapters.PtrString("id"),
					ValidTill:    adapters.PtrTime(time.Now()),
				},
				CharacterSetName:         adapters.PtrString("something"),
				CustomIamInstanceProfile: adapters.PtrString("arn:aws:service:region:account:type/id"), // link?
				DBInstanceAutomatedBackupsReplications: []types.DBInstanceAutomatedBackupsReplication{
					{
						DBInstanceAutomatedBackupsArn: adapters.PtrString("arn:aws:service:region:account:type/id"), // link
					},
				},
				DBName:                       adapters.PtrString("name"),
				DBSystemId:                   adapters.PtrString("id"),
				EnabledCloudwatchLogsExports: []string{},
				Iops:                         adapters.PtrInt32(10),
				LatestRestorableTime:         adapters.PtrTime(time.Now()),
				ListenerEndpoint: &types.Endpoint{
					Address:      adapters.PtrString("foo.bar.com"), // link
					HostedZoneId: adapters.PtrString("id"),          // link
					Port:         adapters.PtrInt32(5432),           // link
				},
				MasterUserSecret: &types.MasterUserSecret{
					KmsKeyId:     adapters.PtrString("id"),                                     // link
					SecretArn:    adapters.PtrString("arn:aws:service:region:account:type/id"), // link
					SecretStatus: adapters.PtrString("okay"),
				},
				MaxAllocatedStorage:                   adapters.PtrInt32(10),
				NcharCharacterSetName:                 adapters.PtrString("english"),
				ProcessorFeatures:                     []types.ProcessorFeature{},
				ReadReplicaDBClusterIdentifiers:       []string{},
				ReadReplicaSourceDBInstanceIdentifier: adapters.PtrString("id"),
				ReplicaMode:                           types.ReplicaModeMounted,
				ResumeFullAutomationModeTime:          adapters.PtrTime(time.Now()),
				SecondaryAvailabilityZone:             adapters.PtrString("eu-west-1"), // link
				StatusInfos:                           []types.DBInstanceStatusInfo{},
				TdeCredentialArn:                      adapters.PtrString("arn:aws:service:region:account:type/id"), // I don't have a good example for this so skipping for now. PR if required
				Timezone:                              adapters.PtrString("GB"),
			},
		},
	}

	items, err := dBInstanceOutputMapper(context.Background(), mockRdsClient{}, "foo", nil, output)

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

	if item.GetTags()["key"] != "value" {
		t.Errorf("got %v, expected %v", item.GetTags()["key"], "value")
	}

	tests := adapters.QueryTests{
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
	client, account, region := GetAutoConfig(t)

	source := NewDBInstanceSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
