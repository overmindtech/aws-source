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

func TestDirectConnectGatewayAttachmentOutputMapper_Health_OK(t *testing.T) {
	output := &directconnect.DescribeDirectConnectGatewayAttachmentsOutput{
		DirectConnectGatewayAttachments: []types.DirectConnectGatewayAttachment{
			{
				VirtualInterfaceOwnerAccount: sources.PtrString("123456789012"),
				VirtualInterfaceRegion:       sources.PtrString("us-east-2"),
				VirtualInterfaceId:           sources.PtrString("dxvif-ffhhk74f"),
				DirectConnectGatewayId:       sources.PtrString("cf68415c-f4ae-48f2-87a7-3b52cexample"),
				AttachmentState:              "detaching",
			},
		},
	}

	items, err := directConnectGatewayAttachmentOutputMapper(context.Background(), nil, "foo", nil, output)
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

	if item.GetHealth() != sdp.Health_HEALTH_OK {
		t.Fatalf("expected health to be OK, got: %v", item.GetHealth())
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "directconnect-direct-connect-gateway",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cf68415c-f4ae-48f2-87a7-3b52cexample",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "directconnect-virtual-interface",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxvif-ffhhk74f",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestDirectConnectGatewayAttachmentOutputMapper_Health_Error(t *testing.T) {
	output := &directconnect.DescribeDirectConnectGatewayAttachmentsOutput{
		DirectConnectGatewayAttachments: []types.DirectConnectGatewayAttachment{
			{
				VirtualInterfaceOwnerAccount: sources.PtrString("123456789012"),
				VirtualInterfaceRegion:       sources.PtrString("us-east-2"),
				VirtualInterfaceId:           sources.PtrString("dxvif-ffhhk74f"),
				DirectConnectGatewayId:       sources.PtrString("cf68415c-f4ae-48f2-87a7-3b52cexample"),
				AttachmentState:              "detaching",
				StateChangeError:             sources.PtrString("error"),
			},
		},
	}

	items, err := directConnectGatewayAttachmentOutputMapper(context.Background(), nil, "foo", nil, output)
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

	if item.GetHealth() != sdp.Health_HEALTH_ERROR {
		t.Fatalf("expected health to be ERROR, got: %v", item.GetHealth())
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "directconnect-direct-connect-gateway",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cf68415c-f4ae-48f2-87a7-3b52cexample",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "directconnect-virtual-interface",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxvif-ffhhk74f",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewDirectConnectGatewayAttachmentSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewDirectConnectGatewayAttachmentSource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
