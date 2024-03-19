package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestRouterConfigurationOutputMapper(t *testing.T) {
	output := &directconnect.DescribeRouterConfigurationOutput{
		CustomerRouterConfig: sources.PtrString("some config"),
		Router: &types.RouterType{
			Platform:                  sources.PtrString("2900 Series Routers"),
			RouterTypeIdentifier:      sources.PtrString("CiscoSystemsInc-2900SeriesRouters-IOS124"),
			Software:                  sources.PtrString("IOS 12.4+"),
			Vendor:                    sources.PtrString("Cisco Systems, Inc."),
			XsltTemplateName:          sources.PtrString("customer-router-cisco-generic.xslt"),
			XsltTemplateNameForMacSec: sources.PtrString(""),
		},
		VirtualInterfaceId:   sources.PtrString("dxvif-ffhhk74f"),
		VirtualInterfaceName: sources.PtrString("PrivateVirtualInterface"),
	}

	items, err := routerConfigurationOutputMapper(context.Background(), nil, "foo", nil, output)
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	tests := sources.QueryTests{
		{
			ExpectedType:   "directconnect-virtual-interface",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxvif-ffhhk74f",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewRouterConfigurationSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewRouterConfigurationSource(client, account, region, &TestRateLimit)

	test := sources.E2ETest{
		Source:   source,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}
