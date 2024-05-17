package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/overmindtech/sdp-go"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestLagHealth(t *testing.T) {
	cases := []struct {
		state  types.LagState
		health sdp.Health
	}{
		{
			state:  types.LagStateRequested,
			health: sdp.Health_HEALTH_PENDING,
		},
		{
			state:  types.LagStatePending,
			health: sdp.Health_HEALTH_PENDING,
		},
		{
			state:  types.LagStateAvailable,
			health: sdp.Health_HEALTH_OK,
		},
		{
			state:  types.LagStateDown,
			health: sdp.Health_HEALTH_ERROR,
		},
		{
			state:  types.LagStateDeleting,
			health: sdp.Health_HEALTH_UNKNOWN,
		},
		{
			state:  types.LagStateDeleted,
			health: sdp.Health_HEALTH_UNKNOWN,
		},
		{
			state:  types.LagStateUnknown,
			health: sdp.Health_HEALTH_UNKNOWN,
		},
	}

	for _, c := range cases {
		output := &directconnect.DescribeLagsOutput{
			Lags: []types.Lag{
				{
					LagState: c.state,
					LagId:    sources.PtrString("dxlag-fgsu9erb"),
				},
			},
		}

		items, err := lagOutputMapper(context.Background(), nil, "foo", nil, output)
		if err != nil {
			t.Fatal(err)
		}

		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %v", len(items))
		}

		item := items[0]

		if item.GetHealth() != c.health {
			t.Errorf("expected health to be %v, got: %v", c.health, item.GetHealth())
		}
	}
}

func TestLagOutputMapper(t *testing.T) {
	output := &directconnect.DescribeLagsOutput{
		Lags: []types.Lag{
			{
				AwsDeviceV2:         sources.PtrString("EqDC2-19y7z3m17xpuz"),
				NumberOfConnections: int32(2),
				LagState:            types.LagStateAvailable,
				OwnerAccount:        sources.PtrString("123456789012"),
				LagName:             sources.PtrString("DA-LAG"),
				Connections: []types.Connection{
					{
						OwnerAccount:    sources.PtrString("123456789012"),
						ConnectionId:    sources.PtrString("dxcon-ffnikghc"),
						LagId:           sources.PtrString("dxlag-fgsu9erb"),
						ConnectionState: "requested",
						Bandwidth:       sources.PtrString("10Gbps"),
						Location:        sources.PtrString("EqDC2"),
						ConnectionName:  sources.PtrString("Requested Connection 1 for Lag dxlag-fgsu9erb"),
						Region:          sources.PtrString("us-east-1"),
					},
					{
						OwnerAccount:    sources.PtrString("123456789012"),
						ConnectionId:    sources.PtrString("dxcon-fglgbdea"),
						LagId:           sources.PtrString("dxlag-fgsu9erb"),
						ConnectionState: "requested",
						Bandwidth:       sources.PtrString("10Gbps"),
						Location:        sources.PtrString("EqDC2"),
						ConnectionName:  sources.PtrString("Requested Connection 2 for Lag dxlag-fgsu9erb"),
						Region:          sources.PtrString("us-east-1"),
					},
				},
				LagId:                sources.PtrString("dxlag-fgsu9erb"),
				MinimumLinks:         int32(0),
				ConnectionsBandwidth: sources.PtrString("10Gbps"),
				Region:               sources.PtrString("us-east-1"),
				Location:             sources.PtrString("EqDC2"),
			},
		},
	}

	items, err := lagOutputMapper(context.Background(), nil, "foo", nil, output)
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
			ExpectedType:   "directconnect-connection",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxcon-ffnikghc",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-connection",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dxcon-fglgbdea",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-location",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "EqDC2",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "directconnect-hosted-connection",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "dxlag-fgsu9erb",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewLagSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewLagSource(client, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
