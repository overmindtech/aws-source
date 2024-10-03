package efs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/efs"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func FileSystemOutputMapper(_ context.Context, _ *efs.Client, scope string, input *efs.DescribeFileSystemsInput, output *efs.DescribeFileSystemsOutput) ([]*sdp.Item, error) {
	if output == nil {
		return nil, errors.New("nil output from AWS")
	}

	items := make([]*sdp.Item, 0)

	for _, fs := range output.FileSystems {
		attrs, err := sources.ToAttributesWithExclude(fs, "tags")

		if err != nil {
			return nil, err
		}

		if fs.FileSystemId == nil {
			return nil, errors.New("filesystem has nil id")
		}

		item := sdp.Item{
			Type:            "efs-file-system",
			UniqueAttribute: "FileSystemId",
			Scope:           scope,
			Attributes:      attrs,
			Health:          lifeCycleStateToHealth(fs.LifeCycleState),
			Tags:            tagsToMap(fs.Tags),
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						Type:   "efs-backup-policy",
						Method: sdp.QueryMethod_GET,
						Query:  *fs.FileSystemId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the backup policy could effect the file
						// system in that it might no longer be backed up
						In: true,
						// Changing the file system will not effect the backup
						Out: false,
					},
				},
				{
					Query: &sdp.Query{
						Type:   "efs-mount-target",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *fs.FileSystemId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// These are tightly coupled
						In:  true,
						Out: true,
					},
				},
			},
		}

		if fs.KmsKeyId != nil {
			// KMS key ID is an ARN
			if arn, err := sources.ParseARN(*fs.KmsKeyId); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "kms-key",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *fs.KmsKeyId,
						Scope:  sources.FormatScope(arn.AccountID, arn.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the key will affect us
						In: true,
						// We can't affect the key
						Out: false,
					},
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type efs-file-system
// +overmind:descriptiveType EFS File System
// +overmind:get Get an file system by ID
// +overmind:list List all file systems
// +overmind:search Search for an file system by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_efs_file_system.id

func NewFileSystemSource(client *efs.Client, accountID string, region string) *sources.DescribeOnlySource[*efs.DescribeFileSystemsInput, *efs.DescribeFileSystemsOutput, *efs.Client, *efs.Options] {
	return &sources.DescribeOnlySource[*efs.DescribeFileSystemsInput, *efs.DescribeFileSystemsOutput, *efs.Client, *efs.Options]{
		ItemType:        "efs-file-system",
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		AdapterMetadata: FileSystemMetadata(),
		DescribeFunc: func(ctx context.Context, client *efs.Client, input *efs.DescribeFileSystemsInput) (*efs.DescribeFileSystemsOutput, error) {
			return client.DescribeFileSystems(ctx, input)
		},
		PaginatorBuilder: func(client *efs.Client, params *efs.DescribeFileSystemsInput) sources.Paginator[*efs.DescribeFileSystemsOutput, *efs.Options] {
			return efs.NewDescribeFileSystemsPaginator(client, params)
		},
		InputMapperGet: func(scope, query string) (*efs.DescribeFileSystemsInput, error) {
			return &efs.DescribeFileSystemsInput{
				FileSystemId: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*efs.DescribeFileSystemsInput, error) {
			return &efs.DescribeFileSystemsInput{}, nil
		},
		OutputMapper: FileSystemOutputMapper,
	}
}

func FileSystemMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "efs-file-system",
		DescriptiveName: "EFS File System",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:            true,
			List:           true,
			Search:         true,
			GetDescription: "Get an file system by ID",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_efs_file_system.id"},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_STORAGE,
	}
}
