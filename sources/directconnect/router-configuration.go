package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func routerConfigurationOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeRouterConfigurationInput, output *directconnect.DescribeRouterConfigurationOutput) ([]*sdp.Item, error) {
	if output == nil || output.Router == nil {
		return nil, nil
	}

	attributes, err := sources.ToAttributesCase(output, "tags")
	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "directconnect-router-configuration",
		UniqueAttribute: "virtualInterfaceId",
		Attributes:      attributes,
		Scope:           scope,
	}

	if output.VirtualInterfaceId != nil {
		// +overmind:link directconnect-virtual-interface
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "directconnect-virtual-interface",
				Method: sdp.QueryMethod_GET,
				Query:  *output.VirtualInterfaceId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// They are tightly coupled
				In:  true,
				Out: true,
			},
		})
	}

	return []*sdp.Item{
		&item,
	}, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-router-configuration
// +overmind:descriptiveType Direct Connect Router Configuration
// +overmind:get Get a Router Configuration by Virtual Interface ID
// +overmind:search Search Router Configuration by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dx_router_configuration.virtual_interface_id

func NewRouterConfigurationSource(client *directconnect.Client, accountID string, region string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeRouterConfigurationInput, *directconnect.DescribeRouterConfigurationOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeRouterConfigurationInput, *directconnect.DescribeRouterConfigurationOutput, *directconnect.Client, *directconnect.Options]{

		Client:    client,
		AccountID: accountID,
		ItemType:  "directconnect-router-configuration",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeRouterConfigurationInput) (*directconnect.DescribeRouterConfigurationOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeRouterConfiguration(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeRouterConfigurationInput, error) {
			return &directconnect.DescribeRouterConfigurationInput{
				VirtualInterfaceId: &query,
			}, nil
		},
		OutputMapper: routerConfigurationOutputMapper,
	}
}
