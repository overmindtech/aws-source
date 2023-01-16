package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestVolumeInputMapperGet(t *testing.T) {
	input, err := VolumeInputMapperGet("foo", "bar")

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
	input, err := VolumeInputMapperList("foo")

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
						AttachTime:          sources.PtrTime(time.Now()),
						Device:              sources.PtrString("/dev/sdb"),
						InstanceId:          sources.PtrString("i-0667d3ca802741e30"),
						State:               types.VolumeAttachmentStateAttaching,
						VolumeId:            sources.PtrString("vol-0eae6976b359d8825"),
						DeleteOnTermination: sources.PtrBool(false),
					},
				},
				AvailabilityZone:   sources.PtrString("eu-west-2c"),
				CreateTime:         sources.PtrTime(time.Now()),
				Encrypted:          sources.PtrBool(false),
				Size:               sources.PtrInt32(8),
				State:              types.VolumeStateInUse,
				VolumeId:           sources.PtrString("vol-0eae6976b359d8825"),
				Iops:               sources.PtrInt32(3000),
				VolumeType:         types.VolumeTypeGp3,
				MultiAttachEnabled: sources.PtrBool(false),
				Throughput:         sources.PtrInt32(125),
			},
		},
	}

	items, err := VolumeOutputMapper("foo", output)

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
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "i-0667d3ca802741e30",
			ExpectedScope:  item.Scope,
		},
	}

	tests.Execute(t, item)

}

func TestNewVolumeSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewVolumeSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
