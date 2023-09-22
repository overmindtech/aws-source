package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// AvailabilityZoneInputMapperGet Maps source calls to the correct input for the AZ API
func availabilityZoneInputMapperGet(scope, query string) (*ec2.DescribeAvailabilityZonesInput, error) {
	return &ec2.DescribeAvailabilityZonesInput{
		ZoneNames: []string{
			query,
		},
	}, nil
}

// AvailabilityZoneInputMapperList Maps source calls to the correct input for the AZ API
func availabilityZoneInputMapperList(scope string) (*ec2.DescribeAvailabilityZonesInput, error) {
	return &ec2.DescribeAvailabilityZonesInput{}, nil
}

// AvailabilityZoneOutputMapper Maps API output to items
func availabilityZoneOutputMapper(scope string, _ *ec2.DescribeAvailabilityZonesInput, output *ec2.DescribeAvailabilityZonesOutput) ([]*sdp.Item, error) {
	if output == nil {
		return nil, errors.New("empty output")
	}

	items := make([]*sdp.Item, len(output.AvailabilityZones))
	var err error
	var attrs *sdp.ItemAttributes

	for i, az := range output.AvailabilityZones {
		attrs, err = sources.ToAttributesCase(az)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "ec2-availability-zone",
			UniqueAttribute: "zoneName",
			Scope:           scope,
			Attributes:      attrs,
		}

		// Link to region
		if az.RegionName != nil {
			// +overmind:link ec2-region
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-region",
					Method: sdp.QueryMethod_GET,
					Query:  *az.RegionName,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Regions don't change
					In:  false,
					Out: false,
				},
			})
		}

		items[i] = &item
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-availability-zone
// +overmind:descriptiveType Availability Zone
// +overmind:get Get an Availability Zone by Name
// +overmind:list List all Availability Zones
// +overmind:group AWS

// NewAvailabilityZoneSource Creates a new source for aws-availabilityzone resources
func NewAvailabilityZoneSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeAvailabilityZonesInput, *ec2.DescribeAvailabilityZonesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeAvailabilityZonesInput, *ec2.DescribeAvailabilityZonesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-availability-zone",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeAvailabilityZones(ctx, input)
		},
		InputMapperGet:  availabilityZoneInputMapperGet,
		InputMapperList: availabilityZoneInputMapperList,
		OutputMapper:    availabilityZoneOutputMapper,
	}
}
