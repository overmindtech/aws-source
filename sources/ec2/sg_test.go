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

func TestSecurityGroupInputMapperGet(t *testing.T) {
	input, err := SecurityGroupInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.GroupIds) != 1 {
		t.Fatalf("expected 1 SecurityGroup ID, got %v", len(input.GroupIds))
	}

	if input.GroupIds[0] != "bar" {
		t.Errorf("expected SecurityGroup ID to be bar, got %v", input.GroupIds[0])
	}
}

func TestSecurityGroupInputMapperList(t *testing.T) {
	input, err := SecurityGroupInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.GroupIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestSecurityGroupOutputMapper(t *testing.T) {
	output := &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: []types.SecurityGroup{
			{
				Description: sources.PtrString("default VPC security group"),
				GroupName:   sources.PtrString("default"),
				IpPermissions: []types.IpPermission{
					{
						IpProtocol:    sources.PtrString("-1"),
						IpRanges:      []types.IpRange{},
						Ipv6Ranges:    []types.Ipv6Range{},
						PrefixListIds: []types.PrefixListId{},
						UserIdGroupPairs: []types.UserIdGroupPair{
							{
								GroupId: sources.PtrString("sg-094e151c9fc5da181"),
								UserId:  sources.PtrString("052392120704"),
							},
						},
					},
				},
				OwnerId: sources.PtrString("052392120703"),
				GroupId: sources.PtrString("sg-094e151c9fc5da181"),
				IpPermissionsEgress: []types.IpPermission{
					{
						IpProtocol: sources.PtrString("-1"),
						IpRanges: []types.IpRange{
							{
								CidrIp: sources.PtrString("0.0.0.0/0"),
							},
						},
						Ipv6Ranges:       []types.Ipv6Range{},
						PrefixListIds:    []types.PrefixListId{},
						UserIdGroupPairs: []types.UserIdGroupPair{},
					},
				},
				VpcId: sources.PtrString("vpc-0d7892e00e573e701"),
			},
		},
	}

	items, err := SecurityGroupOutputMapper("052392120703.eu-west-2", nil, output)

	if err != nil {
		t.Fatal(err)
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
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "sg-094e151c9fc5da181",
			ExpectedScope:  "052392120704.eu-west-2",
		},
	}

	tests.Execute(t, item)

}

func TestNewSecurityGroupSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewSecurityGroupSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
