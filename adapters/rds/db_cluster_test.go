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

func TestDBClusterOutputMapper(t *testing.T) {
	output := rds.DescribeDBClustersOutput{
		DBClusters: []types.DBCluster{
			{
				AllocatedStorage: adapters.PtrInt32(100),
				AvailabilityZones: []string{
					"eu-west-2c", // link
				},
				BackupRetentionPeriod:      adapters.PtrInt32(7),
				DBClusterIdentifier:        adapters.PtrString("database-2"),
				DBClusterParameterGroup:    adapters.PtrString("default.postgres13"),
				DBSubnetGroup:              adapters.PtrString("default-vpc-0d7892e00e573e701"), // link
				Status:                     adapters.PtrString("available"),
				EarliestRestorableTime:     adapters.PtrTime(time.Now()),
				Endpoint:                   adapters.PtrString("database-2.cluster-camcztjohmlj.eu-west-2.rds.amazonaws.com"),    // link
				ReaderEndpoint:             adapters.PtrString("database-2.cluster-ro-camcztjohmlj.eu-west-2.rds.amazonaws.com"), // link
				MultiAZ:                    adapters.PtrBool(true),
				Engine:                     adapters.PtrString("postgres"),
				EngineVersion:              adapters.PtrString("13.7"),
				LatestRestorableTime:       adapters.PtrTime(time.Now()),
				Port:                       adapters.PtrInt32(5432), // link
				MasterUsername:             adapters.PtrString("postgres"),
				PreferredBackupWindow:      adapters.PtrString("04:48-05:18"),
				PreferredMaintenanceWindow: adapters.PtrString("fri:04:05-fri:04:35"),
				ReadReplicaIdentifiers: []string{
					"arn:aws:rds:eu-west-1:052392120703:cluster:read-replica", // link
				},
				DBClusterMembers: []types.DBClusterMember{
					{
						DBInstanceIdentifier:          adapters.PtrString("database-2-instance-3"), // link
						IsClusterWriter:               adapters.PtrBool(false),
						DBClusterParameterGroupStatus: adapters.PtrString("in-sync"),
						PromotionTier:                 adapters.PtrInt32(1),
					},
				},
				VpcSecurityGroups: []types.VpcSecurityGroupMembership{
					{
						VpcSecurityGroupId: adapters.PtrString("sg-094e151c9fc5da181"), // link
						Status:             adapters.PtrString("active"),
					},
				},
				HostedZoneId:                     adapters.PtrString("Z1TTGA775OQIYO"), // link
				StorageEncrypted:                 adapters.PtrBool(true),
				KmsKeyId:                         adapters.PtrString("arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933"), // link
				DbClusterResourceId:              adapters.PtrString("cluster-2EW4PDVN7F7V57CUJPYOEAA74M"),
				DBClusterArn:                     adapters.PtrString("arn:aws:rds:eu-west-2:052392120703:cluster:database-2"),
				IAMDatabaseAuthenticationEnabled: adapters.PtrBool(false),
				ClusterCreateTime:                adapters.PtrTime(time.Now()),
				EngineMode:                       adapters.PtrString("provisioned"),
				DeletionProtection:               adapters.PtrBool(false),
				HttpEndpointEnabled:              adapters.PtrBool(false),
				ActivityStreamStatus:             types.ActivityStreamStatusStopped,
				CopyTagsToSnapshot:               adapters.PtrBool(false),
				CrossAccountClone:                adapters.PtrBool(false),
				DomainMemberships:                []types.DomainMembership{},
				TagList:                          []types.Tag{},
				DBClusterInstanceClass:           adapters.PtrString("db.m5d.large"),
				StorageType:                      adapters.PtrString("io1"),
				Iops:                             adapters.PtrInt32(1000),
				PubliclyAccessible:               adapters.PtrBool(true),
				AutoMinorVersionUpgrade:          adapters.PtrBool(true),
				MonitoringInterval:               adapters.PtrInt32(0),
				PerformanceInsightsEnabled:       adapters.PtrBool(false),
				NetworkType:                      adapters.PtrString("IPV4"),
				ActivityStreamKinesisStreamName:  adapters.PtrString("aws-rds-das-db-AB1CDEFG23GHIJK4LMNOPQRST"), // link
				ActivityStreamKmsKeyId:           adapters.PtrString("ab12345e-1111-2bc3-12a3-ab1cd12345e"),      // Not linking at the moment because there are too many possible formats. If you want to change this, submit a PR
				ActivityStreamMode:               types.ActivityStreamModeAsync,
				AutomaticRestartTime:             adapters.PtrTime(time.Now()),
				AssociatedRoles:                  []types.DBClusterRole{}, // EC2 classic roles, ignore
				BacktrackConsumedChangeRecords:   adapters.PtrInt64(1),
				BacktrackWindow:                  adapters.PtrInt64(2),
				Capacity:                         adapters.PtrInt32(2),
				CharacterSetName:                 adapters.PtrString("english"),
				CloneGroupId:                     adapters.PtrString("id"),
				CustomEndpoints: []string{
					"endpoint1", // link dns
				},
				DBClusterOptionGroupMemberships: []types.DBClusterOptionGroupStatus{
					{
						DBClusterOptionGroupName: adapters.PtrString("optionGroupName"), // link
						Status:                   adapters.PtrString("good"),
					},
				},
				DBSystemId:            adapters.PtrString("systemId"),
				DatabaseName:          adapters.PtrString("databaseName"),
				EarliestBacktrackTime: adapters.PtrTime(time.Now()),
				EnabledCloudwatchLogsExports: []string{
					"logExport1",
				},
				GlobalWriteForwardingRequested: adapters.PtrBool(true),
				GlobalWriteForwardingStatus:    types.WriteForwardingStatusDisabled,
				MasterUserSecret: &types.MasterUserSecret{
					KmsKeyId:     adapters.PtrString("arn:aws:kms:eu-west-2:052392120703:key/something"), // link
					SecretArn:    adapters.PtrString("arn:aws:service:region:account:type/id"),           // link
					SecretStatus: adapters.PtrString("okay"),
				},
				MonitoringRoleArn:                  adapters.PtrString("arn:aws:service:region:account:type/id"), // link
				PendingModifiedValues:              &types.ClusterPendingModifiedValues{},
				PercentProgress:                    adapters.PtrString("99"),
				PerformanceInsightsKMSKeyId:        adapters.PtrString("arn:aws:service:region:account:type/id"), // link, assuming it's an ARN
				PerformanceInsightsRetentionPeriod: adapters.PtrInt32(99),
				ReplicationSourceIdentifier:        adapters.PtrString("arn:aws:rds:eu-west-2:052392120703:cluster:database-1"), // link
				ScalingConfigurationInfo: &types.ScalingConfigurationInfo{
					AutoPause:             adapters.PtrBool(true),
					MaxCapacity:           adapters.PtrInt32(10),
					MinCapacity:           adapters.PtrInt32(1),
					SecondsBeforeTimeout:  adapters.PtrInt32(10),
					SecondsUntilAutoPause: adapters.PtrInt32(10),
					TimeoutAction:         adapters.PtrString("error"),
				},
				ServerlessV2ScalingConfiguration: &types.ServerlessV2ScalingConfigurationInfo{
					MaxCapacity: adapters.PtrFloat64(10),
					MinCapacity: adapters.PtrFloat64(1),
				},
			},
		},
	}

	items, err := dBClusterOutputMapper(context.Background(), mockRdsClient{}, "foo", nil, &output)

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
		t.Errorf("expected tag key to be value, got %v", item.GetTags()["key"])
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "rds-db-subnet-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default-vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "database-2.cluster-ro-camcztjohmlj.eu-west-2.rds.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "database-2.cluster-camcztjohmlj.eu-west-2.rds.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "rds-db-cluster",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:rds:eu-west-1:052392120703:cluster:read-replica",
			ExpectedScope:  "052392120703.eu-west-1",
		},
		{
			ExpectedType:   "rds-db-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "database-2-instance-3",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-094e151c9fc5da181",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "route53-hosted-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "Z1TTGA775OQIYO",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933",
			ExpectedScope:  "052392120703.eu-west-2",
		},
		{
			ExpectedType:   "kinesis-stream",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "aws-rds-das-db-AB1CDEFG23GHIJK4LMNOPQRST",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "endpoint1",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "rds-option-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "optionGroupName",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kms:eu-west-2:052392120703:key/something",
			ExpectedScope:  "052392120703.eu-west-2",
		},
		{
			ExpectedType:   "secretsmanager-secret",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "rds-db-cluster",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:rds:eu-west-2:052392120703:cluster:database-1",
			ExpectedScope:  "052392120703.eu-west-2",
		},
	}

	tests.Execute(t, item)
}

func TestNewDBClusterAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewDBClusterAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
