package efs

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestAccessPointOutputMapper(t *testing.T) {
	output := &efs.DescribeAccessPointsOutput{
		AccessPoints: []types.AccessPointDescription{
			{
				AccessPointArn: adapters.PtrString("arn:aws:elasticfilesystem:eu-west-2:944651592624:access-point/fsap-073b1534eafbc5ee2"),
				AccessPointId:  adapters.PtrString("fsap-073b1534eafbc5ee2"),
				ClientToken:    adapters.PtrString("pvc-66e4418c-edf5-4a0e-9834-5945598d51fe"),
				FileSystemId:   adapters.PtrString("fs-0c6f2f41e957f42a9"),
				LifeCycleState: types.LifeCycleStateAvailable,
				Name:           adapters.PtrString("example access point"),
				OwnerId:        adapters.PtrString("944651592624"),
				PosixUser: &types.PosixUser{
					Gid: adapters.PtrInt64(1000),
					Uid: adapters.PtrInt64(1000),
					SecondaryGids: []int64{
						1002,
					},
				},
				RootDirectory: &types.RootDirectory{
					CreationInfo: &types.CreationInfo{
						OwnerGid:    adapters.PtrInt64(1000),
						OwnerUid:    adapters.PtrInt64(1000),
						Permissions: adapters.PtrString("700"),
					},
					Path: adapters.PtrString("/etc/foo"),
				},
				Tags: []types.Tag{
					{
						Key:   adapters.PtrString("Name"),
						Value: adapters.PtrString("example access point"),
					},
				},
			},
		},
	}

	items, err := AccessPointOutputMapper(context.Background(), nil, "foo", nil, output)

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
	tests := adapters.QueryTests{
		{
			ExpectedType:   "efs-file-system",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "fs-0c6f2f41e957f42a9",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewAccessPointAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewAccessPointAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
