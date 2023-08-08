package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

type CloudFrontClient interface {
	GetCachePolicy(ctx context.Context, params *cloudfront.GetCachePolicyInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetCachePolicyOutput, error)
	ListCachePolicies(ctx context.Context, params *cloudfront.ListCachePoliciesInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListCachePoliciesOutput, error)

	GetDistribution(ctx context.Context, params *cloudfront.GetDistributionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetDistributionOutput, error)
	ListDistributions(ctx context.Context, params *cloudfront.ListDistributionsInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error)
}
