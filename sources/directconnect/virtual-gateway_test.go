package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestVirtualGatewayOutputMapper(t *testing.T) {
	output := &directconnect.DescribeVirtualGatewaysOutput{
		VirtualGateways: []types.VirtualGateway{
			{
				VirtualGatewayId:    sources.PtrString("cf68415c-f4ae-48f2-87a7-3b52cexample"),
				VirtualGatewayState: sources.PtrString("available"),
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

func TestNewVirtualGatewaySource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewVirtualGatewaySource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
