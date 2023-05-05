package ec2

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
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
				CreationDate:    sources.PtrString("2022-12-16T19:37:36.000Z"),
				ImageId:         sources.PtrString("ami-0ed3646be6ecd97c5"),
				ImageLocation:   sources.PtrString("052392120703/test"),
				ImageType:       types.ImageTypeValuesMachine,
				Public:          sources.PtrBool(false),
				OwnerId:         sources.PtrString("052392120703"),
				PlatformDetails: sources.PtrString("Linux/UNIX"),
				UsageOperation:  sources.PtrString("RunInstances"),
				State:           types.ImageStateAvailable,
				BlockDeviceMappings: []types.BlockDeviceMapping{
					{
						DeviceName: sources.PtrString("/dev/xvda"),
						Ebs: &types.EbsBlockDevice{
							DeleteOnTermination: sources.PtrBool(true),
							SnapshotId:          sources.PtrString("snap-0efd796ecbd599f8d"),
							VolumeSize:          sources.PtrInt32(8),
							VolumeType:          types.VolumeTypeGp2,
							Encrypted:           sources.PtrBool(false),
						},
					},
				},
				EnaSupport:         sources.PtrBool(true),
				Hypervisor:         types.HypervisorTypeXen,
				Name:               sources.PtrString("test"),
				RootDeviceName:     sources.PtrString("/dev/xvda"),
				RootDeviceType:     types.DeviceTypeEbs,
				SriovNetSupport:    sources.PtrString("simple"),
				VirtualizationType: types.VirtualizationTypeHvm,
			},
		},
	}

	items, err := imageOutputMapper("foo", nil, &output)

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

func TestNewImageSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewImageSource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
