package route53

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestHealthCheckItemMapper(t *testing.T) {
	hc := HealthCheck{
		HealthCheck: types.HealthCheck{
			Id:              adapters.PtrString("d7ce5d72-6d1f-4147-8246-d0ca3fb505d6"),
			CallerReference: adapters.PtrString("85d56b3f-873c-498b-a2dd-554ec13c5289"),
			HealthCheckConfig: &types.HealthCheckConfig{
				IPAddress:                adapters.PtrString("1.1.1.1"),
				Port:                     adapters.PtrInt32(443),
				Type:                     types.HealthCheckTypeHttps,
				FullyQualifiedDomainName: adapters.PtrString("one.one.one.one"),
				RequestInterval:          adapters.PtrInt32(30),
				FailureThreshold:         adapters.PtrInt32(3),
				MeasureLatency:           adapters.PtrBool(false),
				Inverted:                 adapters.PtrBool(false),
				Disabled:                 adapters.PtrBool(false),
				EnableSNI:                adapters.PtrBool(true),
			},
			HealthCheckVersion: adapters.PtrInt64(1),
		},
		HealthCheckObservations: []types.HealthCheckObservation{
			{
				Region:    types.HealthCheckRegionApNortheast1,
				IPAddress: adapters.PtrString("15.177.62.21"),
				StatusReport: &types.StatusReport{
					Status:      adapters.PtrString("Success: HTTP Status Code 200, OK"),
					CheckedTime: adapters.PtrTime(time.Now()),
				},
			},
			{
				Region:    types.HealthCheckRegionEuWest1,
				IPAddress: adapters.PtrString("15.177.10.21"),
				StatusReport: &types.StatusReport{
					Status:      adapters.PtrString("Failure: Connection timed out. The endpoint or the internet connection is down, or requests are being blocked by your firewall. See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-router-firewall-rules.html"),
					CheckedTime: adapters.PtrTime(time.Now()),
				},
			},
		},
	}

	item, err := healthCheckItemMapper("", "foo", &hc)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "cloudwatch-alarm",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "{\"MetricName\":\"HealthCheckStatus\",\"Namespace\":\"AWS/Route53\",\"Dimensions\":[{\"Name\":\"HealthCheckId\",\"Value\":\"d7ce5d72-6d1f-4147-8246-d0ca3fb505d6\"}],\"ExtendedStatistic\":null,\"Period\":null,\"Statistic\":\"\",\"Unit\":\"\"}",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewHealthCheckAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewHealthCheckAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
