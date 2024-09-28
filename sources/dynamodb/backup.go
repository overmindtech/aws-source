package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func backupGetFunc(ctx context.Context, client Client, scope string, input *dynamodb.DescribeBackupInput) (*sdp.Item, error) {
	out, err := client.DescribeBackup(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.BackupDescription == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "backup description was nil",
		}
	}

	if out.BackupDescription.BackupDetails == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "backup details were nil",
		}
	}

	details := out.BackupDescription.BackupDetails

	attributes, err := sources.ToAttributesWithExclude(details)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "dynamodb-backup",
		UniqueAttribute: "BackupName",
		Attributes:      attributes,
		Scope:           scope,
	}

	if out.BackupDescription.SourceTableDetails != nil {
		if out.BackupDescription.SourceTableDetails.TableName != nil {
			// +overmind:link dynamodb-table
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dynamodb-table",
					Method: sdp.QueryMethod_GET,
					Query:  *out.BackupDescription.SourceTableDetails.TableName,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the table could probably affect the backup
					In: true,
					// Changing the backup won't exactly affect the table in
					// that it won't break it. But it could mean that it's no
					// longer backed up so, blast propagation should be here too
					Out: true,
				},
			})
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type dynamodb-backup
// +overmind:descriptiveType DynamoDB Backup
// +overmind:list List all DynamoDB backups
// +overmind:search Search for a DynamoDB backup by table name
// +overmind:group AWS

// NewBackupSource This source is a bit strange. This is the only thing I've
// found so far that can only be queries by ARN for Get. For this reason I'm
// going to just disable GET. LIST works fine and allows it to be linked to the
// table so this is enough for me at the moment
func NewBackupSource(client Client, accountID string, region string) *sources.AlwaysGetSource[*dynamodb.ListBackupsInput, *dynamodb.ListBackupsOutput, *dynamodb.DescribeBackupInput, *dynamodb.DescribeBackupOutput, Client, *dynamodb.Options] {
	return &sources.AlwaysGetSource[*dynamodb.ListBackupsInput, *dynamodb.ListBackupsOutput, *dynamodb.DescribeBackupInput, *dynamodb.DescribeBackupOutput, Client, *dynamodb.Options]{
		ItemType:  "dynamodb-backup",
		Client:    client,
		AccountID: accountID,
		Region:    region,
		GetFunc:   backupGetFunc,
		ListInput: &dynamodb.ListBackupsInput{},
		GetInputMapper: func(scope, query string) *dynamodb.DescribeBackupInput {
			// Get is not supported since you can't search by name
			return nil
		},
		ListFuncOutputMapper: func(output *dynamodb.ListBackupsOutput, input *dynamodb.ListBackupsInput) ([]*dynamodb.DescribeBackupInput, error) {
			inputs := make([]*dynamodb.DescribeBackupInput, 0)

			for _, summary := range output.BackupSummaries {
				if summary.BackupArn != nil {
					inputs = append(inputs, &dynamodb.DescribeBackupInput{
						BackupArn: summary.BackupArn,
					})
				}
			}

			return inputs, nil
		},
		ListFuncPaginatorBuilder: func(client Client, input *dynamodb.ListBackupsInput) sources.Paginator[*dynamodb.ListBackupsOutput, *dynamodb.Options] {
			return NewListBackupsPaginator(client, input)
		},
		SearchInputMapper: func(scope, query string) (*dynamodb.ListBackupsInput, error) {
			// Search by table name since you can't so it by ARN
			return &dynamodb.ListBackupsInput{
				TableName: &query,
			}, nil
		},
	}
}

// Another AWS API that doesn't provide a paginator *and* does pagination
// completely differently from everything else? You don't say.
//
// ░░░░░░░░░░░░░░▄▄▄▄▄▄▄▄▄▄▄▄░░░░░░░░░░░░░░
// ░░░░░░░░░░░░▄████████████████▄░░░░░░░░░░
// ░░░░░░░░░░▄██▀░░░░░░░▀▀████████▄░░░░░░░░
// ░░░░░░░░░▄█▀░░░░░░░░░░░░░▀▀██████▄░░░░░░
// ░░░░░░░░░███▄░░░░░░░░░░░░░░░▀██████░░░░░
// ░░░░░░░░▄░░▀▀█░░░░░░░░░░░░░░░░██████░░░░
// ░░░░░░░█▄██▀▄░░░░░▄███▄▄░░░░░░███████░░░
// ░░░░░░▄▀▀▀██▀░░░░░▄▄▄░░▀█░░░░█████████░░
// ░░░░░▄▀░░░░▄▀░▄░░█▄██▀▄░░░░░██████████░░
// ░░░░░█░░░░▀░░░█░░░▀▀▀▀▀░░░░░██████████▄░
// ░░░░░░░▄█▄░░░░░▄░░░░░░░░░░░░██████████▀░
// ░░░░░░█▀░░░░▀▀░░░░░░░░░░░░░███▀███████░░
// ░░░▄▄░▀░▄░░░░░░░░░░░░░░░░░░▀░░░██████░░░
// ██████░░█▄█▀░▄░░██░░░░░░░░░░░█▄█████▀░░░
// ██████░░░▀████▀░▀░░░░░░░░░░░▄▀█████████▄
// ██████░░░░░░░░░░░░░░░░░░░░▀▄████████████
// ██████░░▄░░░░░░░░░░░░░▄░░░██████████████
// ██████░░░░░░░░░░░░░▄█▀░░▄███████████████
// ███████▄▄░░░░░░░░░▀░░░▄▀▄███████████████

// ListBackupsPaginator is a paginator for DescribeCapacityProviders
type ListBackupsPaginator struct {
	client    Client
	params    *dynamodb.ListBackupsInput
	lastARN   *string
	firstPage bool
}

// NewListBackupsPaginator returns a new ListBackupsPaginator
func NewListBackupsPaginator(client Client, params *dynamodb.ListBackupsInput) *ListBackupsPaginator {
	if params == nil {
		params = &dynamodb.ListBackupsInput{}
	}

	return &ListBackupsPaginator{
		client:    client,
		params:    params,
		firstPage: true,
		lastARN:   params.ExclusiveStartBackupArn,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListBackupsPaginator) HasMorePages() bool {
	return p.firstPage || (p.lastARN != nil && len(*p.lastARN) != 0)
}

// NextPage retrieves the next DescribeCapacityProviders page.
func (p *ListBackupsPaginator) NextPage(ctx context.Context, optFns ...func(*dynamodb.Options)) (*dynamodb.ListBackupsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.ExclusiveStartBackupArn = p.lastARN

	result, err := p.client.ListBackups(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.lastARN
	p.lastARN = result.LastEvaluatedBackupArn

	if prevToken != nil &&
		p.lastARN != nil &&
		*prevToken == *p.lastARN {
		p.lastARN = nil
	}

	return result, nil
}
