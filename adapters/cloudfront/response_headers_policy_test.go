package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestResponseHeadersPolicyItemMapper(t *testing.T) {
	x := types.ResponseHeadersPolicy{
		Id:               adapters.PtrString("test"),
		LastModifiedTime: adapters.PtrTime(time.Now()),
		ResponseHeadersPolicyConfig: &types.ResponseHeadersPolicyConfig{
			Name:    adapters.PtrString("example-policy"),
			Comment: adapters.PtrString("example comment"),
			CorsConfig: &types.ResponseHeadersPolicyCorsConfig{
				AccessControlAllowCredentials: adapters.PtrBool(true),
				AccessControlAllowHeaders: &types.ResponseHeadersPolicyAccessControlAllowHeaders{
					Items:    []string{"X-Customer-Header"},
					Quantity: adapters.PtrInt32(1),
				},
			},
			CustomHeadersConfig: &types.ResponseHeadersPolicyCustomHeadersConfig{
				Quantity: adapters.PtrInt32(1),
				Items: []types.ResponseHeadersPolicyCustomHeader{
					{
						Header:   adapters.PtrString("X-Customer-Header"),
						Override: adapters.PtrBool(true),
						Value:    adapters.PtrString("test"),
					},
				},
			},
			RemoveHeadersConfig: &types.ResponseHeadersPolicyRemoveHeadersConfig{
				Quantity: adapters.PtrInt32(1),
				Items: []types.ResponseHeadersPolicyRemoveHeader{
					{
						Header: adapters.PtrString("X-Private-Header"),
					},
				},
			},
			SecurityHeadersConfig: &types.ResponseHeadersPolicySecurityHeadersConfig{
				ContentSecurityPolicy: &types.ResponseHeadersPolicyContentSecurityPolicy{
					ContentSecurityPolicy: adapters.PtrString("default-src 'none';"),
					Override:              adapters.PtrBool(true),
				},
				ContentTypeOptions: &types.ResponseHeadersPolicyContentTypeOptions{
					Override: adapters.PtrBool(true),
				},
				FrameOptions: &types.ResponseHeadersPolicyFrameOptions{
					FrameOption: types.FrameOptionsListDeny,
					Override:    adapters.PtrBool(true),
				},
				ReferrerPolicy: &types.ResponseHeadersPolicyReferrerPolicy{
					Override:       adapters.PtrBool(true),
					ReferrerPolicy: types.ReferrerPolicyListNoReferrer,
				},
				StrictTransportSecurity: &types.ResponseHeadersPolicyStrictTransportSecurity{
					AccessControlMaxAgeSec: adapters.PtrInt32(86400),
					Override:               adapters.PtrBool(true),
					IncludeSubdomains:      adapters.PtrBool(true),
					Preload:                adapters.PtrBool(true),
				},
				XSSProtection: &types.ResponseHeadersPolicyXSSProtection{
					Override:   adapters.PtrBool(true),
					Protection: adapters.PtrBool(true),
					ModeBlock:  adapters.PtrBool(true),
					ReportUri:  adapters.PtrString("https://example.com/report"),
				},
			},
			ServerTimingHeadersConfig: &types.ResponseHeadersPolicyServerTimingHeadersConfig{
				Enabled:      adapters.PtrBool(true),
				SamplingRate: adapters.PtrFloat64(0.1),
			},
		},
	}

	item, err := ResponseHeadersPolicyItemMapper("", "test", &x)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewResponseHeadersPolicyAdapter(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	adapter := NewResponseHeadersPolicyAdapter(client, account)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
