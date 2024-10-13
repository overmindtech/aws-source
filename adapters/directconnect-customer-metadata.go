package adapters

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"

	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

func customerMetadataOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeCustomerMetadataInput, output *directconnect.DescribeCustomerMetadataOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, agreement := range output.Agreements {
		attributes, err := adapterhelpers.ToAttributesWithExclude(agreement, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-customer-metadata",
			UniqueAttribute: "AgreementName",
			Attributes:      attributes,
			Scope:           scope,
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-customer-metadata
// +overmind:descriptiveType Direct Connect Customer Metadata
// +overmind:get Get a Customer Agreement by Name
// +overmind:list List all Customer Agreements
// +overmind:search Search Customer Agreements by ARN
// +overmind:group AWS

func NewCustomerMetadataAdapter(client *directconnect.Client, accountID string, region string) *adapterhelpers.DescribeOnlyAdapter[*directconnect.DescribeCustomerMetadataInput, *directconnect.DescribeCustomerMetadataOutput, *directconnect.Client, *directconnect.Options] {
	return &adapterhelpers.DescribeOnlyAdapter[*directconnect.DescribeCustomerMetadataInput, *directconnect.DescribeCustomerMetadataOutput, *directconnect.Client, *directconnect.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "directconnect-customer-metadata",
		AdapterMetadata: customerMetadataAdapterMetadata,
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeCustomerMetadataInput) (*directconnect.DescribeCustomerMetadataOutput, error) {
			return client.DescribeCustomerMetadata(ctx, input)
		},
		// We want to use the list API for get and list operations
		UseListForGet: true,
		InputMapperGet: func(scope, _ string) (*directconnect.DescribeCustomerMetadataInput, error) {
			return &directconnect.DescribeCustomerMetadataInput{}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeCustomerMetadataInput, error) {
			return &directconnect.DescribeCustomerMetadataInput{}, nil
		},
		OutputMapper: customerMetadataOutputMapper,
	}
}

var customerMetadataAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:            "directconnect-customer-metadata",
	DescriptiveName: "Customer Metadata",
	SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
		Get:               true,
		List:              true,
		Search:            true,
		GetDescription:    "Get a customer agreement by name",
		ListDescription:   "List all customer agreements",
		SearchDescription: "Search customer agreements by ARN",
	},
	Category: sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
})
