package vpc

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

var testVpc1 = types.Vpc{
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
}

var testVpc2 = types.Vpc{
	CidrBlock: sources.PtrString("10.0.0.2/24"),
	CidrBlockAssociationSet: []types.VpcCidrBlockAssociation{
		{
			AssociationId: sources.PtrString("sdc"),
			CidrBlock:     sources.PtrString("10.0.0.2/24"),
			CidrBlockState: &types.VpcCidrBlockState{
				State:         types.VpcCidrBlockStateCodeAssociated,
				StatusMessage: sources.PtrString("working..."),
			},
		},
	},
	DhcpOptionsId:               sources.PtrString("something"),
	InstanceTenancy:             types.TenancyDefault,
	Ipv6CidrBlockAssociationSet: []types.VpcIpv6CidrBlockAssociation{},
	IsDefault:                   sources.PtrBool(false),
	OwnerId:                     sources.PtrString("owner"),
	State:                       types.VpcStateAvailable,
	Tags:                        []types.Tag{},
	VpcId:                       sources.PtrString("example2"),
}

func TestVPCsMapping(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		item, err := mapVpcToItem(&testVpc1, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if item.Attributes == nil || item.Attributes.AttrStruct.Fields["vpcId"].GetStringValue() != "vpc-0d7892e00e573e701" {
			t.Errorf("unexpected item: %v", item)
		}
	})
}

func TestGet(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := VpcSource{}

		items, err := src.Get(context.Background(), "foo.bar", "query")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested scope foo.bar does not match source scope .") {
			t.Errorf("expected 'requested scope foo.bar does not match source scope .', got '%v'", err.Error())
		}
	})
}

type fakeClient struct {
	sgClientCalls int

	DescribeVpcsMock func(ctx context.Context, m fakeClient, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
}

// DescribeVpcs implements VPCsClient
func (m fakeClient) DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	return m.DescribeVpcsMock(ctx, m, params, optFns...)
}

func createFakeClient(t *testing.T) VpcClient {
	return fakeClient{
		DescribeVpcsMock: func(ctx context.Context, m fakeClient, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
			m.sgClientCalls += 1
			if m.sgClientCalls > 2 {
				t.Error("Called DescribeVpcsMock too often (>2)")
				return nil, nil
			}
			if params.NextToken == nil {
				nextToken := "page2"
				return &ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{
						testVpc1,
					},
					NextToken: &nextToken,
				}, nil
			} else if *params.NextToken == "page2" {
				return &ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{
						testVpc2,
					},
				}, nil
			}
			return nil, nil
		},
	}
}

func TestGetV2Impl(t *testing.T) {
	t.Parallel()
	t.Run("with client", func(t *testing.T) {
		item, err := getImpl(context.Background(), createFakeClient(t), "foo.bar", "query")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if item == nil {
			t.Fatalf("item is nil")
		}
		if item.Attributes.AttrStruct.Fields["vpcId"].GetStringValue() != "vpc-0d7892e00e573e701" {
			t.Errorf("unexpected first item: %v", item)
		}
	})
}

func TestList(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := VpcSource{}

		items, err := src.List(context.Background(), "foo.bar")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested scope foo.bar does not match source scope .") {
			t.Errorf("expected 'requested scope foo.bar does not match source scope .', got '%v'", err.Error())
		}
	})
}

func TestListV2Impl(t *testing.T) {
	t.Parallel()
	t.Run("with client", func(t *testing.T) {
		items, err := listImpl(context.Background(), createFakeClient(t), "foo.bar")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("unexpected items (len=%v): %v", len(items), items)
		}
		if items[0].Attributes.AttrStruct.Fields["vpcId"].GetStringValue() != "vpc-0d7892e00e573e701" {
			t.Errorf("unexpected first item: %v", items[0])
		}
		if items[1].Attributes.AttrStruct.Fields["vpcId"].GetStringValue() != "example2" {
			t.Errorf("unexpected second item: %v", items[0])
		}
	})
}
