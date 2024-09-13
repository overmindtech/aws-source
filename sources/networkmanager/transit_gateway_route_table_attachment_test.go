package networkmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestTransitGatewayRouteTableAttachmentItemMapper(t *testing.T) {
	scope := "123456789012.eu-west-2"
	tests := []struct {
		name         string
		input        types.TransitGatewayRouteTableAttachment
		expectedAttr string
		tests        sources.QueryTests
	}{
		{
			name: "ok",
			input: types.TransitGatewayRouteTableAttachment{
				Attachment: &types.Attachment{
					AttachmentId:  sources.PtrString("attachment1"),
					CoreNetworkId: sources.PtrString("corenetwork1"),
				},
				TransitGatewayRouteTableArn: sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway-route-table/tgw-rtb-9876543210123456"),
				PeeringId:                   sources.PtrString("peer1"),
			},
			expectedAttr: "attachment1",
			tests: sources.QueryTests{
				{
					ExpectedType:   "networkmanager-core-network",
					ExpectedMethod: sdp.QueryMethod_GET,
					ExpectedQuery:  "corenetwork1",
					ExpectedScope:  scope,
				},
				{
					ExpectedType:   "networkmanager-transit-gateway-peering",
					ExpectedMethod: sdp.QueryMethod_GET,
					ExpectedQuery:  "peer1",
					ExpectedScope:  scope,
				},
				{
					ExpectedType:   "ec2-transit-gateway-route-table",
					ExpectedMethod: sdp.QueryMethod_SEARCH,
					ExpectedQuery:  "arn:aws:ec2:us-west-2:123456789012:transit-gateway-route-table/tgw-rtb-9876543210123456",
					ExpectedScope:  "123456789012.us-west-2",
				},
			},
		},
		{
			name: "missing ec2-transit-gateway-route-table",
			input: types.TransitGatewayRouteTableAttachment{
				Attachment: &types.Attachment{
					AttachmentId:  sources.PtrString("attachment1"),
					CoreNetworkId: sources.PtrString("corenetwork1"),
				},
			},
			expectedAttr: "attachment1",
			tests: sources.QueryTests{
				{
					ExpectedType:   "networkmanager-core-network",
					ExpectedMethod: sdp.QueryMethod_GET,
					ExpectedQuery:  "corenetwork1",
					ExpectedScope:  scope,
				},
			},
		},
		{
			name: "invalid ec2-transit-gateway-route-table",
			input: types.TransitGatewayRouteTableAttachment{
				Attachment: &types.Attachment{
					AttachmentId:  sources.PtrString("attachment1"),
					CoreNetworkId: sources.PtrString("corenetwork1"),
				},
				TransitGatewayRouteTableArn: sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway-route-table-tgw-rtb-9876543210123456"),
			},
			expectedAttr: "attachment1",
			tests: sources.QueryTests{
				{
					ExpectedType:   "networkmanager-core-network",
					ExpectedMethod: sdp.QueryMethod_GET,
					ExpectedQuery:  "corenetwork1",
					ExpectedScope:  scope,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := transitGatewayRouteTableAttachmentItemMapper("", scope, &tt.input)
			if err != nil {
				t.Error(err)
			}
			if err := item.Validate(); err != nil {
				t.Error(err)
			}
			// Ensure unique attribute
			if item.UniqueAttributeValue() != tt.expectedAttr {
				t.Fatalf("expected %s, got %s", tt.expectedAttr, item.UniqueAttributeValue())
			}
			tt.tests.Execute(t, item)
		})
	}

}
