package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func originRequestPolicyItemMapper(scope string, awsItem *types.OriginRequestPolicy) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-origin-request-policy",
		UniqueAttribute: "id",
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

func NewOriginRequestPolicySource(config aws.Config, accountID string) *sources.GetListSource[*types.OriginRequestPolicy, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.OriginRequestPolicy, *cloudfront.Client, *cloudfront.Options]{
		ItemType:  "cloudfront-origin-request-policy",
		Client:    cloudfront.NewFromConfig(config),
		AccountID: accountID,
		Region:    "", // Cloudfront resources aren't tied to a region
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

			policies := make([]*types.OriginRequestPolicy, len(out.OriginRequestPolicyList.Items))

			for i, policy := range out.OriginRequestPolicyList.Items {
				policies[i] = policy.OriginRequestPolicy
			}

			return policies, nil
		},
		ItemMapper: originRequestPolicyItemMapper,
	}
}
