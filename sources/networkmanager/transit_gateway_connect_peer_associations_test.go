package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestTransitGatewayConnectPeerAssociationsOutputMapper(t *testing.T) {
	scope := "123456789012.eu-west-2"
	tests := []struct {
		name           string
		out            networkmanager.GetTransitGatewayConnectPeerAssociationsOutput
		expectedHealth sdp.Health
		expectedAttr   string
		tests          sources.QueryTests
	}{
		{
			name: "ok",
			out: networkmanager.GetTransitGatewayConnectPeerAssociationsOutput{
				TransitGatewayConnectPeerAssociations: []types.TransitGatewayConnectPeerAssociation{
					{
						GlobalNetworkId:              sources.PtrString("default"),
						TransitGatewayConnectPeerArn: sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway-connect-peer-association/tgw-1234"),
						State:                        types.TransitGatewayConnectPeerAssociationStateAvailable,
						DeviceId:                     sources.PtrString("device-1"),
						LinkId:                       sources.PtrString("link-1"),
					},
				},
			},
			expectedHealth: sdp.Health_HEALTH_OK,
			expectedAttr:   "default|arn:aws:ec2:us-west-2:123456789012:transit-gateway-connect-peer-association/tgw-1234",
			tests: sources.QueryTests{
				{
					ExpectedType:   "networkmanager-global-network",
					ExpectedMethod: sdp.QueryMethod_GET,
					ExpectedQuery:  "default",
					ExpectedScope:  scope,
				},
				{
					ExpectedType:   "networkmanager-device",
					ExpectedMethod: sdp.QueryMethod_SEARCH,
					ExpectedQuery:  "default|device-1",
					ExpectedScope:  scope,
				},
				{
					ExpectedType:   "networkmanager-link",
					ExpectedMethod: sdp.QueryMethod_SEARCH,
					ExpectedQuery:  "default|link-1",
					ExpectedScope:  scope,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := transitGatewayConnectPeerAssociationsOutputMapper(context.Background(), &networkmanager.Client{}, scope, &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{}, &tt.out)
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

			if item.UniqueAttributeValue() != tt.expectedAttr {
				t.Fatalf("want %s, got %s", tt.expectedAttr, item.UniqueAttributeValue())
			}

			if tt.expectedHealth != *item.Health {
				t.Fatalf("want %d, got %d", tt.expectedHealth, *item.Health)
			}

			tt.tests.Execute(t, item)
		})
	}

}
