package cloudfront

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
)

var testCachePolicy = &types.CachePolicy{
	Id:               sources.PtrString("test-id"),
	LastModifiedTime: sources.PtrTime(time.Now()),
	CachePolicyConfig: &types.CachePolicyConfig{
		MinTTL:     sources.PtrInt64(1),
		Name:       sources.PtrString("test-name"),
		Comment:    sources.PtrString("test-comment"),
		DefaultTTL: sources.PtrInt64(1),
		MaxTTL:     sources.PtrInt64(1),
		ParametersInCacheKeyAndForwardedToOrigin: &types.ParametersInCacheKeyAndForwardedToOrigin{
			CookiesConfig: &types.CachePolicyCookiesConfig{
				CookieBehavior: types.CachePolicyCookieBehaviorAll,
				Cookies: &types.CookieNames{
					Quantity: sources.PtrInt32(1),
					Items: []string{
						"test-cookie",
					},
				},
			},
			EnableAcceptEncodingGzip: sources.PtrBool(true),
			HeadersConfig: &types.CachePolicyHeadersConfig{
				HeaderBehavior: types.CachePolicyHeaderBehaviorWhitelist,
				Headers: &types.Headers{
					Quantity: sources.PtrInt32(1),
					Items: []string{
						"test-header",
					},
				},
			},
			QueryStringsConfig: &types.CachePolicyQueryStringsConfig{
				QueryStringBehavior: types.CachePolicyQueryStringBehaviorWhitelist,
				QueryStrings: &types.QueryStringNames{
					Quantity: sources.PtrInt32(1),
					Items: []string{
						"test-query-string",
					},
				},
			},
			EnableAcceptEncodingBrotli: sources.PtrBool(true),
		},
	},
}

func (t TestCloudFrontClient) ListCachePolicies(ctx context.Context, params *cloudfront.ListCachePoliciesInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListCachePoliciesOutput, error) {
	return &cloudfront.ListCachePoliciesOutput{
		CachePolicyList: &types.CachePolicyList{
			Items: []types.CachePolicySummary{
				{
					Type:        types.CachePolicyTypeManaged,
					CachePolicy: testCachePolicy,
				},
			},
		},
	}, nil
}

func (t TestCloudFrontClient) GetCachePolicy(ctx context.Context, params *cloudfront.GetCachePolicyInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetCachePolicyOutput, error) {
	return &cloudfront.GetCachePolicyOutput{
		CachePolicy: testCachePolicy,
	}, nil
}

func TestCachePolicyListFunc(t *testing.T) {
	policies, err := cachePolicyListFunc(context.Background(), TestCloudFrontClient{}, "aws")

	if err != nil {
		t.Fatal(err)
	}

	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}
}

func TestNewCachePolicySource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewCachePolicySource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
