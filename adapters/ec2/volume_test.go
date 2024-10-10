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

func TestVolumeInputMapperGet(t *testing.T) {
	input, err := volumeInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.VolumeIds) != 1 {
		t.Fatalf("expected 1 Volume ID, got %v", len(input.VolumeIds))
	}

	if input.VolumeIds[0] != "bar" {
		t.Errorf("expected Volume ID to be bar, got %v", input.VolumeIds[0])
	}
}

func TestVolumeInputMapperList(t *testing.T) {
	input, err := volumeInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.VolumeIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestVolumeOutputMapper(t *testing.T) {
	output := &ec2.DescribeVolumesOutput{
		Volumes: []types.Volume{
			{
				Attachments: []types.VolumeAttachment{
					{
						AttachTime:          adapters.PtrTime(time.Now()),
						Device:              adapters.PtrString("/dev/sdb"),
						InstanceId:          adapters.PtrString("i-0667d3ca802741e30"),
						State:               types.VolumeAttachmentStateAttaching,
						VolumeId:            adapters.PtrString("vol-0eae6976b359d8825"),
						DeleteOnTermination: adapters.PtrBool(false),
					},
				},
				AvailabilityZone:   adapters.PtrString("eu-west-2c"),
				CreateTime:         adapters.PtrTime(time.Now()),
				Encrypted:          adapters.PtrBool(false),
				Size:               adapters.PtrInt32(8),
				State:              types.VolumeStateInUse,
				VolumeId:           adapters.PtrString("vol-0eae6976b359d8825"),
				Iops:               adapters.PtrInt32(3000),
				VolumeType:         types.VolumeTypeGp3,
				MultiAttachEnabled: adapters.PtrBool(false),
				Throughput:         adapters.PtrInt32(125),
			},
		},
	}

	items, err := volumeOutputMapper(context.Background(), nil, "foo", nil, output)

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
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "i-0667d3ca802741e30",
			ExpectedScope:  item.GetScope(),
		},
	}

	tests.Execute(t, item)

}

func TestNewVolumeSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewVolumeSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
