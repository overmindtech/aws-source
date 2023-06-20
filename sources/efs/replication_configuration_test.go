package efs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestReplicationConfigurationOutputMapper(t *testing.T) {
	output := &efs.DescribeReplicationConfigurationsOutput{
		Replications: []types.ReplicationConfigurationDescription{
			{
				CreationTime: sources.PtrTime(time.Now()),
				Destinations: []types.Destination{
					{
						FileSystemId:            sources.PtrString("fs-12345678"),
						Region:                  sources.PtrString("eu-west-1"),
						Status:                  types.ReplicationStatusEnabled,
						LastReplicatedTimestamp: sources.PtrTime(time.Now()),
					},
					{
						FileSystemId:            sources.PtrString("fs-98765432"),
						Region:                  sources.PtrString("us-west-2"),
						Status:                  types.ReplicationStatusError,
						LastReplicatedTimestamp: sources.PtrTime(time.Now()),
					},
				},
				OriginalSourceFileSystemArn: sources.PtrString(""), // TODO
				SourceFileSystemArn:         sources.PtrString(""), // TODO
				SourceFileSystemId:          sources.PtrString("fs-748927493"),
				SourceFileSystemRegion:      sources.PtrString("us-east-1"),
			},
		},
	}

	accountID := "1234"
	items, err := ReplicationConfigurationOutputMapper(sources.FormatScope(accountID, "eu-west-1"), nil, output)

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
			ExpectedType:   "efs-file-system",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "fs-748927493",
			ExpectedScope:  sources.FormatScope(accountID, "us-east-1"),
		},
		{
			ExpectedType:   "efs-file-system",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "fs-12345678",
			ExpectedScope:  sources.FormatScope(accountID, "eu-west-1"),
		},
		{
			ExpectedType:   "efs-file-system",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "fs-98765432",
			ExpectedScope:  sources.FormatScope(accountID, "us-west-2"),
		},
		{
			ExpectedType:   "efs-file-system",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "", // TODO
			ExpectedScope:  "", // TODO
		},
	}

	tests.Execute(t, item)

	if *item.Health != sdp.Health_HEALTH_ERROR {
		t.Errorf("expected health to be ERROR, got %v", item.Health.String())
	}
}

func TestNewReplicationConfigurationSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewReplicationConfigurationSource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
