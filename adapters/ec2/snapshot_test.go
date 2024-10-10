package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestSnapshotInputMapperGet(t *testing.T) {
	input, err := snapshotInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.SnapshotIds) != 1 {
		t.Fatalf("expected 1 Snapshot ID, got %v", len(input.SnapshotIds))
	}

	if input.SnapshotIds[0] != "bar" {
		t.Errorf("expected Snapshot ID to be bar, got %v", input.SnapshotIds[0])
	}
}

func TestSnapshotInputMapperList(t *testing.T) {
	input, err := snapshotInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.SnapshotIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestSnapshotOutputMapper(t *testing.T) {
	output := &ec2.DescribeSnapshotsOutput{
		Snapshots: []types.Snapshot{
			{
				DataEncryptionKeyId: adapters.PtrString("ek"),
				KmsKeyId:            adapters.PtrString("key"),
				SnapshotId:          adapters.PtrString("id"),
				Description:         adapters.PtrString("foo"),
				Encrypted:           adapters.PtrBool(false),
				OutpostArn:          adapters.PtrString("something"),
				OwnerAlias:          adapters.PtrString("something"),
				OwnerId:             adapters.PtrString("owner"),
				Progress:            adapters.PtrString("50%"),
				RestoreExpiryTime:   adapters.PtrTime(time.Now()),
				StartTime:           adapters.PtrTime(time.Now()),
				State:               types.SnapshotStatePending,
				StateMessage:        adapters.PtrString("pending"),
				StorageTier:         types.StorageTierArchive,
				Tags:                []types.Tag{},
				VolumeId:            adapters.PtrString("volumeId"),
				VolumeSize:          adapters.PtrInt32(1024),
			},
		},
	}

	items, err := snapshotOutputMapper(context.Background(), nil, "foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-volume",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "volumeId",
			ExpectedScope:  item.GetScope(),
		},
	}

	tests.Execute(t, item)

}

func TestNewSnapshotSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewSnapshotSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
