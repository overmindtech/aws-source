package ec2

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestNatGatewayInputMapperGet(t *testing.T) {
	input, err := NatGatewayInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.NatGatewayIds) != 1 {
		t.Fatalf("expected 1 NatGateway ID, got %v", len(input.NatGatewayIds))
	}

	if input.NatGatewayIds[0] != "bar" {
		t.Errorf("expected NatGateway ID to be bar, got %v", input.NatGatewayIds[0])
	}
}

func TestNatGatewayInputMapperList(t *testing.T) {
	input, err := NatGatewayInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filter) != 0 || len(input.NatGatewayIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestNatGatewayOutputMapper(t *testing.T) {
	output := &ec2.DescribeNatGatewaysOutput{
		NatGateways: []types.NatGateway{
			{
				CreateTime:     sources.PtrTime(time.Now()),
				DeleteTime:     sources.PtrTime(time.Now()),
				FailureCode:    sources.PtrString("Gateway.NotAttached"),
				FailureMessage: sources.PtrString("Network vpc-0d7892e00e573e701 has no Internet gateway attached"),
				NatGatewayAddresses: []types.NatGatewayAddress{
					{
						AllocationId:       sources.PtrString("eipalloc-000a9739291350592"),
						NetworkInterfaceId: sources.PtrString("eni-0c59532b8e10343ae"),
						PrivateIp:          sources.PtrString("172.31.89.23"),
					},
				},
				NatGatewayId: sources.PtrString("nat-0e4e73d7ac46af25e"),
				State:        types.NatGatewayStateFailed,
				SubnetId:     sources.PtrString("subnet-0450a637af9984235"),
				VpcId:        sources.PtrString("vpc-0d7892e00e573e701"),
				Tags: []types.Tag{
					{
						Key:   sources.PtrString("Name"),
						Value: sources.PtrString("test"),
					},
				},
				ConnectivityType: types.ConnectivityTypePublic,
			},
			{
				CreateTime: sources.PtrTime(time.Now()),
				NatGatewayAddresses: []types.NatGatewayAddress{
					{
						AllocationId:       sources.PtrString("eipalloc-000a9739291350592"),
						NetworkInterfaceId: sources.PtrString("eni-0b4652e6f2aa36d78"),
						PrivateIp:          sources.PtrString("172.31.35.98"),
						PublicIp:           sources.PtrString("18.170.133.9"),
					},
				},
				NatGatewayId: sources.PtrString("nat-0e07f7530ef076766"),
				State:        types.NatGatewayStateAvailable,
				SubnetId:     sources.PtrString("subnet-0d8ae4b4e07647efa"),
				VpcId:        sources.PtrString("vpc-0d7892e00e573e701"),
				Tags: []types.Tag{
					{
						Key:   sources.PtrString("Name"),
						Value: sources.PtrString("test"),
					},
				},
				ConnectivityType: types.ConnectivityTypePublic,
			},
		},
	}

	items, err := NatGatewayOutputMapper("foo", output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %v", len(items))
	}

	item := items[1]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-network-interface",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "eni-0b4652e6f2aa36d78",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "172.31.35.98",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "18.170.133.9",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "subnet-0d8ae4b4e07647efa",
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
