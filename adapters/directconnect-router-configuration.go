package adapters

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"

	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

func routerConfigurationOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeRouterConfigurationInput, output *directconnect.DescribeRouterConfigurationOutput) ([]*sdp.Item, error) {
	if output == nil || output.Router == nil {
		return nil, nil
	}

	attributes, err := adapterhelpers.ToAttributesWithExclude(output, "tags")
	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "directconnect-router-configuration",
		UniqueAttribute: "VirtualInterfaceId",
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

func NewRouterConfigurationAdapter(client *directconnect.Client, accountID string, region string) *adapterhelpers.DescribeOnlyAdapter[*directconnect.DescribeRouterConfigurationInput, *directconnect.DescribeRouterConfigurationOutput, *directconnect.Client, *directconnect.Options] {
	return &adapterhelpers.DescribeOnlyAdapter[*directconnect.DescribeRouterConfigurationInput, *directconnect.DescribeRouterConfigurationOutput, *directconnect.Client, *directconnect.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "directconnect-router-configuration",
		AdapterMetadata: routerConfigurationAdapterMetadata,
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeRouterConfigurationInput) (*directconnect.DescribeRouterConfigurationOutput, error) {
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

var routerConfigurationAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:            "directconnect-router-configuration",
	DescriptiveName: "Router Configuration",
	Category:        sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
		Get:               true,
		Search:            true,
		GetDescription:    "Get a Router Configuration by Virtual Interface ID",
		SearchDescription: "Search Router Configuration by ARN",
	},
	PotentialLinks: []string{"directconnect-virtual-interface"},
	TerraformMappings: []*sdp.TerraformMapping{
		{TerraformQueryMap: "aws_dx_router_configuration.virtual_interface_id"},
	},
})
