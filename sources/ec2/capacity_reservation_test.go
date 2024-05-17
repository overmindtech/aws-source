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

func TestCapacityReservationOutputMapper(t *testing.T) {
	output := &ec2.DescribeCapacityReservationsOutput{
		CapacityReservations: []types.CapacityReservation{
			{
				AvailabilityZone:           sources.PtrString("us-east-1a"), // links
				AvailabilityZoneId:         sources.PtrString("use1-az1"),
				AvailableInstanceCount:     sources.PtrInt32(1),
				CapacityReservationArn:     sources.PtrString("arn:aws:ec2:us-east-1:123456789012:capacity-reservation/cr-1234567890abcdef0"),
				CapacityReservationId:      sources.PtrString("cr-1234567890abcdef0"),
				CapacityReservationFleetId: sources.PtrString("crf-1234567890abcdef0"), // link
				CreateDate:                 sources.PtrTime(time.Now()),
				EbsOptimized:               sources.PtrBool(true),
				EndDateType:                types.EndDateTypeUnlimited,
				EndDate:                    nil,
				InstanceMatchCriteria:      types.InstanceMatchCriteriaTargeted,
				InstancePlatform:           types.CapacityReservationInstancePlatformLinuxUnix,
				InstanceType:               sources.PtrString("t2.micro"),
				OutpostArn:                 sources.PtrString("arn:aws:ec2:us-east-1:123456789012:outpost/op-1234567890abcdef0"), // link
				OwnerId:                    sources.PtrString("123456789012"),
				PlacementGroupArn:          sources.PtrString("arn:aws:ec2:us-east-1:123456789012:placement-group/pg-1234567890abcdef0"), // link
				StartDate:                  sources.PtrTime(time.Now()),
				State:                      types.CapacityReservationStateActive,
				Tenancy:                    types.CapacityReservationTenancyDefault,
				TotalInstanceCount:         sources.PtrInt32(1),
				CapacityAllocations: []types.CapacityAllocation{
					{
						AllocationType: types.AllocationTypeUsed,
						Count:          sources.PtrInt32(1),
					},
				},
			},
		},
	}

	items, err := capacityReservationOutputMapper(context.Background(), nil, "foo", nil, output)

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
			ExpectedType:   "ec2-capacity-reservation-fleet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "crf-1234567890abcdef0",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "outposts-outpost",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:ec2:us-east-1:123456789012:outpost/op-1234567890abcdef0",
			ExpectedScope:  "123456789012.us-east-1",
		},
		{
			ExpectedType:   "ec2-placement-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:ec2:us-east-1:123456789012:placement-group/pg-1234567890abcdef0",
			ExpectedScope:  "123456789012.us-east-1",
		},
	}

	tests.Execute(t, item)

}

func TestNewCapacityReservationSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewCapacityReservationSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
