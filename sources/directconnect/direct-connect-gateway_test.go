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

func TestDirectConnectGatewayOutputMapper_Health_OK(t *testing.T) {
	output := &directconnect.DescribeDirectConnectGatewaysOutput{
		DirectConnectGateways: []types.DirectConnectGateway{
			{
				AmazonSideAsn:             sources.PtrInt64(64512),
				DirectConnectGatewayId:    sources.PtrString("cf68415c-f4ae-48f2-87a7-3b52cexample"),
				OwnerAccount:              sources.PtrString("123456789012"),
				DirectConnectGatewayName:  sources.PtrString("DxGateway2"),
				DirectConnectGatewayState: types.DirectConnectGatewayStateAvailable,
			},
		},
	}

	items, err := directConnectGatewayOutputMapper(context.Background(), nil, "foo", nil, output)
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

	if items[0].GetHealth() != sdp.Health_HEALTH_OK {
		t.Fatalf("expected health to be OK, got: %v", items[0].GetHealth())
	}
}

func TestDirectConnectGatewayOutputMapper_Health_ERROR(t *testing.T) {
	output := &directconnect.DescribeDirectConnectGatewaysOutput{
		DirectConnectGateways: []types.DirectConnectGateway{
			{
				AmazonSideAsn:             sources.PtrInt64(64512),
				DirectConnectGatewayId:    sources.PtrString("cf68415c-f4ae-48f2-87a7-3b52cexample"),
				OwnerAccount:              sources.PtrString("123456789012"),
				DirectConnectGatewayName:  sources.PtrString("DxGateway2"),
				DirectConnectGatewayState: types.DirectConnectGatewayStateAvailable,
				StateChangeError:          sources.PtrString("error"),
			},
		},
	}

	items, err := directConnectGatewayOutputMapper(context.Background(), nil, "foo", nil, output)
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

	if items[0].GetHealth() != sdp.Health_HEALTH_ERROR {
		t.Fatalf("expected health to be ERROR, got: %v", items[0].GetHealth())
	}
}

func TestNewDirectConnectGatewaySource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewDirectConnectGatewaySource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
