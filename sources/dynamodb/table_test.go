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

func (t *TestClient) DescribeTable(context.Context, *dynamodb.DescribeTableInput, ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	return &dynamodb.DescribeTableOutput{
		Table: &types.TableDescription{
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: sources.PtrString("ArtistId"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: sources.PtrString("Concert"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: sources.PtrString("TicketSales"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			TableName: sources.PtrString("test-DDBTable-1X52D7BWAAB2H"),
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
			TableStatus:      types.TableStatusActive,
			CreationDateTime: sources.PtrTime(time.Now()),
			ProvisionedThroughput: &types.ProvisionedThroughputDescription{
				NumberOfDecreasesToday: sources.PtrInt64(0),
				ReadCapacityUnits:      sources.PtrInt64(5),
				WriteCapacityUnits:     sources.PtrInt64(5),
			},
			TableSizeBytes: sources.PtrInt64(0),
			ItemCount:      sources.PtrInt64(0),
			TableArn:       sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test-DDBTable-1X52D7BWAAB2H"),
			TableId:        sources.PtrString("32ef65bf-d6f3-4508-a3db-f201df09e437"),
			GlobalSecondaryIndexes: []types.GlobalSecondaryIndexDescription{
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
					IndexStatus: types.IndexStatusActive,
					ProvisionedThroughput: &types.ProvisionedThroughputDescription{
						NumberOfDecreasesToday: sources.PtrInt64(0),
						ReadCapacityUnits:      sources.PtrInt64(5),
						WriteCapacityUnits:     sources.PtrInt64(5),
					},
					IndexSizeBytes: sources.PtrInt64(0),
					ItemCount:      sources.PtrInt64(0),
					IndexArn:       sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test-DDBTable-1X52D7BWAAB2H/index/GSI"), // no link, t
				},
			},
			ArchivalSummary: &types.ArchivalSummary{
				ArchivalBackupArn: sources.PtrString("arn:aws:backups:eu-west-1:052392120703:some-backup/one"), // link
				ArchivalDateTime:  sources.PtrTime(time.Now()),
				ArchivalReason:    sources.PtrString("fear"),
			},
			BillingModeSummary: &types.BillingModeSummary{
				BillingMode: types.BillingModePayPerRequest,
			},
			GlobalTableVersion: sources.PtrString("1"),
			LatestStreamArn:    sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test-DDBTable-1X52D7BWAAB2H/stream/2023-01-11T16:53:02.371"), // This doesn't get linked because there is no more data to get
			LatestStreamLabel:  sources.PtrString("2023-01-11T16:53:02.371"),
			LocalSecondaryIndexes: []types.LocalSecondaryIndexDescription{
				{
					IndexArn:       sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test-DDBTable-1X52D7BWAAB2H/index/GSX"), // no link
					IndexName:      sources.PtrString("GSX"),
					IndexSizeBytes: sources.PtrInt64(29103),
					ItemCount:      sources.PtrInt64(234234),
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: sources.PtrString("TicketSales"),
							KeyType:       types.KeyTypeHash,
						},
					},
					Projection: &types.Projection{
						NonKeyAttributes: []string{
							"att1",
						},
						ProjectionType: types.ProjectionTypeInclude,
					},
				},
			},
			Replicas: []types.ReplicaDescription{
				{
					GlobalSecondaryIndexes: []types.ReplicaGlobalSecondaryIndexDescription{
						{
							IndexName: sources.PtrString("name"),
						},
					},
					KMSMasterKeyId: sources.PtrString("keyID"),
					RegionName:     sources.PtrString("eu-west-2"), // link
					ReplicaStatus:  types.ReplicaStatusActive,
					ReplicaTableClassSummary: &types.TableClassSummary{
						TableClass: types.TableClassStandard,
					},
				},
			},
			RestoreSummary: &types.RestoreSummary{
				RestoreDateTime:   sources.PtrTime(time.Now()),
				RestoreInProgress: sources.PtrBool(false),
				SourceBackupArn:   sources.PtrString("arn:aws:backup:eu-west-1:052392120703:recovery-point:89d0f956-d3a6-42fd-abbd-7d397766bc7e"), // link
				SourceTableArn:    sources.PtrString("arn:aws:dynamodb:eu-west-1:052392120703:table/test-DDBTable-1X52D7BWAAB2H"),                 // link
			},
			SSEDescription: &types.SSEDescription{
				InaccessibleEncryptionDateTime: sources.PtrTime(time.Now()),
				KMSMasterKeyArn:                sources.PtrString("arn:aws:service:region:account:type/id"), // link
				SSEType:                        types.SSETypeAes256,
				Status:                         types.SSEStatusDisabling,
			},
			StreamSpecification: &types.StreamSpecification{
				StreamEnabled:  sources.PtrBool(true),
				StreamViewType: types.StreamViewTypeKeysOnly,
			},
			TableClassSummary: &types.TableClassSummary{
				LastUpdateDateTime: sources.PtrTime(time.Now()),
				TableClass:         types.TableClassStandard,
			},
		},
	}, nil
}

func (t *TestClient) ListTables(context.Context, *dynamodb.ListTablesInput, ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
	return &dynamodb.ListTablesOutput{
		TableNames: []string{
			"test-DDBTable-1X52D7BWAAB2H",
		},
	}, nil
}

func (t *TestClient) DescribeKinesisStreamingDestination(ctx context.Context, params *dynamodb.DescribeKinesisStreamingDestinationInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeKinesisStreamingDestinationOutput, error) {
	return &dynamodb.DescribeKinesisStreamingDestinationOutput{
		KinesisDataStreamDestinations: []types.KinesisDataStreamDestination{
			{
				DestinationStatus:            types.DestinationStatusActive,
				DestinationStatusDescription: sources.PtrString("description"),
				StreamArn:                    sources.PtrString("arn:aws:kinesis:eu-west-1:052392120703:stream/test"),
			},
		},
	}, nil
}

func (t *TestClient) ListTagsOfResource(context.Context, *dynamodb.ListTagsOfResourceInput, ...func(*dynamodb.Options)) (*dynamodb.ListTagsOfResourceOutput, error) {
	return &dynamodb.ListTagsOfResourceOutput{
		Tags: []types.Tag{
			{
				Key:   sources.PtrString("key"),
				Value: sources.PtrString("value"),
			},
		},
		NextToken: nil,
	}, nil
}

func TestTableGetFunc(t *testing.T) {
	item, err := tableGetFunc(context.Background(), &TestClient{}, "foo", &dynamodb.DescribeTableInput{})

	if err != nil {
		t.Fatal(err)
	}

	if item.Tags["key"] != "value" {
		t.Errorf("expected tag key to be 'value', got '%s'", item.Tags["key"])
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "kinesis-stream",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kinesis:eu-west-1:052392120703:stream/test",
			ExpectedScope:  "052392120703.eu-west-1",
		},
		{
			ExpectedType:   "backup-recovery-point",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:backup:eu-west-1:052392120703:recovery-point:89d0f956-d3a6-42fd-abbd-7d397766bc7e",
			ExpectedScope:  "052392120703.eu-west-1",
		},
		{
			ExpectedType:   "dynamodb-table",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:dynamodb:eu-west-1:052392120703:table/test-DDBTable-1X52D7BWAAB2H",
			ExpectedScope:  "052392120703.eu-west-1",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
	}

	tests.Execute(t, item)
}

func TestNewTableSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewTableSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
