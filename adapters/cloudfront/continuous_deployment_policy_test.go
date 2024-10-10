package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestContinuousDeploymentPolicyItemMapper(t *testing.T) {
	item, err := continuousDeploymentPolicyItemMapper("", "test", &types.ContinuousDeploymentPolicy{
		Id:               adapters.PtrString("test-id"),
		LastModifiedTime: adapters.PtrTime(time.Now()),
		ContinuousDeploymentPolicyConfig: &types.ContinuousDeploymentPolicyConfig{
			Enabled: adapters.PtrBool(true),
			StagingDistributionDnsNames: &types.StagingDistributionDnsNames{
				Quantity: adapters.PtrInt32(1),
				Items: []string{
					"staging.test.com", // link
				},
			},
			TrafficConfig: &types.TrafficConfig{
				Type: types.ContinuousDeploymentPolicyTypeSingleWeight,
				SingleHeaderConfig: &types.ContinuousDeploymentSingleHeaderConfig{
					Header: adapters.PtrString("test-header"),
					Value:  adapters.PtrString("test-value"),
				},
				SingleWeightConfig: &types.ContinuousDeploymentSingleWeightConfig{
					Weight: adapters.PtrFloat32(1),
					SessionStickinessConfig: &types.SessionStickinessConfig{
						IdleTTL:    adapters.PtrInt32(1),
						MaximumTTL: adapters.PtrInt32(2),
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

	tests := adapters.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "staging.test.com",
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewContinuousDeploymentPolicySource(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	source := NewContinuousDeploymentPolicySource(client, account)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
