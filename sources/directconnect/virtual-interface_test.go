package directconnect

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestVirtualInterfaceOutputMapper(t *testing.T) {
	output := &directconnect.DescribeVirtualInterfacesOutput{
		VirtualInterfaces: []types.VirtualInterface{
			{
				VirtualInterfaceId:     sources.PtrString("dxvif-ffhhk74f"),
				ConnectionId:           sources.PtrString("dxcon-fguhmqlc"),
				VirtualInterfaceState:  "verifying",
				CustomerAddress:        sources.PtrString("192.168.1.2/30"),
				AmazonAddress:          sources.PtrString("192.168.1.1/30"),
				VirtualInterfaceType:   sources.PtrString("private"),
				VirtualInterfaceName:   sources.PtrString("PrivateVirtualInterface"),
				DirectConnectGatewayId: sources.PtrString("cf68415c-f4ae-48f2-87a7-3b52cexample"),
			},
		},
	}

	items, err := virtualInterfaceOutputMapper(context.Background(), nil, "foo", nil, output)
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
			ExpectedType:   "directconnect-connection",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxcon-fguhmqlc",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-direct-connect-gateway",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cf68415c-f4ae-48f2-87a7-3b52cexample",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "rdap-ip-network",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "192.168.1.1/30",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "rdap-ip-network",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "192.168.1.2/30",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "directconnect-direct-connect-gateway-attachment",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  fmt.Sprintf("%s/%s", "cf68415c-f4ae-48f2-87a7-3b52cexample", "dxvif-ffhhk74f"),
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-direct-connect-gateway-attachment",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "dxvif-ffhhk74f",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewVirtualInterfaceSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewVirtualInterfaceSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
