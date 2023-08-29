package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func ResponseHeadersPolicyItemMapper(scope string, awsItem *types.ResponseHeadersPolicy) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-response-headers-policy",
		UniqueAttribute: "id",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-response-headers-policy
// +overmind:descriptiveType CloudFront Response Headers Policy
// +overmind:get Get Response Headers Policy by ID
// +overmind:list List Response Headers Policies
// +overmind:search Response Headers Policy by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_response_headers_policy.id

func NewResponseHeadersPolicySource(config aws.Config, accountID string) *sources.GetListSource[*types.ResponseHeadersPolicy, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.ResponseHeadersPolicy, *cloudfront.Client, *cloudfront.Options]{
		ItemType:  "cloudfront-response-headers-policy",
		Client:    cloudfront.NewFromConfig(config),
		AccountID: accountID,
		Region:    "global",
		GetFunc: func(ctx context.Context, client *cloudfront.Client, scope, query string) (*types.ResponseHeadersPolicy, error) {
			out, err := client.GetResponseHeadersPolicy(ctx, &cloudfront.GetResponseHeadersPolicyInput{
				Id: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.ResponseHeadersPolicy, nil
		},
		ListFunc: func(ctx context.Context, client *cloudfront.Client, scope string) ([]*types.ResponseHeadersPolicy, error) {
			out, err := client.ListResponseHeadersPolicies(ctx, &cloudfront.ListResponseHeadersPoliciesInput{})

			if err != nil {
				return nil, err
			}

			policies := make([]*types.ResponseHeadersPolicy, len(out.ResponseHeadersPolicyList.Items))

			for i, policy := range out.ResponseHeadersPolicyList.Items {
				policies[i] = policy.ResponseHeadersPolicy
			}

			return policies, nil
		},
		ItemMapper: ResponseHeadersPolicyItemMapper,
	}
}
