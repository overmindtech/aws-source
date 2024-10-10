package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestImageInputMapperGet(t *testing.T) {
	input, err := imageInputMapperGet("foo", "az-name")

	if err != nil {
		t.Error(err)
	}

	if len(input.ImageIds) != 1 {
		t.Fatalf("expected 1 zone names, got %v", len(input.ImageIds))
	}

	if input.ImageIds[0] != "az-name" {
		t.Errorf("expected zone name to be to be az-name, got %v", input.ImageIds[0])
	}
}

func TestImageInputMapperList(t *testing.T) {

	input, err := imageInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.ImageIds) != 0 {
		t.Fatalf("expected 0 zone names, got %v", len(input.ImageIds))
	}
}

func TestImageOutputMapper(t *testing.T) {
	output := ec2.DescribeImagesOutput{
		Images: []types.Image{
			{
				Architecture:    "x86_64",
				CreationDate:    adapters.PtrString("2022-12-16T19:37:36.000Z"),
				ImageId:         adapters.PtrString("ami-0ed3646be6ecd97c5"),
				ImageLocation:   adapters.PtrString("052392120703/test"),
				ImageType:       types.ImageTypeValuesMachine,
				Public:          adapters.PtrBool(false),
				OwnerId:         adapters.PtrString("052392120703"),
				PlatformDetails: adapters.PtrString("Linux/UNIX"),
				UsageOperation:  adapters.PtrString("RunInstances"),
				State:           types.ImageStateAvailable,
				BlockDeviceMappings: []types.BlockDeviceMapping{
					{
						DeviceName: adapters.PtrString("/dev/xvda"),
						Ebs: &types.EbsBlockDevice{
							DeleteOnTermination: adapters.PtrBool(true),
							SnapshotId:          adapters.PtrString("snap-0efd796ecbd599f8d"),
							VolumeSize:          adapters.PtrInt32(8),
							VolumeType:          types.VolumeTypeGp2,
							Encrypted:           adapters.PtrBool(false),
						},
					},
				},
				EnaSupport:         adapters.PtrBool(true),
				Hypervisor:         types.HypervisorTypeXen,
				Name:               adapters.PtrString("test"),
				RootDeviceName:     adapters.PtrString("/dev/xvda"),
				RootDeviceType:     types.DeviceTypeEbs,
				SriovNetSupport:    adapters.PtrString("simple"),
				VirtualizationType: types.VirtualizationTypeHvm,
				Tags: []types.Tag{
					{
						Key:   adapters.PtrString("Name"),
						Value: adapters.PtrString("test"),
					},
				},
			},
		},
	}

	items, err := imageOutputMapper(context.Background(), nil, "foo", nil, &output)

	if err != nil {
		t.Error(err)
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

	if item.UniqueAttributeValue() != *output.Images[0].ImageId {
		t.Errorf("Expected item unique attribute value to be %v, got %v", *output.Images[0].ImageId, item.UniqueAttributeValue())
	}
}

func TestNewImageAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewImageAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
