package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestVirtualGatewayOutputMapper(t *testing.T) {
	output := &directconnect.DescribeVirtualGatewaysOutput{
		VirtualGateways: []types.VirtualGateway{
			{
				VirtualGatewayId:    adapters.PtrString("cf68415c-f4ae-48f2-87a7-3b52cexample"),
				VirtualGatewayState: adapters.PtrString("available"),
			},
		},
	}

	items, err := virtualGatewayOutputMapper(context.Background(), nil, "foo", nil, output)
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
}

func TestNewVirtualGatewayAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewVirtualGatewayAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
