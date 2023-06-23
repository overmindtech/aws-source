package efs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func BackupPolicyOutputMapper(scope string, input *efs.DescribeBackupPolicyInput, output *efs.DescribeBackupPolicyOutput) ([]*sdp.Item, error) {
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

	attrs, err := sources.ToAttributesCase(output)

	if err != nil {
		return nil, err
	}

	// Add the filesystem ID as an attribute
	err = attrs.Set("fileSystemId", *input.FileSystemId)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "efs-backup-policy",
		UniqueAttribute: "fileSystemId",
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

func NewBackupPolicySource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*efs.DescribeBackupPolicyInput, *efs.DescribeBackupPolicyOutput, *efs.Client, *efs.Options] {
	return &sources.DescribeOnlySource[*efs.DescribeBackupPolicyInput, *efs.DescribeBackupPolicyOutput, *efs.Client, *efs.Options]{
		ItemType:  "efs-backup-policy",
		Config:    config,
		Client:    efs.NewFromConfig(config),
		AccountID: accountID,
		DescribeFunc: func(ctx context.Context, client *efs.Client, input *efs.DescribeBackupPolicyInput) (*efs.DescribeBackupPolicyOutput, error) {
			// Wait for rate limiting
			<-limit.C
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
