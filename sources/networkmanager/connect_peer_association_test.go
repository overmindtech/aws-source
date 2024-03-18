package networkmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestConnectPeerAssociationsOutputMapper(t *testing.T) {
	output := networkmanager.GetConnectPeerAssociationsOutput{
		ConnectPeerAssociations: []types.ConnectPeerAssociation{
			{
				ConnectPeerId:   sources.PtrString("cp-1"),
				DeviceId:        sources.PtrString("dvc-1"),
				GlobalNetworkId: sources.PtrString("default"),
				LinkId:          sources.PtrString("link-1"),
			},
		},
	}
	scope := "123456789012.eu-west-2"
	items, err := connectPeerAssociationsOutputMapper(context.Background(), &networkmanager.Client{}, scope, &networkmanager.GetConnectPeerAssociationsInput{}, &output)

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

	if item.UniqueAttributeValue() != "default|cp-1" {
		t.Fatalf("expected default|cp-1, got %v", item.UniqueAttributeValue())
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-global-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-connect-peer",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cp-1",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-link",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default|link-1",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-device",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default|dvc-1",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}
