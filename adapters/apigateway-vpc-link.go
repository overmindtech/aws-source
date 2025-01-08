package adapters

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
)

// convertGetVpcLinkOutputToVpcLink converts a GetVpcLinkOutput to a VpcLink
func convertGetVpcLinkOutputToVpcLink(output *apigateway.GetVpcLinkOutput) *types.VpcLink {
	return &types.VpcLink{
		Id:          output.Id,
		Name:        output.Name,
		Description: output.Description,
		TargetArns:  output.TargetArns,
		Status:      output.Status,
		Tags:        output.Tags,
	}
}

func vpcLinkListFunc(ctx context.Context, client *apigateway.Client, _ string) ([]*types.VpcLink, error) {
	out, err := client.GetVpcLinks(ctx, &apigateway.GetVpcLinksInput{})
	if err != nil {
		return nil, err
	}

	var items []*types.VpcLink
	for _, vpcLink := range out.Items {
		items = append(items, &vpcLink)
	}

	return items, nil
}

func vpcLinkOutputMapper(scope string, awsItem *types.VpcLink) (*sdp.Item, error) {
	attributes, err := adapterhelpers.ToAttributesWithExclude(awsItem, "tags")
	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "apigateway-vpc-link",
		UniqueAttribute: "Id",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            awsItem.Tags,
	}

	// The status of the VPC link. The valid values are AVAILABLE , PENDING , DELETING , or FAILED.
	switch awsItem.Status {
	case types.VpcLinkStatusAvailable:
		item.Health = sdp.Health_HEALTH_OK.Enum()
	case types.VpcLinkStatusPending:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.VpcLinkStatusDeleting:
		item.Health = sdp.Health_HEALTH_PENDING.Enum()
	case types.VpcLinkStatusFailed:
		item.Health = sdp.Health_HEALTH_ERROR.Enum()
	}

	for _, targetArn := range awsItem.TargetArns {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "elbv2-load-balancer",
				Method: sdp.QueryMethod_SEARCH,
				Query:  targetArn,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Any change on the load balancer will affect the VPC link
				In: true,
				// Any change on the VPC link won't affect the load balancer
				Out: false,
			},
		})
	}

	return &item, nil
}

func NewAPIGatewayVpcLinkAdapter(client *apigateway.Client, accountID string, region string) *adapterhelpers.GetListAdapter[*types.VpcLink, *apigateway.Client, *apigateway.Options] {
	return &adapterhelpers.GetListAdapter[*types.VpcLink, *apigateway.Client, *apigateway.Options]{
		ItemType:        "apigateway-vpc-link",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		AdapterMetadata: vpcLinkAdapterMetadata,
		GetFunc: func(ctx context.Context, client *apigateway.Client, scope, query string) (*types.VpcLink, error) {
			out, err := client.GetVpcLink(ctx, &apigateway.GetVpcLinkInput{
				VpcLinkId: &query,
			})
			if err != nil {
				return nil, err
			}
			return convertGetVpcLinkOutputToVpcLink(out), nil
		},
		ListFunc: vpcLinkListFunc,
		SearchFunc: func(ctx context.Context, client *apigateway.Client, scope string, query string) ([]*types.VpcLink, error) {
			out, err := client.GetVpcLinks(ctx, &apigateway.GetVpcLinksInput{})
			if err != nil {
				return nil, err
			}

			var items []*types.VpcLink
			for _, vpcLink := range out.Items {
				if strings.Contains(*vpcLink.Name, query) {
					items = append(items, &vpcLink)
				}
			}

			return items, nil
		},
		ItemMapper: func(_, scope string, awsItem *types.VpcLink) (*sdp.Item, error) {
			return vpcLinkOutputMapper(scope, awsItem)
		},
	}
}

var vpcLinkAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:            "apigateway-vpc-link",
	DescriptiveName: "VPC Link",
	Category:        sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
	SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
		Get:               true,
		List:              true,
		Search:            true,
		GetDescription:    "Get a VPC Link by ID",
		ListDescription:   "List all VPC Links",
		SearchDescription: "Search for VPC Links by their name",
	},
})
