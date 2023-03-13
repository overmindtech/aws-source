package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestNetworkInterfaceInputMapperGet(t *testing.T) {
	input, err := NetworkInterfaceInputMapperGet("foo", "bar")

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
	input, err := NetworkInterfaceInputMapperList("foo")

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
					AllocationId:  sources.PtrString("eipalloc-000a9739291350592"),
					AssociationId: sources.PtrString("eipassoc-049cda1f947e5efe6"),
					IpOwnerId:     sources.PtrString("052392120703"),
					PublicDnsName: sources.PtrString("ec2-18-170-133-9.eu-west-2.compute.amazonaws.com"),
					PublicIp:      sources.PtrString("18.170.133.9"),
				},
				Attachment: &types.NetworkInterfaceAttachment{
					AttachmentId:        sources.PtrString("ela-attach-03e560efca8c9e5d8"),
					DeleteOnTermination: sources.PtrBool(false),
					DeviceIndex:         sources.PtrInt32(1),
					InstanceOwnerId:     sources.PtrString("amazon-aws"),
					Status:              types.AttachmentStatusAttached,
					InstanceId:          sources.PtrString("foo"),
				},
				AvailabilityZone: sources.PtrString("eu-west-2b"),
				Description:      sources.PtrString("Interface for NAT Gateway nat-0e07f7530ef076766"),
				Groups: []types.GroupIdentifier{
					{
						GroupId:   sources.PtrString("group-123"),
						GroupName: sources.PtrString("something"),
					},
				},
				InterfaceType: types.NetworkInterfaceTypeNatGateway,
				Ipv6Addresses: []types.NetworkInterfaceIpv6Address{
					{
						Ipv6Address: sources.PtrString("2001:db8:1234:0000:0000:0000:0000:0000"),
					},
				},
				MacAddress:         sources.PtrString("0a:f4:55:b0:6c:be"),
				NetworkInterfaceId: sources.PtrString("eni-0b4652e6f2aa36d78"),
				OwnerId:            sources.PtrString("052392120703"),
				PrivateDnsName:     sources.PtrString("ip-172-31-35-98.eu-west-2.compute.internal"),
				PrivateIpAddress:   sources.PtrString("172.31.35.98"),
				PrivateIpAddresses: []types.NetworkInterfacePrivateIpAddress{
					{
						Association: &types.NetworkInterfaceAssociation{
							AllocationId:    sources.PtrString("eipalloc-000a9739291350592"),
							AssociationId:   sources.PtrString("eipassoc-049cda1f947e5efe6"),
							IpOwnerId:       sources.PtrString("052392120703"),
							PublicDnsName:   sources.PtrString("ec2-18-170-133-9.eu-west-2.compute.amazonaws.com"),
							PublicIp:        sources.PtrString("18.170.133.9"),
							CarrierIp:       sources.PtrString("18.170.133.10"),
							CustomerOwnedIp: sources.PtrString("18.170.133.11"),
						},
						Primary:          sources.PtrBool(true),
						PrivateDnsName:   sources.PtrString("ip-172-31-35-98.eu-west-2.compute.internal"),
						PrivateIpAddress: sources.PtrString("172.31.35.98"),
					},
				},
				RequesterId:      sources.PtrString("440527171281"),
				RequesterManaged: sources.PtrBool(true),
				SourceDestCheck:  sources.PtrBool(false),
				Status:           types.NetworkInterfaceStatusInUse,
				SubnetId:         sources.PtrString("subnet-0d8ae4b4e07647efa"),
				TagSet:           []types.Tag{},
				VpcId:            sources.PtrString("vpc-0d7892e00e573e701"),
			},
		},
	}

	items, err := NetworkInterfaceOutputMapper("foo", nil, output)

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
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "foo",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-availability-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "eu-west-2b",
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

func TestNewNetworkInterfaceSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewNetworkInterfaceSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
