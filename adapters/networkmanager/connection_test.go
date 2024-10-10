package networkmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestConnectionOutputMapper(t *testing.T) {
	output := networkmanager.GetConnectionsOutput{
		Connections: []types.Connection{
			{
				GlobalNetworkId:   adapters.PtrString("default"),
				ConnectionId:      adapters.PtrString("conn-1"),
				DeviceId:          adapters.PtrString("dvc-1"),
				ConnectedDeviceId: adapters.PtrString("dvc-2"),
				LinkId:            adapters.PtrString("link-1"),
				ConnectedLinkId:   adapters.PtrString("link-2"),
			},
		},
	}
	scope := "123456789012.eu-west-2"
	items, err := connectionOutputMapper(context.Background(), &networkmanager.Client{}, scope, &networkmanager.GetConnectionsInput{}, &output)

	if err != nil {
		t.Error(err)
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

	// Ensure unique attribute
	err = item.Validate()

	if err != nil {
		t.Error(err)
	}

	if item.UniqueAttributeValue() != "default|conn-1" {
		t.Fatalf("expected default|conn-1, got %v", item.UniqueAttributeValue())
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "networkmanager-global-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-device",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default|dvc-1",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-device",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default|dvc-2",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-link",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default|link-1",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-link",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default|link-2",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}
