package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func customerMetadataOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeCustomerMetadataInput, output *directconnect.DescribeCustomerMetadataOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, agreement := range output.Agreements {
		attributes, err := sources.ToAttributesCase(agreement, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-customer-metadata",
			UniqueAttribute: "agreementName",
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

func NewCustomerMetadataSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeCustomerMetadataInput, *directconnect.DescribeCustomerMetadataOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeCustomerMetadataInput, *directconnect.DescribeCustomerMetadataOutput, *directconnect.Client, *directconnect.Options]{
		Config:    config,
		Client:    directconnect.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "directconnect-customer-metadata",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeCustomerMetadataInput) (*directconnect.DescribeCustomerMetadataOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
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
