package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestRouteTableInputMapperGet(t *testing.T) {
	input, err := RouteTableInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.RouteTableIds) != 1 {
		t.Fatalf("expected 1 RouteTable ID, got %v", len(input.RouteTableIds))
	}

	if input.RouteTableIds[0] != "bar" {
		t.Errorf("expected RouteTable ID to be bar, got %v", input.RouteTableIds[0])
	}
}

func TestRouteTableInputMapperList(t *testing.T) {
	input, err := RouteTableInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.RouteTableIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestRouteTableOutputMapper(t *testing.T) {
	output := &ec2.DescribeRouteTablesOutput{
		RouteTables: []types.RouteTable{
			{
				Associations: []types.RouteTableAssociation{
					{
						Main:                    sources.PtrBool(false),
						RouteTableAssociationId: sources.PtrString("rtbassoc-0aa1442039abff3db"),
						RouteTableId:            sources.PtrString("rtb-00b1197fa95a6b35f"),
						SubnetId:                sources.PtrString("subnet-06c0dea0437180c61"),
						GatewayId:               sources.PtrString("ID"),
						AssociationState: &types.RouteTableAssociationState{
							State: types.RouteTableAssociationStateCodeAssociated,
						},
					},
				},
				PropagatingVgws: []types.PropagatingVgw{
					{
						GatewayId: sources.PtrString("goo"),
					},
				},
				RouteTableId: sources.PtrString("rtb-00b1197fa95a6b35f"),
				Routes: []types.Route{
					{
						DestinationCidrBlock: sources.PtrString("172.31.0.0/16"),
						GatewayId:            sources.PtrString("local"),
						Origin:               types.RouteOriginCreateRouteTable,
						State:                types.RouteStateActive,
					},
					{
						DestinationPrefixListId:     sources.PtrString("pl-7ca54015"),
						GatewayId:                   sources.PtrString("vpce-09fcbac4dcf142db3"),
						Origin:                      types.RouteOriginCreateRoute,
						State:                       types.RouteStateActive,
						CarrierGatewayId:            sources.PtrString("id"),
						EgressOnlyInternetGatewayId: sources.PtrString("id"),
						InstanceId:                  sources.PtrString("id"),
						InstanceOwnerId:             sources.PtrString("id"),
						LocalGatewayId:              sources.PtrString("id"),
						NatGatewayId:                sources.PtrString("id"),
						NetworkInterfaceId:          sources.PtrString("id"),
						TransitGatewayId:            sources.PtrString("id"),
						VpcPeeringConnectionId:      sources.PtrString("id"),
					},
				},
				VpcId:   sources.PtrString("vpc-0d7892e00e573e701"),
				OwnerId: sources.PtrString("052392120703"),
			},
		},
	}

	items, err := RouteTableOutputMapper("foo", output)

	if err != nil {
		t.Fatal(err)
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

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "subnet-06c0dea0437180c61",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-internet-gateway",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "ID",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-carrier-gateway",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-egress-only-internet-gateway",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-local-gateway",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-nat-gateway",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-network-interface",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-transit-gateway",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-vpc-peering-connection",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "id",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}