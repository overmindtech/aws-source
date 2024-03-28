package networkmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestCoreNetworkItemMapper(t *testing.T) {

	scope := "123456789012.eu-west-2"
	item, err := coreNetworkItemMapper(scope, &types.CoreNetwork{
		GlobalNetworkId: sources.PtrString("default"),
		CoreNetworkId:   sources.PtrString("cn-1"),
	})
	if err != nil {
		t.Error(err)
	}

	// Ensure unique attribute
	err = item.Validate()
	if err != nil {
		t.Error(err)
	}

	if item.UniqueAttributeValue() != "cn-1" {
		t.Fatalf("expected cn-1, got %v", item.UniqueAttributeValue())
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-global-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-core-network-policy",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cn-1",
			ExpectedScope:  scope,
		},
		{
			ExpectedType:   "networkmanager-connect-peer",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "cn-1",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}
