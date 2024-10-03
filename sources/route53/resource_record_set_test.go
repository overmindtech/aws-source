package route53

import (
	"context"
	"fmt"
	"strings"
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

	item, err := resourceRecordSetItemMapper("", "foo", &recordSet)

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

	zoneSource := NewHostedZoneSource(client, account, region)

	zones, err := zoneSource.List(context.Background(), zoneSource.Scopes()[0], true)
	if err != nil {
		t.Fatal(err)
	}

	if len(zones) == 0 {
		t.Skip("no zones found")
	}

	source := NewResourceRecordSetSource(client, account, region)

	search := zones[0].UniqueAttributeValue()
	test := sources.E2ETest{
		Adapter:         source,
		Timeout:         10 * time.Second,
		SkipGet:         true,
		GoodSearchQuery: &search,
	}

	test.Run(t)

	items, err := source.Search(context.Background(), zoneSource.Scopes()[0], search, true)
	if err != nil {
		t.Fatal(err)
	}

	numItems := len(items)

	rawZone := strings.TrimPrefix(search, "/hostedzone/")

	items, err = source.Search(context.Background(), zoneSource.Scopes()[0], rawZone, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != numItems {
		t.Errorf("expected %d items, got %d", numItems, len(items))
	}

	if len(items) > 0 {
		item := items[0]

		// Construct a terraform style ID
		name, _ := item.GetAttributes().Get("Name")
		typ, _ := item.GetAttributes().Get("Type")
		search = fmt.Sprintf("%s_%s_%s", rawZone, name, typ)

		items, err := source.Search(context.Background(), zoneSource.Scopes()[0], search, true)
		if err != nil {
			t.Fatal(err)
		}

		if len(items) != 1 {
			t.Errorf("expected 1 item, got %d", len(items))
		}
	}
}
