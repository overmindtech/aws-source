package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestContinuousDeploymentPolicyItemMapper(t *testing.T) {
	item, err := continuousDeploymentPolicyItemMapper("test", &types.ContinuousDeploymentPolicy{
		Id:               sources.PtrString("test-id"),
		LastModifiedTime: sources.PtrTime(time.Now()),
		ContinuousDeploymentPolicyConfig: &types.ContinuousDeploymentPolicyConfig{
			Enabled: sources.PtrBool(true),
			StagingDistributionDnsNames: &types.StagingDistributionDnsNames{
				Quantity: sources.PtrInt32(1),
				Items: []string{
					"staging.test.com", // link
				},
			},
			TrafficConfig: &types.TrafficConfig{
				Type: types.ContinuousDeploymentPolicyTypeSingleWeight,
				SingleHeaderConfig: &types.ContinuousDeploymentSingleHeaderConfig{
					Header: sources.PtrString("test-header"),
					Value:  sources.PtrString("test-value"),
				},
				SingleWeightConfig: &types.ContinuousDeploymentSingleWeightConfig{
					Weight: sources.PtrFloat32(1),
					SessionStickinessConfig: &types.SessionStickinessConfig{
						IdleTTL:    sources.PtrInt32(1),
						MaximumTTL: sources.PtrInt32(2),
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "staging.test.com",
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewContinuousDeploymentPolicySource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewContinuousDeploymentPolicySource(config, account)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
