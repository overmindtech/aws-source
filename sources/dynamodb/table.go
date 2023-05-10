package dynamodb

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func tableGetFunc(ctx context.Context, client Client, scope string, input *dynamodb.DescribeTableInput) (*sdp.Item, error) {
	out, err := client.DescribeTable(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.Table == nil {
		return nil, errors.New("returned table is nil")
	}

	table := out.Table

	attributes, err := sources.ToAttributesCase(table)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "dynamodb-table",
		UniqueAttribute: "tableName",
		Scope:           scope,
		Attributes:      attributes,
	}

	var a *sources.ARN

	streamsOut, err := client.DescribeKinesisStreamingDestination(ctx, &dynamodb.DescribeKinesisStreamingDestinationInput{
		TableName: table.TableName,
	})

	if err == nil {
		for _, dest := range streamsOut.KinesisDataStreamDestinations {
			if dest.StreamArn != nil {
				if a, err = sources.ParseARN(*dest.StreamArn); err == nil {
					// +overmind:link kinesis-stream
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "kinesis-stream",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *dest.StreamArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					}})
				}
			}
		}
	}

	if table.RestoreSummary != nil {
		if table.RestoreSummary.SourceBackupArn != nil {
			if a, err = sources.ParseARN(*table.RestoreSummary.SourceBackupArn); err == nil {
				// +overmind:link backup-recovery-point
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "backup-recovery-point",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *table.RestoreSummary.SourceBackupArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}

		if table.RestoreSummary.SourceTableArn != nil {
			if a, err = sources.ParseARN(*table.RestoreSummary.SourceTableArn); err == nil {
				// +overmind:link dynamodb-table
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "dynamodb-table",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *table.RestoreSummary.SourceTableArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}
	}

	if table.SSEDescription != nil {
		if table.SSEDescription.KMSMasterKeyArn != nil {
			if a, err = sources.ParseARN(*table.SSEDescription.KMSMasterKeyArn); err == nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *table.SSEDescription.KMSMasterKeyArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}
	}

	for _, replica := range table.Replicas {
		if replica.RegionName != nil {
			// +overmind:link ec2-region
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
				Type:   "ec2-region",
				Method: sdp.QueryMethod_GET,
				Query:  *replica.RegionName,
				Scope:  scope,
			}})
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type dynamodb-table
// +overmind:descriptiveType DynamoDB Table
// +overmind:get Get a DynamoDB table by name
// +overmind:list List all DynamoDB tables
// +overmind:search Search for DynamoDB tables by ARN
// +overmind:group AWS

func NewTableSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*dynamodb.ListTablesInput, *dynamodb.ListTablesOutput, *dynamodb.DescribeTableInput, *dynamodb.DescribeTableOutput, Client, *dynamodb.Options] {
	return &sources.AlwaysGetSource[*dynamodb.ListTablesInput, *dynamodb.ListTablesOutput, *dynamodb.DescribeTableInput, *dynamodb.DescribeTableOutput, Client, *dynamodb.Options]{
		ItemType:  "dynamodb-table",
		Client:    dynamodb.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		GetFunc:   tableGetFunc,
		ListInput: &dynamodb.ListTablesInput{},
		GetInputMapper: func(scope, query string) *dynamodb.DescribeTableInput {
			return &dynamodb.DescribeTableInput{
				TableName: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client Client, input *dynamodb.ListTablesInput) sources.Paginator[*dynamodb.ListTablesOutput, *dynamodb.Options] {
			return dynamodb.NewListTablesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *dynamodb.ListTablesOutput, input *dynamodb.ListTablesInput) ([]*dynamodb.DescribeTableInput, error) {
			if output == nil {
				return nil, errors.New("cannot map nil output")
			}

			inputs := make([]*dynamodb.DescribeTableInput, len(output.TableNames))

			for i := range output.TableNames {
				inputs = append(inputs, &dynamodb.DescribeTableInput{
					TableName: &output.TableNames[i],
				})
			}

			return inputs, nil
		},
	}
}
