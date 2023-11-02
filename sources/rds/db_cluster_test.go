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

func TestDBClusterOutputMapper(t *testing.T) {
	output := rds.DescribeDBClustersOutput{
		DBClusters: []types.DBCluster{
			{
				AllocatedStorage: sources.PtrInt32(100),
				AvailabilityZones: []string{
					"eu-west-2c", // link
				},
				BackupRetentionPeriod:      sources.PtrInt32(7),
				DBClusterIdentifier:        sources.PtrString("database-2"),
				DBClusterParameterGroup:    sources.PtrString("default.postgres13"),
				DBSubnetGroup:              sources.PtrString("default-vpc-0d7892e00e573e701"), // link
				Status:                     sources.PtrString("available"),
				EarliestRestorableTime:     sources.PtrTime(time.Now()),
				Endpoint:                   sources.PtrString("database-2.cluster-camcztjohmlj.eu-west-2.rds.amazonaws.com"),    // link
				ReaderEndpoint:             sources.PtrString("database-2.cluster-ro-camcztjohmlj.eu-west-2.rds.amazonaws.com"), // link
				MultiAZ:                    sources.PtrBool(true),
				Engine:                     sources.PtrString("postgres"),
				EngineVersion:              sources.PtrString("13.7"),
				LatestRestorableTime:       sources.PtrTime(time.Now()),
				Port:                       sources.PtrInt32(5432), // link
				MasterUsername:             sources.PtrString("postgres"),
				PreferredBackupWindow:      sources.PtrString("04:48-05:18"),
				PreferredMaintenanceWindow: sources.PtrString("fri:04:05-fri:04:35"),
				ReadReplicaIdentifiers: []string{
					"arn:aws:rds:eu-west-1:052392120703:cluster:read-replica", // link
				},
				DBClusterMembers: []types.DBClusterMember{
					{
						DBInstanceIdentifier:          sources.PtrString("database-2-instance-3"), // link
						IsClusterWriter:               sources.PtrBool(false),
						DBClusterParameterGroupStatus: sources.PtrString("in-sync"),
						PromotionTier:                 sources.PtrInt32(1),
					},
				},
				VpcSecurityGroups: []types.VpcSecurityGroupMembership{
					{
						VpcSecurityGroupId: sources.PtrString("sg-094e151c9fc5da181"), // link
						Status:             sources.PtrString("active"),
					},
				},
				HostedZoneId:                     sources.PtrString("Z1TTGA775OQIYO"), // link
				StorageEncrypted:                 sources.PtrBool(true),
				KmsKeyId:                         sources.PtrString("arn:aws:kms:eu-west-2:052392120703:key/9653cbdd-1590-464a-8456-67389cef6933"), // link
				DbClusterResourceId:              sources.PtrString("cluster-2EW4PDVN7F7V57CUJPYOEAA74M"),
				DBClusterArn:                     sources.PtrString("arn:aws:rds:eu-west-2:052392120703:cluster:database-2"),
				IAMDatabaseAuthenticationEnabled: sources.PtrBool(false),
				ClusterCreateTime:                sources.PtrTime(time.Now()),
				EngineMode:                       sources.PtrString("provisioned"),
				DeletionProtection:               sources.PtrBool(false),
				HttpEndpointEnabled:              sources.PtrBool(false),
				ActivityStreamStatus:             types.ActivityStreamStatusStopped,
				CopyTagsToSnapshot:               sources.PtrBool(false),
				CrossAccountClone:                sources.PtrBool(false),
				DomainMemberships:                []types.DomainMembership{},
				TagList:                          []types.Tag{},
				DBClusterInstanceClass:           sources.PtrString("db.m5d.large"),
				StorageType:                      sources.PtrString("io1"),
				Iops:                             sources.PtrInt32(1000),
				PubliclyAccessible:               sources.PtrBool(true),
				AutoMinorVersionUpgrade:          sources.PtrBool(true),
				MonitoringInterval:               sources.PtrInt32(0),
				PerformanceInsightsEnabled:       sources.PtrBool(false),
				NetworkType:                      sources.PtrString("IPV4"),
				ActivityStreamKinesisStreamName:  sources.PtrString("aws-rds-das-db-AB1CDEFG23GHIJK4LMNOPQRST"), // link
				ActivityStreamKmsKeyId:           sources.PtrString("ab12345e-1111-2bc3-12a3-ab1cd12345e"),      // Not linking at the moment because there are too many possible formats. If you want to change this, submit a PR
				ActivityStreamMode:               types.ActivityStreamModeAsync,
				AutomaticRestartTime:             sources.PtrTime(time.Now()),
				AssociatedRoles:                  []types.DBClusterRole{}, // EC2 classic roles, ignore
				BacktrackConsumedChangeRecords:   sources.PtrInt64(1),
				BacktrackWindow:                  sources.PtrInt64(2),
				Capacity:                         sources.PtrInt32(2),
				CharacterSetName:                 sources.PtrString("english"),
				CloneGroupId:                     sources.PtrString("id"),
				CustomEndpoints: []string{
					"endpoint1", // link dns
				},
				DBClusterOptionGroupMemberships: []types.DBClusterOptionGroupStatus{
					{
						DBClusterOptionGroupName: sources.PtrString("optionGroupName"), // link
						Status:                   sources.PtrString("good"),
					},
				},
				DBSystemId:            sources.PtrString("systemId"),
				DatabaseName:          sources.PtrString("databaseName"),
				EarliestBacktrackTime: sources.PtrTime(time.Now()),
				EnabledCloudwatchLogsExports: []string{
					"logExport1",
				},
				GlobalWriteForwardingRequested: sources.PtrBool(true),
				GlobalWriteForwardingStatus:    types.WriteForwardingStatusDisabled,
				MasterUserSecret: &types.MasterUserSecret{
					KmsKeyId:     sources.PtrString("arn:aws:kms:eu-west-2:052392120703:key/something"), // link
					SecretArn:    sources.PtrString("arn:aws:service:region:account:type/id"),           // link
					SecretStatus: sources.PtrString("okay"),
				},
				MonitoringRoleArn:                  sources.PtrString("arn:aws:service:region:account:type/id"), // link
				PendingModifiedValues:              &types.ClusterPendingModifiedValues{},
				PercentProgress:                    sources.PtrString("99"),
				PerformanceInsightsKMSKeyId:        sources.PtrString("arn:aws:service:region:account:type/id"), // link, assuming it's an ARN
				PerformanceInsightsRetentionPeriod: sources.PtrInt32(99),
				ReplicationSourceIdentifier:        sources.PtrString("arn:aws:rds:eu-west-2:052392120703:cluster:database-1"), // link
				ScalingConfigurationInfo: &types.ScalingConfigurationInfo{
					AutoPause:             sources.PtrBool(true),
					MaxCapacity:           sources.PtrInt32(10),
					MinCapacity:           sources.PtrInt32(1),
					SecondsBeforeTimeout:  sources.PtrInt32(10),
					SecondsUntilAutoPause: sources.PtrInt32(10),
					TimeoutAction:         sources.PtrString("error"),
				},
				ServerlessV2ScalingConfiguration: &types.ServerlessV2ScalingConfigurationInfo{
					MaxCapacity: sources.PtrFloat64(10),
					MinCapacity: sources.PtrFloat64(1),
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

	if item.Tags["key"] != "value" {
		t.Errorf("expected tag key to be value, got %v", item.Tags["key"])
	}

	tests := sources.QueryTests{
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

func TestNewDBClusterSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewDBClusterSource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
