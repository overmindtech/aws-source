package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestNetworkInterfaceInputMapperGet(t *testing.T) {
	input, err := networkInterfaceInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.NetworkInterfaceIds) != 1 {
		t.Fatalf("expected 1 NetworkInterface ID, got %v", len(input.NetworkInterfaceIds))
	}

	if input.NetworkInterfaceIds[0] != "bar" {
		t.Errorf("expected NetworkInterface ID to be bar, got %v", input.NetworkInterfaceIds[0])
	}
}

func TestNetworkInterfaceInputMapperList(t *testing.T) {
	input, err := networkInterfaceInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.NetworkInterfaceIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestNetworkInterfaceOutputMapper(t *testing.T) {
	output := &ec2.DescribeNetworkInterfacesOutput{
		NetworkInterfaces: []types.NetworkInterface{
			{
				Association: &types.NetworkInterfaceAssociation{
					AllocationId:  adapters.PtrString("eipalloc-000a9739291350592"),
					AssociationId: adapters.PtrString("eipassoc-049cda1f947e5efe6"),
					IpOwnerId:     adapters.PtrString("052392120703"),
					PublicDnsName: adapters.PtrString("ec2-18-170-133-9.eu-west-2.compute.amazonaws.com"),
					PublicIp:      adapters.PtrString("18.170.133.9"),
				},
				Attachment: &types.NetworkInterfaceAttachment{
					AttachmentId:        adapters.PtrString("ela-attach-03e560efca8c9e5d8"),
					DeleteOnTermination: adapters.PtrBool(false),
					DeviceIndex:         adapters.PtrInt32(1),
					InstanceOwnerId:     adapters.PtrString("amazon-aws"),
					Status:              types.AttachmentStatusAttached,
					InstanceId:          adapters.PtrString("foo"),
				},
				AvailabilityZone: adapters.PtrString("eu-west-2b"),
				Description:      adapters.PtrString("Interface for NAT Gateway nat-0e07f7530ef076766"),
				Groups: []types.GroupIdentifier{
					{
						GroupId:   adapters.PtrString("group-123"),
						GroupName: adapters.PtrString("something"),
					},
				},
				InterfaceType: types.NetworkInterfaceTypeNatGateway,
				Ipv6Addresses: []types.NetworkInterfaceIpv6Address{
					{
						Ipv6Address: adapters.PtrString("2001:db8:1234:0000:0000:0000:0000:0000"),
					},
				},
				MacAddress:         adapters.PtrString("0a:f4:55:b0:6c:be"),
				NetworkInterfaceId: adapters.PtrString("eni-0b4652e6f2aa36d78"),
				OwnerId:            adapters.PtrString("052392120703"),
				PrivateDnsName:     adapters.PtrString("ip-172-31-35-98.eu-west-2.compute.internal"),
				PrivateIpAddress:   adapters.PtrString("172.31.35.98"),
				PrivateIpAddresses: []types.NetworkInterfacePrivateIpAddress{
					{
						Association: &types.NetworkInterfaceAssociation{
							AllocationId:    adapters.PtrString("eipalloc-000a9739291350592"),
							AssociationId:   adapters.PtrString("eipassoc-049cda1f947e5efe6"),
							IpOwnerId:       adapters.PtrString("052392120703"),
							PublicDnsName:   adapters.PtrString("ec2-18-170-133-9.eu-west-2.compute.amazonaws.com"),
							PublicIp:        adapters.PtrString("18.170.133.9"),
							CarrierIp:       adapters.PtrString("18.170.133.10"),
							CustomerOwnedIp: adapters.PtrString("18.170.133.11"),
						},
						Primary:          adapters.PtrBool(true),
						PrivateDnsName:   adapters.PtrString("ip-172-31-35-98.eu-west-2.compute.internal"),
						PrivateIpAddress: adapters.PtrString("172.31.35.98"),
					},
				},
				RequesterId:      adapters.PtrString("440527171281"),
				RequesterManaged: adapters.PtrBool(true),
				SourceDestCheck:  adapters.PtrBool(false),
				Status:           types.NetworkInterfaceStatusInUse,
				SubnetId:         adapters.PtrString("subnet-0d8ae4b4e07647efa"),
				TagSet:           []types.Tag{},
				VpcId:            adapters.PtrString("vpc-0d7892e00e573e701"),
			},
		},
	}

	items, err := networkInterfaceOutputMapper(context.Background(), nil, "foo", nil, output)

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
	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "foo",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "group-123",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "2001:db8:1234:0000:0000:0000:0000:0000",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "ip-172-31-35-98.eu-west-2.compute.internal",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "172.31.35.98",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "18.170.133.9",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "18.170.133.10",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "ec2-18-170-133-9.eu-west-2.compute.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "18.170.133.11",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0d8ae4b4e07647efa",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewNetworkInterfaceAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewNetworkInterfaceAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
