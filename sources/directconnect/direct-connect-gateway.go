package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func directConnectGatewayOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeDirectConnectGatewaysInput, output *directconnect.DescribeDirectConnectGatewaysOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, directConnectGateway := range output.DirectConnectGateways {
		attributes, err := sources.ToAttributesCase(directConnectGateway, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-virtual-gateway",
			UniqueAttribute: "directConnectGatewayId",
			Attributes:      attributes,
			Scope:           scope,
		}

		// stateChangeError =>The error message if the state of an object failed to advance.
		if directConnectGateway.StateChangeError != nil {
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		} else {
			item.Health = sdp.Health_HEALTH_OK.Enum()
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-direct-connect-gateway
// +overmind:descriptiveType Direct Connect Gateway
// +overmind:get Get a direct connect gateway by ID
// +overmind:list List all direct connect gateways
// +overmind:search Search direct connect gateway by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_gateway.id

func NewDirectConnectGatewaySource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewaysInput, *directconnect.DescribeDirectConnectGatewaysOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeDirectConnectGatewaysInput, *directconnect.DescribeDirectConnectGatewaysOutput, *directconnect.Client, *directconnect.Options]{
		Config:    config,
		Client:    directconnect.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "directconnect-virtual-gateway",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeDirectConnectGatewaysInput) (*directconnect.DescribeDirectConnectGatewaysOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeDirectConnectGateways(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeDirectConnectGatewaysInput, error) {
			return &directconnect.DescribeDirectConnectGatewaysInput{
				DirectConnectGatewayId: &query,
			}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeDirectConnectGatewaysInput, error) {
			return &directconnect.DescribeDirectConnectGatewaysInput{}, nil
		},
		OutputMapper: directConnectGatewayOutputMapper,
	}
}
