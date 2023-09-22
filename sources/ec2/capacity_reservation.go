package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func capacityReservationOutputMapper(scope string, _ *ec2.DescribeCapacityReservationsInput, output *ec2.DescribeCapacityReservationsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, cr := range output.CapacityReservations {
		attributes, err := sources.ToAttributesCase(cr)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "ec2-capacity-reservation",
			UniqueAttribute: "capacityReservationId",
			Attributes:      attributes,
			Scope:           scope,
		}

		if cr.AvailabilityZone != nil {
			// +overmind:link ec2-availability-zone
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-availability-zone",
					Method: sdp.QueryMethod_GET,
					Query:  *cr.AvailabilityZone,
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

		if cr.CapacityReservationFleetId != nil {
			// +overmind:link ec2-capacity-reservation-fleet
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-capacity-reservation-fleet",
					Method: sdp.QueryMethod_GET,
					Query:  *cr.CapacityReservationFleetId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the fleet will affect this
					In: true,
					// We can't affect the fleet
					Out: false,
				},
			})
		}

		if cr.OutpostArn != nil {
			if arn, err := sources.ParseARN(*cr.OutpostArn); err == nil {
				// +overmind:link outposts-outpost
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "outposts-outpost",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cr.OutpostArn,
						Scope:  sources.FormatScope(arn.AccountID, arn.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the outpost will affect this
						In: true,
						// We can't affect the outpost
						Out: false,
					},
				})
			}
		}

		if cr.PlacementGroupArn != nil {
			if arn, err := sources.ParseARN(*cr.PlacementGroupArn); err == nil {
				// +overmind:link ec2-placement-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-placement-group",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cr.PlacementGroupArn,
						Scope:  sources.FormatScope(arn.AccountID, arn.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the placement group will affect this
						In: true,
						// We can't affect the placement group
						Out: false,
					},
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-capacity-reservation
// +overmind:descriptiveType Capacity Reservation
// +overmind:get Get a capacity reservation by ID
// +overmind:list List all capacity reservations
// +overmind:search Search capacity reservations by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_ec2_capacity_reservation.id

func NewCapacityReservationSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeCapacityReservationsInput, *ec2.DescribeCapacityReservationsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeCapacityReservationsInput, *ec2.DescribeCapacityReservationsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-capacity-reservation",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeCapacityReservationsInput) (*ec2.DescribeCapacityReservationsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeCapacityReservations(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*ec2.DescribeCapacityReservationsInput, error) {
			return &ec2.DescribeCapacityReservationsInput{
				CapacityReservationIds: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*ec2.DescribeCapacityReservationsInput, error) {
			return &ec2.DescribeCapacityReservationsInput{}, nil
		},
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeCapacityReservationsInput) sources.Paginator[*ec2.DescribeCapacityReservationsOutput, *ec2.Options] {
			return ec2.NewDescribeCapacityReservationsPaginator(client, params)
		},
		OutputMapper: capacityReservationOutputMapper,
	}
}
