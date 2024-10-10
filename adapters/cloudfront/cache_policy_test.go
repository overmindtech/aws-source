package cloudfront

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
)

var testCachePolicy = &types.CachePolicy{
	Id:               adapters.PtrString("test-id"),
	LastModifiedTime: adapters.PtrTime(time.Now()),
	CachePolicyConfig: &types.CachePolicyConfig{
		MinTTL:     adapters.PtrInt64(1),
		Name:       adapters.PtrString("test-name"),
		Comment:    adapters.PtrString("test-comment"),
		DefaultTTL: adapters.PtrInt64(1),
		MaxTTL:     adapters.PtrInt64(1),
		ParametersInCacheKeyAndForwardedToOrigin: &types.ParametersInCacheKeyAndForwardedToOrigin{
			CookiesConfig: &types.CachePolicyCookiesConfig{
				CookieBehavior: types.CachePolicyCookieBehaviorAll,
				Cookies: &types.CookieNames{
					Quantity: adapters.PtrInt32(1),
					Items: []string{
						"test-cookie",
					},
				},
			},
			EnableAcceptEncodingGzip: adapters.PtrBool(true),
			HeadersConfig: &types.CachePolicyHeadersConfig{
				HeaderBehavior: types.CachePolicyHeaderBehaviorWhitelist,
				Headers: &types.Headers{
					Quantity: adapters.PtrInt32(1),
					Items: []string{
						"test-header",
					},
				},
			},
			QueryStringsConfig: &types.CachePolicyQueryStringsConfig{
				QueryStringBehavior: types.CachePolicyQueryStringBehaviorWhitelist,
				QueryStrings: &types.QueryStringNames{
					Quantity: adapters.PtrInt32(1),
					Items: []string{
						"test-query-string",
					},
				},
			},
			EnableAcceptEncodingBrotli: adapters.PtrBool(true),
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
	client, account, _ := GetAutoConfig(t)

	source := NewCachePolicySource(client, account)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
