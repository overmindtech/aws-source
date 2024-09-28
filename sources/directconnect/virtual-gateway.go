package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func virtualGatewayOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeVirtualGatewaysInput, output *directconnect.DescribeVirtualGatewaysOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, virtualGateway := range output.VirtualGateways {
		attributes, err := sources.ToAttributesWithExclude(virtualGateway, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-virtual-gateway",
			UniqueAttribute: "VirtualGatewayId",
			Attributes:      attributes,
			Scope:           scope,
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-virtual-gateway
// +overmind:descriptiveType Direct Connect Virtual Gateway
// +overmind:get Get a virtual gateway by ID
// +overmind:list List all virtual gateways
// +overmind:search Search virtual gateways by ARN
// +overmind:group AWS

func NewVirtualGatewaySource(client *directconnect.Client, accountID string, region string) *sources.DescribeOnlySource[*directconnect.DescribeVirtualGatewaysInput, *directconnect.DescribeVirtualGatewaysOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeVirtualGatewaysInput, *directconnect.DescribeVirtualGatewaysOutput, *directconnect.Client, *directconnect.Options]{
		Region:    region,
		Client:    client,
		AccountID: accountID,
		ItemType:  "directconnect-virtual-gateway",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeVirtualGatewaysInput) (*directconnect.DescribeVirtualGatewaysOutput, error) {
			return client.DescribeVirtualGateways(ctx, input)
		},
		// We want to use the list API for get and list operations
		UseListForGet: true,
		InputMapperGet: func(scope, _ string) (*directconnect.DescribeVirtualGatewaysInput, error) {
			return &directconnect.DescribeVirtualGatewaysInput{}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeVirtualGatewaysInput, error) {
			return &directconnect.DescribeVirtualGatewaysInput{}, nil
		},
		OutputMapper: virtualGatewayOutputMapper,
	}
}
