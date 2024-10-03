package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestResponseHeadersPolicyItemMapper(t *testing.T) {
	x := types.ResponseHeadersPolicy{
		Id:               sources.PtrString("test"),
		LastModifiedTime: sources.PtrTime(time.Now()),
		ResponseHeadersPolicyConfig: &types.ResponseHeadersPolicyConfig{
			Name:    sources.PtrString("example-policy"),
			Comment: sources.PtrString("example comment"),
			CorsConfig: &types.ResponseHeadersPolicyCorsConfig{
				AccessControlAllowCredentials: sources.PtrBool(true),
				AccessControlAllowHeaders: &types.ResponseHeadersPolicyAccessControlAllowHeaders{
					Items:    []string{"X-Customer-Header"},
					Quantity: sources.PtrInt32(1),
				},
			},
			CustomHeadersConfig: &types.ResponseHeadersPolicyCustomHeadersConfig{
				Quantity: sources.PtrInt32(1),
				Items: []types.ResponseHeadersPolicyCustomHeader{
					{
						Header:   sources.PtrString("X-Customer-Header"),
						Override: sources.PtrBool(true),
						Value:    sources.PtrString("test"),
					},
				},
			},
			RemoveHeadersConfig: &types.ResponseHeadersPolicyRemoveHeadersConfig{
				Quantity: sources.PtrInt32(1),
				Items: []types.ResponseHeadersPolicyRemoveHeader{
					{
						Header: sources.PtrString("X-Private-Header"),
					},
				},
			},
			SecurityHeadersConfig: &types.ResponseHeadersPolicySecurityHeadersConfig{
				ContentSecurityPolicy: &types.ResponseHeadersPolicyContentSecurityPolicy{
					ContentSecurityPolicy: sources.PtrString("default-src 'none';"),
					Override:              sources.PtrBool(true),
				},
				ContentTypeOptions: &types.ResponseHeadersPolicyContentTypeOptions{
					Override: sources.PtrBool(true),
				},
				FrameOptions: &types.ResponseHeadersPolicyFrameOptions{
					FrameOption: types.FrameOptionsListDeny,
					Override:    sources.PtrBool(true),
				},
				ReferrerPolicy: &types.ResponseHeadersPolicyReferrerPolicy{
					Override:       sources.PtrBool(true),
					ReferrerPolicy: types.ReferrerPolicyListNoReferrer,
				},
				StrictTransportSecurity: &types.ResponseHeadersPolicyStrictTransportSecurity{
					AccessControlMaxAgeSec: sources.PtrInt32(86400),
					Override:               sources.PtrBool(true),
					IncludeSubdomains:      sources.PtrBool(true),
					Preload:                sources.PtrBool(true),
				},
				XSSProtection: &types.ResponseHeadersPolicyXSSProtection{
					Override:   sources.PtrBool(true),
					Protection: sources.PtrBool(true),
					ModeBlock:  sources.PtrBool(true),
					ReportUri:  sources.PtrString("https://example.com/report"),
				},
			},
			ServerTimingHeadersConfig: &types.ResponseHeadersPolicyServerTimingHeadersConfig{
				Enabled:      sources.PtrBool(true),
				SamplingRate: sources.PtrFloat64(0.1),
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

func TestNewResponseHeadersPolicySource(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	source := NewResponseHeadersPolicySource(client, account)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
