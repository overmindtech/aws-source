package dynamodb

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func (t *TestClient) DescribeBackup(ctx context.Context, params *dynamodb.DescribeBackupInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeBackupOutput, error) {
	return &dynamodb.DescribeBackupOutput{
		BackupDescription: &types.BackupDescription{
			BackupDetails: &types.BackupDetails{
				BackupArn:              adapters.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2/backup/01673461724486-a6007753"),
				BackupName:             adapters.PtrString("test2-backup"),
				BackupSizeBytes:        adapters.PtrInt64(0),
				BackupStatus:           types.BackupStatusAvailable,
				BackupType:             types.BackupTypeUser,
				BackupCreationDateTime: adapters.PtrTime(time.Now()),
			},
			SourceTableDetails: &types.SourceTableDetails{
				TableName:      adapters.PtrString("test2"), // link
				TableId:        adapters.PtrString("12670f3b-8ca1-463b-b15e-f2e27eaf70b0"),
				TableArn:       adapters.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2"),
				TableSizeBytes: adapters.PtrInt64(0),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: adapters.PtrString("ArtistId"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: adapters.PtrString("Concert"),
						KeyType:       types.KeyTypeRange,
					},
				},
				TableCreationDateTime: adapters.PtrTime(time.Now()),
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  adapters.PtrInt64(5),
					WriteCapacityUnits: adapters.PtrInt64(5),
				},
				ItemCount:   adapters.PtrInt64(0),
				BillingMode: types.BillingModeProvisioned,
			},
			SourceTableFeatureDetails: &types.SourceTableFeatureDetails{
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndexInfo{
					{
						IndexName: adapters.PtrString("GSI"),
						KeySchema: []types.KeySchemaElement{
							{
								AttributeName: adapters.PtrString("TicketSales"),
								KeyType:       types.KeyTypeHash,
							},
						},
						Projection: &types.Projection{
							ProjectionType: types.ProjectionTypeKeysOnly,
						},
						ProvisionedThroughput: &types.ProvisionedThroughput{
							ReadCapacityUnits:  adapters.PtrInt64(5),
							WriteCapacityUnits: adapters.PtrInt64(5),
						},
					},
				},
			},
		},
	}, nil
}

func (t *TestClient) ListBackups(ctx context.Context, params *dynamodb.ListBackupsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListBackupsOutput, error) {
	return &dynamodb.ListBackupsOutput{
		BackupSummaries: []types.BackupSummary{
			{
				TableName:              adapters.PtrString("test2"),
				TableId:                adapters.PtrString("12670f3b-8ca1-463b-b15e-f2e27eaf70b0"),
				TableArn:               adapters.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2"),
				BackupArn:              adapters.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2/backup/01673461724486-a6007753"),
				BackupName:             adapters.PtrString("test2-backup"),
				BackupCreationDateTime: adapters.PtrTime(time.Now()),
				BackupStatus:           types.BackupStatusAvailable,
				BackupType:             types.BackupTypeUser,
				BackupSizeBytes:        adapters.PtrInt64(10),
			},
		},
	}, nil
}

func TestBackupGetFunc(t *testing.T) {
	item, err := backupGetFunc(context.Background(), &TestClient{}, "foo", &dynamodb.DescribeBackupInput{})

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "dynamodb-table",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "test2",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewBackupAdapter(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := dynamodb.NewFromConfig(config)

	adapter := NewBackupAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
		SkipGet: true,
	}

	test.Run(t)
}
