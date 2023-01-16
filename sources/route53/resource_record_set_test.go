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
		ResourceRecords: []types.ResourceRecord{
			{
				Value: sources.PtrString("ns-1673.awsdns-17.co.uk."),
			},
			{
				Value: sources.PtrString("ns-1505.awsdns-60.org."),
			},
			{
				Value: sources.PtrString("ns-955.awsdns-55.net."),
			},
			{
				Value: sources.PtrString("ns-276.awsdns-34.com."),
			},
		},
		AliasTarget: &types.AliasTarget{
			DNSName:              sources.PtrString("dnsName"),
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
		HealthCheckId:           sources.PtrString("id"),
		MultiValueAnswer:        sources.PtrBool(true),
		Region:                  types.ResourceRecordSetRegionApEast1,
		SetIdentifier:           sources.PtrString("identifier"),
		TrafficPolicyInstanceId: sources.PtrString("id"),
		Weight:                  sources.PtrInt64(100),
	}

	item, err := ResourceRecordSetItemMapper("foo", &recordSet)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: *sdp.RequestMethod_GET.Enum(),
			ExpectedQuery:  "dnsName",
			ExpectedScope:  "global",
		},
	}

	tests.Execute(t, item)
}

func TestNewResourceRecordSetSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewResourceRecordSetSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
		SkipGet: true,
	}

	test.Run(t)
}
