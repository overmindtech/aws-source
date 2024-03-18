package networkmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestLinkAssociationOutputMapper(t *testing.T) {
	output := networkmanager.GetLinkAssociationsOutput{
		LinkAssociations: []types.LinkAssociation{
			{
				LinkId:          sources.PtrString("link-1"),
				GlobalNetworkId: sources.PtrString("default"),
				DeviceId:        sources.PtrString("dvc-1"),
			},
		},
	}
	scope := "123456789012.eu-west-2"
	items, err := linkAssociationOutputMapper(context.Background(), &networkmanager.Client{}, scope, &networkmanager.GetLinkAssociationsInput{}, &output)

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

	if item.UniqueAttributeValue() != "default|link-1|dvc-1" {
		t.Fatalf("expected default|link-1|dvc-1, got %v", item.UniqueAttributeValue())
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-global-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "default",
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
