package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func locationOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeLocationsInput, output *directconnect.DescribeLocationsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, location := range output.Locations {
		attributes, err := sources.ToAttributesCase(location, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-location",
			UniqueAttribute: "locationCode",
			Attributes:      attributes,
			Scope:           scope,
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-location
// +overmind:descriptiveType Direct Connect Location
// +overmind:get Get a Location by its code
// +overmind:list List all Direct Connect Locations
// +overmind:search Search Direct Connect Locations by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_location.location_code

func NewLocationSource(client *directconnect.Client, accountID string, region string) *sources.DescribeOnlySource[*directconnect.DescribeLocationsInput, *directconnect.DescribeLocationsOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeLocationsInput, *directconnect.DescribeLocationsOutput, *directconnect.Client, *directconnect.Options]{
		Region:    region,
		Client:    client,
		AccountID: accountID,
		ItemType:  "directconnect-location",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeLocationsInput) (*directconnect.DescribeLocationsOutput, error) {
			return client.DescribeLocations(ctx, input)
		},
		// We want to use the list API for get and list operations
		UseListForGet: true,
		InputMapperGet: func(scope, _ string) (*directconnect.DescribeLocationsInput, error) {
			return &directconnect.DescribeLocationsInput{}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeLocationsInput, error) {
			return &directconnect.DescribeLocationsInput{}, nil
		},
		OutputMapper: locationOutputMapper,
	}
}
