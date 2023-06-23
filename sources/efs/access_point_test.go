package efs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestAccessPointOutputMapper(t *testing.T) {
	output := &efs.DescribeAccessPointsOutput{
		AccessPoints: []types.AccessPointDescription{
			{
				AccessPointArn: sources.PtrString("arn:aws:elasticfilesystem:eu-west-2:944651592624:access-point/fsap-073b1534eafbc5ee2"),
				AccessPointId:  sources.PtrString("fsap-073b1534eafbc5ee2"),
				ClientToken:    sources.PtrString("pvc-66e4418c-edf5-4a0e-9834-5945598d51fe"),
				FileSystemId:   sources.PtrString("fs-0c6f2f41e957f42a9"),
				LifeCycleState: types.LifeCycleStateAvailable,
				Name:           sources.PtrString("example access point"),
				OwnerId:        sources.PtrString("944651592624"),
				PosixUser: &types.PosixUser{
					Gid: sources.PtrInt64(1000),
					Uid: sources.PtrInt64(1000),
					SecondaryGids: []int64{
						1002,
					},
				},
				RootDirectory: &types.RootDirectory{
					CreationInfo: &types.CreationInfo{
						OwnerGid:    sources.PtrInt64(1000),
						OwnerUid:    sources.PtrInt64(1000),
						Permissions: sources.PtrString("700"),
					},
					Path: sources.PtrString("/etc/foo"),
				},
				Tags: []types.Tag{
					{
						Key:   sources.PtrString("Name"),
						Value: sources.PtrString("example access point"),
					},
				},
			},
		},
	}

	items, err := AccessPointOutputMapper("foo", nil, output)

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
			ExpectedQuery:  "fs-0c6f2f41e957f42a9",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewAccessPointSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewAccessPointSource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
