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

func TestNetworkAclInputMapperGet(t *testing.T) {
	input, err := networkAclInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.NetworkAclIds) != 1 {
		t.Fatalf("expected 1 NetworkAcl ID, got %v", len(input.NetworkAclIds))
	}

	if input.NetworkAclIds[0] != "bar" {
		t.Errorf("expected NetworkAcl ID to be bar, got %v", input.NetworkAclIds[0])
	}
}

func TestNetworkAclInputMapperList(t *testing.T) {
	input, err := networkAclInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.NetworkAclIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestNetworkAclOutputMapper(t *testing.T) {
	output := &ec2.DescribeNetworkAclsOutput{
		NetworkAcls: []types.NetworkAcl{
			{
				Associations: []types.NetworkAclAssociation{
					{
						NetworkAclAssociationId: sources.PtrString("aclassoc-0f85f8b1fde0a5939"),
						NetworkAclId:            sources.PtrString("acl-0a346e8e6f5a9ad91"),
						SubnetId:                sources.PtrString("subnet-0450a637af9984235"),
					},
					{
						NetworkAclAssociationId: sources.PtrString("aclassoc-064b78003a2d309a4"),
						NetworkAclId:            sources.PtrString("acl-0a346e8e6f5a9ad91"),
						SubnetId:                sources.PtrString("subnet-06c0dea0437180c61"),
					},
					{
						NetworkAclAssociationId: sources.PtrString("aclassoc-0575080579a7381f5"),
						NetworkAclId:            sources.PtrString("acl-0a346e8e6f5a9ad91"),
						SubnetId:                sources.PtrString("subnet-0d8ae4b4e07647efa"),
					},
				},
				Entries: []types.NetworkAclEntry{
					{
						CidrBlock:  sources.PtrString("0.0.0.0/0"),
						Egress:     sources.PtrBool(true),
						Protocol:   sources.PtrString("-1"),
						RuleAction: types.RuleActionAllow,
						RuleNumber: sources.PtrInt32(100),
					},
					{
						CidrBlock:  sources.PtrString("0.0.0.0/0"),
						Egress:     sources.PtrBool(true),
						Protocol:   sources.PtrString("-1"),
						RuleAction: types.RuleActionDeny,
						RuleNumber: sources.PtrInt32(32767),
					},
				},
				IsDefault:    sources.PtrBool(true),
				NetworkAclId: sources.PtrString("acl-0a346e8e6f5a9ad91"),
				Tags:         []types.Tag{},
				VpcId:        sources.PtrString("vpc-0d7892e00e573e701"),
				OwnerId:      sources.PtrString("052392120703"),
			},
		},
	}

	items, err := networkAclOutputMapper(context.Background(), nil, "foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.QueryTests{
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-06c0dea0437180c61",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0d8ae4b4e07647efa",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0450a637af9984235",
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

func TestNewNetworkAclSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewNetworkAclSource(client, account, region)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
