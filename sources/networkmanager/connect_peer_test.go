package networkmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestConnectPeerItemMapper(t *testing.T) {

	scope := "123456789012.eu-west-2"
	item, err := connectPeerItemMapper(scope, &types.ConnectPeer{
		CoreNetworkId: sources.PtrString("cn-1"),
		ConnectPeerId: sources.PtrString("cp-1"),
	})
	if err != nil {
		t.Error(err)
	}

	// Ensure unique attribute
	err = item.Validate()
	if err != nil {
		t.Error(err)
	}

	if item.UniqueAttributeValue() != "cp-1" {
		t.Fatalf("expected cp-1, got %v", item.UniqueAttributeValue())
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-core-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cn-1",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}
