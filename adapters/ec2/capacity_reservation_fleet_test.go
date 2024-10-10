package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestCapacityReservationFleetOutputMapper(t *testing.T) {
	output := &ec2.DescribeCapacityReservationFleetsOutput{
		CapacityReservationFleets: []types.CapacityReservationFleet{
			{
				AllocationStrategy:          adapters.PtrString("prioritized"),
				CapacityReservationFleetArn: adapters.PtrString("arn:aws:ec2:us-east-1:123456789012:capacity-reservation/fleet/crf-1234567890abcdef0"),
				CapacityReservationFleetId:  adapters.PtrString("crf-1234567890abcdef0"),
				CreateTime:                  adapters.PtrTime(time.Now()),
				EndDate:                     nil,
				InstanceMatchCriteria:       types.FleetInstanceMatchCriteriaOpen,
				InstanceTypeSpecifications: []types.FleetCapacityReservation{
					{
						AvailabilityZone:      adapters.PtrString("us-east-1a"), // link
						AvailabilityZoneId:    adapters.PtrString("use1-az1"),
						CapacityReservationId: adapters.PtrString("cr-1234567890abcdef0"), // link
						CreateDate:            adapters.PtrTime(time.Now()),
						EbsOptimized:          adapters.PtrBool(true),
						FulfilledCapacity:     adapters.PtrFloat64(1),
						InstancePlatform:      types.CapacityReservationInstancePlatformLinuxUnix,
						InstanceType:          types.InstanceTypeA12xlarge,
						Priority:              adapters.PtrInt32(1),
						TotalInstanceCount:    adapters.PtrInt32(1),
						Weight:                adapters.PtrFloat64(1),
					},
				},
				State:                  types.CapacityReservationFleetStateActive, // health
				Tenancy:                types.FleetCapacityReservationTenancyDefault,
				TotalFulfilledCapacity: adapters.PtrFloat64(1),
				TotalTargetCapacity:    adapters.PtrInt32(1),
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
	tests := adapters.QueryTests{}

	tests.Execute(t, item)

}

func TestNewCapacityReservationFleetSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewCapacityReservationFleetSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
