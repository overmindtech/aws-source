package networkmanager

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestNetworkResourceOutputMapper(t *testing.T) {
	scope := "123456789012.eu-west-2"
	tests := []struct {
		name   string
		input  networkmanager.GetNetworkResourcesInput
		output networkmanager.GetNetworkResourcesOutput
		tests  []sources.QueryTests
	}{
		{
			name: "ok, one entity",
			input: networkmanager.GetNetworkResourcesInput{
				GlobalNetworkId: sources.PtrString("default"),
			},
			output: networkmanager.GetNetworkResourcesOutput{
				NetworkResources: []types.NetworkResource{
					{
						ResourceId:   sources.PtrString("conn-1"),
						ResourceArn:  sources.PtrString("arn:aws:networkmanager:us-west-2:123456789012:connection/conn-1"),
						ResourceType: sources.PtrString(resourceTypeConnection),
					},
					{
						ResourceId:   sources.PtrString("d-1"),
						ResourceArn:  sources.PtrString("arn:aws:networkmanager:us-west-2:123456789012:device/d-1"),
						ResourceType: sources.PtrString(resourceTypeDevice),
					},
					{
						ResourceId:   sources.PtrString("link-1"),
						ResourceArn:  sources.PtrString("arn:aws:networkmanager:us-west-2:123456789012:link/link-1"),
						ResourceType: sources.PtrString(resourceTypeLink),
					},
					{
						ResourceId:   sources.PtrString("site-1"),
						ResourceArn:  sources.PtrString("arn:aws:networkmanager:us-west-2:123456789012:site/site-1"),
						ResourceType: sources.PtrString(resourceTypeSite),
					},
					{
						ResourceId:   sources.PtrString("dxcon-1"),
						ResourceArn:  sources.PtrString("arn:aws:directconnect:us-west-2:123456789012:connection/dxcon-1"),
						ResourceType: sources.PtrString(resourceTypeDxCon),
					},
					{
						ResourceId:   sources.PtrString("gw-1"),
						ResourceArn:  sources.PtrString("arn:aws:directconnect:us-west-2:123456789012:direct-connect-gateway/gw-1"),
						ResourceType: sources.PtrString(resourceTypeDxGateway),
					},
					{
						ResourceId:   sources.PtrString("vif-1"),
						ResourceArn:  sources.PtrString("arn:aws:directconnect:us-west-2:123456789012:virtual-interface/vif-1"),
						ResourceType: sources.PtrString(resourceTypeDxVif),
					},
					{
						ResourceId:   sources.PtrString("cgtw-1"),
						ResourceArn:  sources.PtrString("arn:aws:ec2:us-west-2:123456789012:customer-gateway/cgtw-1"),
						ResourceType: sources.PtrString(resourceTypeVPCCustomerGateway),
					},
					{
						ResourceId:   sources.PtrString("tgtw-1"),
						ResourceArn:  sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgtw-1"),
						ResourceType: sources.PtrString(resourceTypeVPCTransitGateway),
					},
					{
						ResourceId:   sources.PtrString("att-1"),
						ResourceArn:  sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway-attachment/att-1"),
						ResourceType: sources.PtrString(resourceTypeVPCTransitGatewayAttachment),
					},
					{
						ResourceId:   sources.PtrString("peer-1"),
						ResourceArn:  sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway-connect-peer/peer-1"),
						ResourceType: sources.PtrString(resourceTypeVPCTransitGatewayPeer),
					},
					{
						ResourceId:   sources.PtrString("tgrt-1"),
						ResourceArn:  sources.PtrString("arn:aws:ec2:us-west-2:123456789012:transit-gateway-route-table/tgrt-1"),
						ResourceType: sources.PtrString(resourceTypeVPCTransitGatewayRouteTable),
					},
					{
						ResourceId:   sources.PtrString("conn-1"),
						ResourceArn:  sources.PtrString("arn:aws:ec2:us-west-2:123456789012:vpn-connection/conn-1"),
						ResourceType: sources.PtrString(resourceTypeVPCVPNConnection),
					},
				},
			},
			tests: []sources.QueryTests{
				// connection
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:networkmanager:us-west-2:123456789012:connection/conn-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-connection",
						ExpectedMethod: sdp.QueryMethod_SEARCH,
						ExpectedQuery:  "default|conn-1",
						ExpectedScope:  scope,
					},
				},
				// device
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:networkmanager:us-west-2:123456789012:device/d-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-device",
						ExpectedMethod: sdp.QueryMethod_SEARCH,
						ExpectedQuery:  "default|d-1",
						ExpectedScope:  scope,
					},
				},
				// link
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:networkmanager:us-west-2:123456789012:link/link-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-link",
						ExpectedMethod: sdp.QueryMethod_SEARCH,
						ExpectedQuery:  "default|link-1",
						ExpectedScope:  scope,
					},
				},
				// site
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:networkmanager:us-west-2:123456789012:site/site-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-site",
						ExpectedMethod: sdp.QueryMethod_SEARCH,
						ExpectedQuery:  "default|site-1",
						ExpectedScope:  scope,
					},
				},
				// directconnect-connection
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:directconnect:us-west-2:123456789012:connection/dxcon-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "directconnect-connection",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "dxcon-1",
						ExpectedScope:  scope,
					},
				},
				// directconnect-direct-connect-gateway
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:directconnect:us-west-2:123456789012:direct-connect-gateway/gw-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "directconnect-direct-connect-gateway",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "gw-1",
						ExpectedScope:  scope,
					},
				},
				// directconnect-virtual-interface
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:directconnect:us-west-2:123456789012:virtual-interface/vif-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "directconnect-virtual-interface",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "vif-1",
						ExpectedScope:  scope,
					},
				},
				// ec2-customer-gateway
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:ec2:us-west-2:123456789012:customer-gateway/cgtw-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "ec2-customer-gateway",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "cgtw-1",
						ExpectedScope:  scope,
					},
				},
				// ec2-transit-gateway
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgtw-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "ec2-transit-gateway",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "tgtw-1",
						ExpectedScope:  scope,
					},
				},
				// ec2-transit-gateway-attachment
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:ec2:us-west-2:123456789012:transit-gateway-attachment/att-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "ec2-transit-gateway-attachment",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "att-1",
						ExpectedScope:  scope,
					},
				},
				// ec2-transit-gateway-connect-peer
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:ec2:us-west-2:123456789012:transit-gateway-connect-peer/peer-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "ec2-transit-gateway-connect-peer",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "peer-1",
						ExpectedScope:  scope,
					},
				},
				// ec2-transit-gateway-route-table
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:ec2:us-west-2:123456789012:transit-gateway-route-table/tgrt-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "ec2-transit-gateway-route-table",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "tgrt-1",
						ExpectedScope:  scope,
					},
				},
				// ec2-vpn-connection
				{
					{
						ExpectedType:   "networkmanager-global-network",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "networkmanager-network-resource-relationship",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "default|arn:aws:ec2:us-west-2:123456789012:vpn-connection/conn-1",
						ExpectedScope:  scope,
					},
					{
						ExpectedType:   "ec2-vpn-connection",
						ExpectedMethod: sdp.QueryMethod_GET,
						ExpectedQuery:  "conn-1",
						ExpectedScope:  scope,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := networkResourcesOutputMapper(context.Background(), &networkmanager.Client{}, scope, &tt.input, &tt.output)
			if err != nil {
				t.Error(err)
			}
			for i, _ := range items {
				if err := items[i].Validate(); err != nil {
					t.Error(err)
				}
				if items[i].UniqueAttributeValue() != fmt.Sprintf(`%s|%s`, *tt.input.GlobalNetworkId, *tt.output.NetworkResources[i].ResourceArn) {
					t.Fatalf("expected %s, got %s", fmt.Sprintf(`%s|%s`, *tt.input.GlobalNetworkId, *tt.output.NetworkResources[i].ResourceArn), items[i].UniqueAttributeValue())
				}
				tt.tests[i].Execute(t, items[i])
			}

		})
	}
}
