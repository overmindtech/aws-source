package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func cachePolicyListFunc(ctx context.Context, client CloudFrontClient, scope string) ([]*types.CachePolicy, error) {
	var policyType types.CachePolicyType

	switch scope {
	case "aws":
		policyType = types.CachePolicyTypeManaged
	default:
		policyType = types.CachePolicyTypeCustom
	}

	out, err := client.ListCachePolicies(ctx, &cloudfront.ListCachePoliciesInput{
		Type: policyType,
	})

	if err != nil {
		return nil, err
	}

	policies := make([]*types.CachePolicy, 0, len(out.CachePolicyList.Items))

	for i := range out.CachePolicyList.Items {
		policies[i] = out.CachePolicyList.Items[i].CachePolicy
	}

	return policies, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-cache-policy
// +overmind:descriptiveType CloudFront Cache Policy
// +overmind:get Get a CloudFront Cache Policy
// +overmind:list List CloudFront Cache Policies
// +overmind:search Search CloudFront Cache Policies by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_cache_policy.id

func NewCachePolicySource(client CloudFrontClient, accountID string) *sources.GetListSource[*types.CachePolicy, CloudFrontClient, *cloudfront.Options] {
	return &sources.GetListSource[*types.CachePolicy, CloudFrontClient, *cloudfront.Options]{
		ItemType:               "cloudfront-cache-policy",
		Client:                 client,
		AccountID:              accountID,
		Region:                 "",   // Cloudfront resources aren't tied to a region
		SupportGlobalResources: true, // Some policies are global
		GetFunc: func(ctx context.Context, client CloudFrontClient, scope, query string) (*types.CachePolicy, error) {
			out, err := client.GetCachePolicy(ctx, &cloudfront.GetCachePolicyInput{
				Id: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.CachePolicy, nil
		},
		ListFunc: cachePolicyListFunc,
		ItemMapper: func(scope string, awsItem *types.CachePolicy) (*sdp.Item, error) {
			attributes, err := sources.ToAttributesCase(awsItem)

			if err != nil {
				return nil, err
			}

			item := sdp.Item{
				Type:            "cloudfront-cache-policy",
				UniqueAttribute: "id",
				Attributes:      attributes,
				Scope:           scope,
			}

			return &item, nil
		},
	}
}
