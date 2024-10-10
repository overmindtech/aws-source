package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestOriginRequestPolicyItemMapper(t *testing.T) {
	x := types.OriginRequestPolicy{
		Id:               adapters.PtrString("test"),
		LastModifiedTime: adapters.PtrTime(time.Now()),
		OriginRequestPolicyConfig: &types.OriginRequestPolicyConfig{
			Name:    adapters.PtrString("example-policy"),
			Comment: adapters.PtrString("example comment"),
			QueryStringsConfig: &types.OriginRequestPolicyQueryStringsConfig{
				QueryStringBehavior: types.OriginRequestPolicyQueryStringBehaviorAllExcept,
				QueryStrings: &types.QueryStringNames{
					Quantity: adapters.PtrInt32(1),
					Items:    []string{"test"},
				},
			},
			CookiesConfig: &types.OriginRequestPolicyCookiesConfig{
				CookieBehavior: types.OriginRequestPolicyCookieBehaviorAll,
				Cookies: &types.CookieNames{
					Quantity: adapters.PtrInt32(1),
					Items:    []string{"test"},
				},
			},
			HeadersConfig: &types.OriginRequestPolicyHeadersConfig{
				HeaderBehavior: types.OriginRequestPolicyHeaderBehaviorAllViewer,
				Headers: &types.Headers{
					Quantity: adapters.PtrInt32(1),
					Items:    []string{"test"},
				},
			},
		},
	}

	item, err := originRequestPolicyItemMapper("", "test", &x)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewOriginRequestPolicySource(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	source := NewOriginRequestPolicySource(client, account)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
