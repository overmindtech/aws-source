package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func originRequestPolicyItemMapper(_, scope string, awsItem *types.OriginRequestPolicy) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-origin-request-policy",
		UniqueAttribute: "Id",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-origin-request-policy
// +overmind:descriptiveType CloudFront Origin Request Policy
// +overmind:get Get Origin Request Policy by ID
// +overmind:list List Origin Request Policies
// +overmind:search Origin Request Policy by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_origin_request_policy.id

func NewOriginRequestPolicySource(client *cloudfront.Client, accountID string) *sources.GetListSource[*types.OriginRequestPolicy, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.OriginRequestPolicy, *cloudfront.Client, *cloudfront.Options]{
		ItemType:        "cloudfront-origin-request-policy",
		Client:          client,
		AccountID:       accountID,
		Region:          "", // Cloudfront resources aren't tied to a region
		AdapterMetadata: OriginRequestPolicySourceMetadata(),
		GetFunc: func(ctx context.Context, client *cloudfront.Client, scope, query string) (*types.OriginRequestPolicy, error) {
			out, err := client.GetOriginRequestPolicy(ctx, &cloudfront.GetOriginRequestPolicyInput{
				Id: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.OriginRequestPolicy, nil
		},
		ListFunc: func(ctx context.Context, client *cloudfront.Client, scope string) ([]*types.OriginRequestPolicy, error) {
			out, err := client.ListOriginRequestPolicies(ctx, &cloudfront.ListOriginRequestPoliciesInput{})

			if err != nil {
				return nil, err
			}

			policies := make([]*types.OriginRequestPolicy, 0, len(out.OriginRequestPolicyList.Items))

			for _, policy := range out.OriginRequestPolicyList.Items {
				policies = append(policies, policy.OriginRequestPolicy)
			}

			return policies, nil
		},
		ItemMapper: originRequestPolicyItemMapper,
	}
}

func OriginRequestPolicySourceMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "cloudfront-origin-request-policy",
		DescriptiveName: "CloudFront Origin Request Policy",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get Origin Request Policy by ID",
			ListDescription:   "List Origin Request Policies",
			SearchDescription: "Origin Request Policy by ARN",
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_cloudfront_origin_request_policy.id"},
		},
	}
}
