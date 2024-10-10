package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestRouterConfigurationOutputMapper(t *testing.T) {
	output := &directconnect.DescribeRouterConfigurationOutput{
		CustomerRouterConfig: adapters.PtrString("some config"),
		Router: &types.RouterType{
			Platform:                  adapters.PtrString("2900 Series Routers"),
			RouterTypeIdentifier:      adapters.PtrString("CiscoSystemsInc-2900SeriesRouters-IOS124"),
			Software:                  adapters.PtrString("IOS 12.4+"),
			Vendor:                    adapters.PtrString("Cisco Systems, Inc."),
			XsltTemplateName:          adapters.PtrString("customer-router-cisco-generic.xslt"),
			XsltTemplateNameForMacSec: adapters.PtrString(""),
		},
		VirtualInterfaceId:   adapters.PtrString("dxvif-ffhhk74f"),
		VirtualInterfaceName: adapters.PtrString("PrivateVirtualInterface"),
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

	tests := adapters.QueryTests{
		{
			ExpectedType:   "directconnect-virtual-interface",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxvif-ffhhk74f",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewRouterConfigurationAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewRouterConfigurationAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter:  adapter,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}
