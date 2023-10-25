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

func TestVpcPeeringConnectionOutputMapper(t *testing.T) {
	output := &ec2.DescribeVpcPeeringConnectionsOutput{
		VpcPeeringConnections: []types.VpcPeeringConnection{
			{
				VpcPeeringConnectionId: sources.PtrString("pcx-1234567890"),
				Status: &types.VpcPeeringConnectionStateReason{
					Code:    types.VpcPeeringConnectionStateReasonCodeActive, // health
					Message: sources.PtrString("message"),
				},
				AccepterVpcInfo: &types.VpcPeeringConnectionVpcInfo{
					CidrBlock: sources.PtrString("10.0.0.1/24"),
					CidrBlockSet: []types.CidrBlock{
						{
							CidrBlock: sources.PtrString("10.0.2.1/24"),
						},
					},
					Ipv6CidrBlockSet: []types.Ipv6CidrBlock{
						{
							Ipv6CidrBlock: sources.PtrString("::/64"),
						},
					},
					OwnerId: sources.PtrString("123456789012"),
					Region:  sources.PtrString("eu-west-2"),      // link
					VpcId:   sources.PtrString("vpc-1234567890"), // link
					PeeringOptions: &types.VpcPeeringConnectionOptionsDescription{
						AllowDnsResolutionFromRemoteVpc: sources.PtrBool(true),
					},
				},
				RequesterVpcInfo: &types.VpcPeeringConnectionVpcInfo{
					CidrBlock: sources.PtrString("10.0.0.1/24"),
					CidrBlockSet: []types.CidrBlock{
						{
							CidrBlock: sources.PtrString("10.0.2.1/24"),
						},
					},
					Ipv6CidrBlockSet: []types.Ipv6CidrBlock{
						{
							Ipv6CidrBlock: sources.PtrString("::/64"),
						},
					},
					OwnerId: sources.PtrString("987654321098"),
					PeeringOptions: &types.VpcPeeringConnectionOptionsDescription{
						AllowDnsResolutionFromRemoteVpc: sources.PtrBool(true),
					},
					Region: sources.PtrString("eu-west-5"),      // link
					VpcId:  sources.PtrString("vpc-9887654321"), // link
				},
			},
		},
	}

	items, err := vpcPeeringConnectionOutputMapper(context.Background(), nil, "foo", nil, output)

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
	tests := sources.QueryTests{
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-1234567890",
			ExpectedScope:  "123456789012.eu-west-2",
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-9887654321",
			ExpectedScope:  "987654321098.eu-west-5",
		},
	}

	tests.Execute(t, item)

}

func TestNewVpcPeeringConnectionSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	source := NewVpcPeeringConnectionSource(config, account, &TestRateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
