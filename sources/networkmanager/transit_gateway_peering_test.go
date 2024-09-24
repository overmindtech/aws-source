package networkmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestTransitGatewayPeeringOutputMapper(t *testing.T) {
	scope := "123456789012.eu-west-2"
	tests := []struct {
		name           string
		item           *types.TransitGatewayPeering
		expectedHealth sdp.Health
		expectedAttr   string
		tests          sources.QueryTests
	}{
		{
			name: "ok",
			item: &types.TransitGatewayPeering{
				Peering: &types.Peering{
					PeeringId:     sources.PtrString("tgp-1"),
					CoreNetworkId: sources.PtrString("cn-1"),
					State:         types.PeeringStateAvailable,
				},
				TransitGatewayArn:                 sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgw-1234"),
				TransitGatewayPeeringAttachmentId: sources.PtrString("gpa-1"),
			},
			expectedHealth: sdp.Health_HEALTH_OK,
			expectedAttr:   "tgp-1",
			tests: sources.QueryTests{
				{
					ExpectedType:   "networkmanager-core-network",
					ExpectedMethod: sdp.QueryMethod_GET,
					ExpectedQuery:  "cn-1",
					ExpectedScope:  scope,
				},
				{
					ExpectedType:   "ec2-transit-gateway",
					ExpectedMethod: sdp.QueryMethod_SEARCH,
					ExpectedQuery:  "arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgw-1234",
					ExpectedScope:  "123456789012.us-west-2",
				},
				{
					ExpectedType:   "ec2-transit-gateway-peering-attachment",
					ExpectedMethod: sdp.QueryMethod_GET,
					ExpectedQuery:  "gpa-1",
					ExpectedScope:  scope,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := transitGatewayPeeringItemMapper("", scope, tt.item)
			if err != nil {
				t.Error(err)
			}

			if item.UniqueAttributeValue() != tt.expectedAttr {
				t.Fatalf("want %s, got %s", tt.expectedAttr, item.UniqueAttributeValue())
			}

			if tt.expectedHealth != item.GetHealth() {
				t.Fatalf("want %d, got %d", tt.expectedHealth, item.GetHealth())
			}

			tt.tests.Execute(t, item)
		})
	}

}
