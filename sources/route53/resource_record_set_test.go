package route53

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestResourceRecordSetItemMapper(t *testing.T) {
	recordSet := types.ResourceRecordSet{
		Name: sources.PtrString("overmind-demo.com."),
		Type: types.RRTypeNs,
		TTL:  sources.PtrInt64(172800),
		GeoProximityLocation: &types.GeoProximityLocation{
			AWSRegion:      sources.PtrString("us-east-1"),
			Bias:           sources.PtrInt32(100),
			Coordinates:    &types.Coordinates{},
			LocalZoneGroup: sources.PtrString("group"),
		},
		ResourceRecords: []types.ResourceRecord{
			{
				Value: sources.PtrString("ns-1673.awsdns-17.co.uk."), // link
			},
			{
				Value: sources.PtrString("ns-1505.awsdns-60.org."), // link
			},
			{
				Value: sources.PtrString("ns-955.awsdns-55.net."), // link
			},
			{
				Value: sources.PtrString("ns-276.awsdns-34.com."), // link
			},
		},
		AliasTarget: &types.AliasTarget{
			DNSName:              sources.PtrString("foo.bar.com"), // link
			EvaluateTargetHealth: true,
			HostedZoneId:         sources.PtrString("id"),
		},
		CidrRoutingConfig: &types.CidrRoutingConfig{
			CollectionId: sources.PtrString("id"),
			LocationName: sources.PtrString("somewhere"),
		},
		Failover: types.ResourceRecordSetFailoverPrimary,
		GeoLocation: &types.GeoLocation{
			ContinentCode:   sources.PtrString("GB"),
			CountryCode:     sources.PtrString("GB"),
			SubdivisionCode: sources.PtrString("ENG"),
		},
		HealthCheckId:           sources.PtrString("id"), // link
		MultiValueAnswer:        sources.PtrBool(true),
		Region:                  types.ResourceRecordSetRegionApEast1,
		SetIdentifier:           sources.PtrString("identifier"),
		TrafficPolicyInstanceId: sources.PtrString("id"),
		Weight:                  sources.PtrInt64(100),
	}

	item, err := resourceRecordSetItemMapper("foo", &recordSet)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "foo.bar.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "ns-1673.awsdns-17.co.uk.",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "ns-1505.awsdns-60.org.",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "ns-955.awsdns-55.net.",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "ns-276.awsdns-34.com.",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "route53-health-check",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewResourceRecordSetSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewResourceRecordSetSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
		SkipGet: true,
	}

	test.Run(t)
}
