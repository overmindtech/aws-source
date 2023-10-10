package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func capacityReservationFleetOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeCapacityReservationFleetsInput, output *ec2.DescribeCapacityReservationFleetsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, cr := range output.CapacityReservationFleets {
		attributes, err := sources.ToAttributesCase(cr)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "ec2-capacity-reservation-fleet",
			UniqueAttribute: "capacityReservationFleetId",
			Attributes:      attributes,
			Scope:           scope,
		}

		for _, spec := range cr.InstanceTypeSpecifications {
			if spec.AvailabilityZone != nil {
				// +overmind:link ec2-availability-zone
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-availability-zone",
						Method: sdp.QueryMethod_GET,
						Query:  *spec.AvailabilityZone,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the AZ will affect this
						In: true,
						// We can't affect the AZ
						Out: false,
					},
				})
			}

			if spec.CapacityReservationId != nil {
				// +overmind:link ec2-capacity-reservation
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-capacity-reservation",
						Method: sdp.QueryMethod_GET,
						Query:  *spec.CapacityReservationId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the fleet will affect the reservation
						Out: true,
						// The reservation won't affect us
						In: false,
					},
				})
			}
		}

		switch cr.State {
		case types.CapacityReservationFleetStateSubmitted:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.CapacityReservationFleetStateModifying:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.CapacityReservationFleetStateActive:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.CapacityReservationFleetStatePartiallyFulfilled:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.CapacityReservationFleetStateExpiring:
			item.Health = sdp.Health_HEALTH_WARNING.Enum()
		case types.CapacityReservationFleetStateExpired:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		case types.CapacityReservationFleetStateCancelling:
			item.Health = sdp.Health_HEALTH_WARNING.Enum()
		case types.CapacityReservationFleetStateCancelled:
			item.Health = sdp.Health_HEALTH_UNKNOWN.Enum()
		case types.CapacityReservationFleetStateFailed:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-capacity-reservation-fleet
// +overmind:descriptiveType Capacity Reservation Fleet
// +overmind:get Get a capacity reservation fleet by ID
// +overmind:list List capacity reservation fleets
// +overmind:search Search capacity reservation fleets by ARN
// +overmind:group AWS

func NewCapacityReservationFleetSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeCapacityReservationFleetsInput, *ec2.DescribeCapacityReservationFleetsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeCapacityReservationFleetsInput, *ec2.DescribeCapacityReservationFleetsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-capacity-reservation-fleet",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeCapacityReservationFleetsInput) (*ec2.DescribeCapacityReservationFleetsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeCapacityReservationFleets(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*ec2.DescribeCapacityReservationFleetsInput, error) {
			return &ec2.DescribeCapacityReservationFleetsInput{
				CapacityReservationFleetIds: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*ec2.DescribeCapacityReservationFleetsInput, error) {
			return &ec2.DescribeCapacityReservationFleetsInput{}, nil
		},
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeCapacityReservationFleetsInput) sources.Paginator[*ec2.DescribeCapacityReservationFleetsOutput, *ec2.Options] {
			return ec2.NewDescribeCapacityReservationFleetsPaginator(client, params)
		},
		OutputMapper: capacityReservationFleetOutputMapper,
	}
}
