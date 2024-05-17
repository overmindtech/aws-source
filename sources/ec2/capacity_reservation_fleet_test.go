package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestCapacityReservationFleetOutputMapper(t *testing.T) {
	output := &ec2.DescribeCapacityReservationFleetsOutput{
		CapacityReservationFleets: []types.CapacityReservationFleet{
			{
				AllocationStrategy:          sources.PtrString("prioritized"),
				CapacityReservationFleetArn: sources.PtrString("arn:aws:ec2:us-east-1:123456789012:capacity-reservation/fleet/crf-1234567890abcdef0"),
				CapacityReservationFleetId:  sources.PtrString("crf-1234567890abcdef0"),
				CreateTime:                  sources.PtrTime(time.Now()),
				EndDate:                     nil,
				InstanceMatchCriteria:       types.FleetInstanceMatchCriteriaOpen,
				InstanceTypeSpecifications: []types.FleetCapacityReservation{
					{
						AvailabilityZone:      sources.PtrString("us-east-1a"), // link
						AvailabilityZoneId:    sources.PtrString("use1-az1"),
						CapacityReservationId: sources.PtrString("cr-1234567890abcdef0"), // link
						CreateDate:            sources.PtrTime(time.Now()),
						EbsOptimized:          sources.PtrBool(true),
						FulfilledCapacity:     sources.PtrFloat64(1),
						InstancePlatform:      types.CapacityReservationInstancePlatformLinuxUnix,
						InstanceType:          types.InstanceTypeA12xlarge,
						Priority:              sources.PtrInt32(1),
						TotalInstanceCount:    sources.PtrInt32(1),
						Weight:                sources.PtrFloat64(1),
					},
				},
				State:                  types.CapacityReservationFleetStateActive, // health
				Tenancy:                types.FleetCapacityReservationTenancyDefault,
				TotalFulfilledCapacity: sources.PtrFloat64(1),
				TotalTargetCapacity:    sources.PtrInt32(1),
			},
		},
	}

	items, err := capacityReservationFleetOutputMapper(context.Background(), nil, "foo", nil, output)

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
	tests := sources.QueryTests{}

	tests.Execute(t, item)

}

func TestNewCapacityReservationFleetSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewCapacityReservationFleetSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
