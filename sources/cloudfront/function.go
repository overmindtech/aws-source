package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func functionItemMapper(scope string, awsItem *types.FunctionSummary) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-function",
		UniqueAttribute: "name",
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

func NewFunctionSource(config aws.Config, accountID string) *sources.GetListSource[*types.FunctionSummary, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.FunctionSummary, *cloudfront.Client, *cloudfront.Options]{
		ItemType:  "cloudfront-function",
		Client:    cloudfront.NewFromConfig(config),
		AccountID: accountID,
		Region:    "global",
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

			summaries := make([]*types.FunctionSummary, len(out.FunctionList.Items))

			for i, item := range out.FunctionList.Items {
				summaries[i] = &item
			}

			return summaries, nil
		},
		ItemMapper: functionItemMapper,
	}
}
