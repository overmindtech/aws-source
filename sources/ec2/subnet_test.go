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

func TestSubnetInputMapperGet(t *testing.T) {
	input, err := SubnetInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.SubnetIds) != 1 {
		t.Fatalf("expected 1 Subnet ID, got %v", len(input.SubnetIds))
	}

	if input.SubnetIds[0] != "bar" {
		t.Errorf("expected Subnet ID to be bar, got %v", input.SubnetIds[0])
	}
}

func TestSubnetInputMapperList(t *testing.T) {
	input, err := SubnetInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.SubnetIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestSubnetOutputMapper(t *testing.T) {
	output := &ec2.DescribeSubnetsOutput{
		Subnets: []types.Subnet{
			{
				AvailabilityZone:            sources.PtrString("eu-west-2c"),
				AvailabilityZoneId:          sources.PtrString("euw2-az1"),
				AvailableIpAddressCount:     sources.PtrInt32(4091),
				CidrBlock:                   sources.PtrString("172.31.80.0/20"),
				DefaultForAz:                sources.PtrBool(false),
				MapPublicIpOnLaunch:         sources.PtrBool(false),
				MapCustomerOwnedIpOnLaunch:  sources.PtrBool(false),
				State:                       types.SubnetStateAvailable,
				SubnetId:                    sources.PtrString("subnet-0450a637af9984235"),
				VpcId:                       sources.PtrString("vpc-0d7892e00e573e701"),
				OwnerId:                     sources.PtrString("052392120703"),
				AssignIpv6AddressOnCreation: sources.PtrBool(false),
				Ipv6CidrBlockAssociationSet: []types.SubnetIpv6CidrBlockAssociation{
					{
						AssociationId: sources.PtrString("id-1234"),
						Ipv6CidrBlock: sources.PtrString("something"),
						Ipv6CidrBlockState: &types.SubnetCidrBlockState{
							State:         types.SubnetCidrBlockStateCodeAssociated,
							StatusMessage: sources.PtrString("something here"),
						},
					},
				},
				Tags:        []types.Tag{},
				SubnetArn:   sources.PtrString("arn:aws:ec2:eu-west-2:052392120703:subnet/subnet-0450a637af9984235"),
				EnableDns64: sources.PtrBool(false),
				Ipv6Native:  sources.PtrBool(false),
				PrivateDnsNameOptionsOnLaunch: &types.PrivateDnsNameOptionsOnLaunch{
					HostnameType:                    types.HostnameTypeIpName,
					EnableResourceNameDnsARecord:    sources.PtrBool(false),
					EnableResourceNameDnsAAAARecord: sources.PtrBool(false),
				},
			},
		},
	}

	items, err := SubnetOutputMapper("foo", nil, output)

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
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-availability-zone",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "eu-west-2c",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewSubnetSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewSubnetSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
