package efs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestFileSystemOutputMapper(t *testing.T) {
	output := &efs.DescribeFileSystemsOutput{
		FileSystems: []types.FileSystemDescription{
			{
				CreationTime:         sources.PtrTime(time.Now()),
				CreationToken:        sources.PtrString("TOKEN"),
				FileSystemId:         sources.PtrString("fs-1231123123"),
				LifeCycleState:       types.LifeCycleStateAvailable,
				NumberOfMountTargets: 10,
				OwnerId:              sources.PtrString("944651592624"),
				PerformanceMode:      types.PerformanceModeGeneralPurpose,
				SizeInBytes: &types.FileSystemSize{
					Value:           1024,
					Timestamp:       sources.PtrTime(time.Now()),
					ValueInIA:       sources.PtrInt64(2048),
					ValueInStandard: sources.PtrInt64(128),
				},
				Tags: []types.Tag{
					{
						Key:   sources.PtrString("foo"),
						Value: sources.PtrString("bar"),
					},
				},
				AvailabilityZoneId:           sources.PtrString("use1-az1"),
				AvailabilityZoneName:         sources.PtrString("us-east-1"),
				Encrypted:                    sources.PtrBool(true),
				FileSystemArn:                sources.PtrString("arn:aws:elasticfilesystem:eu-west-2:944651592624:file-system/fs-0c6f2f41e957f42a9"),
				KmsKeyId:                     sources.PtrString("arn:aws:kms:eu-west-2:944651592624:key/be76a6fa-d307-41c2-a4e3-cbfba2440747"),
				Name:                         sources.PtrString("test"),
				ProvisionedThroughputInMibps: sources.PtrFloat64(64),
				ThroughputMode:               types.ThroughputModeBursting,
			},
		},
	}

	items, err := FileSystemOutputMapper(context.Background(), nil, "foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.QueryTests{
		{
			ExpectedType:   "efs-backup-policy",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "fs-1231123123",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:kms:eu-west-2:944651592624:key/be76a6fa-d307-41c2-a4e3-cbfba2440747",
			ExpectedScope:  "944651592624.eu-west-2",
		},
		{
			ExpectedType:   "efs-mount-target",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "fs-1231123123",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewFileSystemSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewFileSystemSource(client, account, region)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
