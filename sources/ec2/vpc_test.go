package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestVpcInputMapperGet(t *testing.T) {
	input, err := VpcInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.VpcIds) != 1 {
		t.Fatalf("expected 1 Vpc ID, got %v", len(input.VpcIds))
	}

	if input.VpcIds[0] != "bar" {
		t.Errorf("expected Vpc ID to be bar, got %v", input.VpcIds[0])
	}
}

func TestVpcInputMapperList(t *testing.T) {
	input, err := VpcInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.VpcIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestVpcOutputMapper(t *testing.T) {
	output := &ec2.DescribeVpcsOutput{
		Vpcs: []types.Vpc{
			{
				CidrBlock:       sources.PtrString("172.31.0.0/16"),
				DhcpOptionsId:   sources.PtrString("dopt-0959b838bf4a4c7b8"),
				State:           types.VpcStateAvailable,
				VpcId:           sources.PtrString("vpc-0d7892e00e573e701"),
				OwnerId:         sources.PtrString("052392120703"),
				InstanceTenancy: types.TenancyDefault,
				CidrBlockAssociationSet: []types.VpcCidrBlockAssociation{
					{
						AssociationId: sources.PtrString("vpc-cidr-assoc-0b77866f37f500af6"),
						CidrBlock:     sources.PtrString("172.31.0.0/16"),
						CidrBlockState: &types.VpcCidrBlockState{
							State: types.VpcCidrBlockStateCodeAssociated,
						},
					},
				},
				IsDefault: sources.PtrBool(false),
				Tags: []types.Tag{
					{
						Key:   sources.PtrString("aws:cloudformation:logical-id"),
						Value: sources.PtrString("VPC"),
					},
					{
						Key:   sources.PtrString("aws:cloudformation:stack-id"),
						Value: sources.PtrString("arn:aws:cloudformation:eu-west-2:052392120703:stack/StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-8c2a9348-a30c-4ac3-94c2-8279157c9243/ccde3240-7afa-11ed-81ff-02845d4c2702"),
					},
					{
						Key:   sources.PtrString("aws:cloudformation:stack-name"),
						Value: sources.PtrString("StackSet-AWSControlTowerBP-VPC-ACCOUNT-FACTORY-V1-8c2a9348-a30c-4ac3-94c2-8279157c9243"),
					},
					{
						Key:   sources.PtrString("Name"),
						Value: sources.PtrString("aws-controltower-VPC"),
					},
				},
			},
		},
	}

	items, err := VpcOutputMapper("foo", output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}
}
