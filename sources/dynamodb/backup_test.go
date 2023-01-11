package dynamodb

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func (t *TestClient) DescribeBackup(ctx context.Context, params *dynamodb.DescribeBackupInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeBackupOutput, error) {
	return &dynamodb.DescribeBackupOutput{
		BackupDescription: &types.BackupDescription{
			BackupDetails: &types.BackupDetails{
				BackupArn:              sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2/backup/01673461724486-a6007753"),
				BackupName:             sources.PtrString("test2-backup"),
				BackupSizeBytes:        sources.PtrInt64(0),
				BackupStatus:           types.BackupStatusAvailable,
				BackupType:             types.BackupTypeUser,
				BackupCreationDateTime: sources.PtrTime(time.Now()),
			},
			SourceTableDetails: &types.SourceTableDetails{
				TableName:      sources.PtrString("test2"), // link
				TableId:        sources.PtrString("12670f3b-8ca1-463b-b15e-f2e27eaf70b0"),
				TableArn:       sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2"),
				TableSizeBytes: sources.PtrInt64(0),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: sources.PtrString("ArtistId"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: sources.PtrString("Concert"),
						KeyType:       types.KeyTypeRange,
					},
				},
				TableCreationDateTime: sources.PtrTime(time.Now()),
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  sources.PtrInt64(5),
					WriteCapacityUnits: sources.PtrInt64(5),
				},
				ItemCount:   sources.PtrInt64(0),
				BillingMode: types.BillingModeProvisioned,
			},
			SourceTableFeatureDetails: &types.SourceTableFeatureDetails{
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndexInfo{
					{
						IndexName: sources.PtrString("GSI"),
						KeySchema: []types.KeySchemaElement{
							{
								AttributeName: sources.PtrString("TicketSales"),
								KeyType:       types.KeyTypeHash,
							},
						},
						Projection: &types.Projection{
							ProjectionType: types.ProjectionTypeKeysOnly,
						},
						ProvisionedThroughput: &types.ProvisionedThroughput{
							ReadCapacityUnits:  sources.PtrInt64(5),
							WriteCapacityUnits: sources.PtrInt64(5),
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
				TableName:              sources.PtrString("test2"),
				TableId:                sources.PtrString("12670f3b-8ca1-463b-b15e-f2e27eaf70b0"),
				TableArn:               sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2"),
				BackupArn:              sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test2/backup/01673461724486-a6007753"),
				BackupName:             sources.PtrString("test2-backup"),
				BackupCreationDateTime: sources.PtrTime(time.Now()),
				BackupStatus:           types.BackupStatusAvailable,
				BackupType:             types.BackupTypeUser,
				BackupSizeBytes:        sources.PtrInt64(10),
			},
		},
	}, nil
}

func TestBackupGetFunc(t *testing.T) {
	item, err := BackupGetFunc(context.Background(), &TestClient{}, "foo", &dynamodb.DescribeBackupInput{})

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "dynamodb-table",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "test2",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}
