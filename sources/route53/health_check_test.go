package route53

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestHealthCheckItemMapper(t *testing.T) {
	hc := HealthCheck{
		HealthCheck: types.HealthCheck{
			Id:              sources.PtrString("d7ce5d72-6d1f-4147-8246-d0ca3fb505d6"),
			CallerReference: sources.PtrString("85d56b3f-873c-498b-a2dd-554ec13c5289"),
			HealthCheckConfig: &types.HealthCheckConfig{
				IPAddress:                sources.PtrString("1.1.1.1"),
				Port:                     sources.PtrInt32(443),
				Type:                     types.HealthCheckTypeHttps,
				FullyQualifiedDomainName: sources.PtrString("one.one.one.one"),
				RequestInterval:          sources.PtrInt32(30),
				FailureThreshold:         sources.PtrInt32(3),
				MeasureLatency:           sources.PtrBool(false),
				Inverted:                 sources.PtrBool(false),
				Disabled:                 sources.PtrBool(false),
				EnableSNI:                sources.PtrBool(true),
			},
			HealthCheckVersion: sources.PtrInt64(1),
		},
		HealthCheckObservations: []types.HealthCheckObservation{
			{
				Region:    types.HealthCheckRegionApNortheast1,
				IPAddress: sources.PtrString("15.177.62.21"),
				StatusReport: &types.StatusReport{
					Status:      sources.PtrString("Success: HTTP Status Code 200, OK"),
					CheckedTime: sources.PtrTime(time.Now()),
				},
			},
			{
				Region:    types.HealthCheckRegionEuWest1,
				IPAddress: sources.PtrString("15.177.10.21"),
				StatusReport: &types.StatusReport{
					Status:      sources.PtrString("Failure: Connection timed out. The endpoint or the internet connection is down, or requests are being blocked by your firewall. See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-router-firewall-rules.html"),
					CheckedTime: sources.PtrTime(time.Now()),
				},
			},
		},
	}

	item, err := healthCheckItemMapper("foo", &hc)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "cloudwatch-alarm",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "{\"MetricName\":\"HealthCheckStatus\",\"Namespace\":\"AWS/Route53\",\"Dimensions\":[{\"Name\":\"HealthCheckId\",\"Value\":\"d7ce5d72-6d1f-4147-8246-d0ca3fb505d6\"}],\"ExtendedStatistic\":null,\"Period\":null,\"Statistic\":\"\",\"Unit\":\"\"}",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewHealthCheckSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewHealthCheckSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
