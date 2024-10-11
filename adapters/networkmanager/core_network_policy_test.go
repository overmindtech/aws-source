package networkmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestCoreNetworkPolicyItemMapper(t *testing.T) {

	scope := "123456789012.eu-west-2"
	item, err := coreNetworkPolicyItemMapper("", scope, &types.CoreNetworkPolicy{
		CoreNetworkId:   adapters.PtrString("cn-1"),
		PolicyVersionId: adapters.PtrInt32(1),
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

	tests := adapters.QueryTests{
		{
			ExpectedType:   "networkmanager-core-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cn-1",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}
