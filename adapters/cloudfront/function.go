package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func functionItemMapper(_, scope string, awsItem *types.FunctionSummary) (*sdp.Item, error) {
	attributes, err := adapters.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-function",
		UniqueAttribute: "Name",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-function
// +overmind:descriptiveType CloudFront Function
// +overmind:get Get a CloudFront Function by name
// +overmind:list List CloudFront Functions
// +overmind:search Search CloudFront Functions by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_function.name

func NewFunctionAdapter(client *cloudfront.Client, accountID string) *adapters.GetListAdapter[*types.FunctionSummary, *cloudfront.Client, *cloudfront.Options] {
	return &adapters.GetListAdapter[*types.FunctionSummary, *cloudfront.Client, *cloudfront.Options]{
		ItemType:        "cloudfront-function",
		Client:          client,
		AccountID:       accountID,
		Region:          "", // Cloudfront resources aren't tied to a region
		AdapterMetadata: FunctionMetadata(),
		GetFunc: func(ctx context.Context, client *cloudfront.Client, scope, query string) (*types.FunctionSummary, error) {
			out, err := client.DescribeFunction(ctx, &cloudfront.DescribeFunctionInput{
				Name: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.FunctionSummary, nil
		},
		ListFunc: func(ctx context.Context, client *cloudfront.Client, scope string) ([]*types.FunctionSummary, error) {
			out, err := client.ListFunctions(ctx, &cloudfront.ListFunctionsInput{
				Stage: types.FunctionStageLive,
			})

			if err != nil {
				return nil, err
			}

			summaries := make([]*types.FunctionSummary, 0, len(out.FunctionList.Items))

			for _, item := range out.FunctionList.Items {
				summaries = append(summaries, &item)
			}

			return summaries, nil
		},
		ItemMapper: functionItemMapper,
	}
}

func FunctionMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "cloudfront-function",
		DescriptiveName: "CloudFront Function",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a CloudFront Function by name",
			ListDescription:   "List CloudFront Functions",
			SearchDescription: "Search CloudFront Functions by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_cloudfront_function.name"},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_COMPUTE_APPLICATION,
	}
}
