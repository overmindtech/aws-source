package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// AvailabilityZoneInputMapper Maps source calls to the correct input for the AZ API
func AvailabilityZoneInputMapper(scope, query string, method sdp.RequestMethod) (*ec2.DescribeAvailabilityZonesInput, error) {
	var input ec2.DescribeAvailabilityZonesInput

	switch method {
	case sdp.RequestMethod_GET:
		input = ec2.DescribeAvailabilityZonesInput{
			ZoneNames: []string{
				query,
			},
		}
	case sdp.RequestMethod_LIST:
		input = ec2.DescribeAvailabilityZonesInput{}
	default:
		return nil, errors.New("unsupported method")
	}

	return &input, nil
}

// AvailabilityZoneOutputMapper Maps API output to items
func AvailabilityZoneOutputMapper(scope string, output *ec2.DescribeAvailabilityZonesOutput) ([]*sdp.Item, error) {
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
			Type:            "ec2-availabilityzone",
			UniqueAttribute: "zoneName",
			Scope:           scope,
			Attributes:      attrs,
		}

		// Link to region
		if az.RegionName != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-region",
				Method: sdp.RequestMethod_GET,
				Query:  *az.RegionName,
				Scope:  scope,
			})
		}

		items[i] = &item
	}

	return items, nil
}

// NewAvailabilityZoneSource Creates a new source for aws-availabilityzone resources
func NewAvailabilityZoneSource(config aws.Config, accountID string) *EC2Source[*ec2.DescribeAvailabilityZonesInput, *ec2.DescribeAvailabilityZonesOutput] {
	return &EC2Source[*ec2.DescribeAvailabilityZonesInput, *ec2.DescribeAvailabilityZonesOutput]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-availabilityzone",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error) {
			return client.DescribeAvailabilityZones(ctx, input)
		},
		InputMapper:  AvailabilityZoneInputMapper,
		OutputMapper: AvailabilityZoneOutputMapper,
	}
}
