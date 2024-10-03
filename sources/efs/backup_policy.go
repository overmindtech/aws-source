package efs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/efs"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func BackupPolicyOutputMapper(_ context.Context, _ *efs.Client, scope string, input *efs.DescribeBackupPolicyInput, output *efs.DescribeBackupPolicyOutput) ([]*sdp.Item, error) {
	if output == nil {
		return nil, errors.New("nil output from AWS")
	}

	if output.BackupPolicy == nil {
		return nil, errors.New("output contains no backup policy")
	}

	if input == nil {
		return nil, errors.New("nil input")
	}

	if input.FileSystemId == nil {
		return nil, errors.New("nil filesystem ID on input")
	}

	attrs, err := sources.ToAttributesWithExclude(output)

	if err != nil {
		return nil, err
	}

	// Add the filesystem ID as an attribute
	err = attrs.Set("FileSystemId", *input.FileSystemId)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "efs-backup-policy",
		UniqueAttribute: "FileSystemId",
		Scope:           scope,
		Attributes:      attrs,
	}

	return []*sdp.Item{&item}, nil
}

//go:generate docgen ../../docs-data
// +overmind:type efs-backup-policy
// +overmind:descriptiveType EFS Backup Policy
// +overmind:get Get an Backup Policy by file system ID
// +overmind:search Search for an Backup Policy by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_efs_backup_policy.id

func NewBackupPolicySource(client *efs.Client, accountID string, region string) *sources.DescribeOnlySource[*efs.DescribeBackupPolicyInput, *efs.DescribeBackupPolicyOutput, *efs.Client, *efs.Options] {
	return &sources.DescribeOnlySource[*efs.DescribeBackupPolicyInput, *efs.DescribeBackupPolicyOutput, *efs.Client, *efs.Options]{
		ItemType:        "efs-backup-policy",
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		AdapterMetadata: BackupPolicyMetadata(),
		DescribeFunc: func(ctx context.Context, client *efs.Client, input *efs.DescribeBackupPolicyInput) (*efs.DescribeBackupPolicyOutput, error) {
			return client.DescribeBackupPolicy(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*efs.DescribeBackupPolicyInput, error) {
			return &efs.DescribeBackupPolicyInput{
				FileSystemId: &query,
			}, nil
		},
		OutputMapper: BackupPolicyOutputMapper,
	}
}

func BackupPolicyMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "efs-backup-policy",
		DescriptiveName: "EFS Backup Policy",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			Search:            true,
			GetDescription:    "Get an Backup Policy by file system ID",
			SearchDescription: "Search for an Backup Policy by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_efs_backup_policy.id"},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_STORAGE,
	}
}
