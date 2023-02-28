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

func TestSecurityGroupRuleInputMapperGet(t *testing.T) {
	input, err := SecurityGroupRuleInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.SecurityGroupRuleIds) != 1 {
		t.Fatalf("expected 1 SecurityGroupRule ID, got %v", len(input.SecurityGroupRuleIds))
	}

	if input.SecurityGroupRuleIds[0] != "bar" {
		t.Errorf("expected SecurityGroupRule ID to be bar, got %v", input.SecurityGroupRuleIds[0])
	}
}

func TestSecurityGroupRuleInputMapperList(t *testing.T) {
	input, err := SecurityGroupRuleInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.SecurityGroupRuleIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestSecurityGroupRuleOutputMapper(t *testing.T) {
	output := &ec2.DescribeSecurityGroupRulesOutput{
		SecurityGroupRules: []types.SecurityGroupRule{
			{
				SecurityGroupRuleId: sources.PtrString("sgr-0b0e42d1431e832bd"),
				GroupId:             sources.PtrString("sg-0814766e46f201c22"),
				GroupOwnerId:        sources.PtrString("052392120703"),
				IsEgress:            sources.PtrBool(false),
				IpProtocol:          sources.PtrString("tcp"),
				FromPort:            sources.PtrInt32(2049),
				ToPort:              sources.PtrInt32(2049),
				ReferencedGroupInfo: &types.ReferencedSecurityGroup{
					GroupId: sources.PtrString("sg-09371b4a54fe7ab38"),
					UserId:  sources.PtrString("052392120703"),
				},
				Description: sources.PtrString("Created by the LIW for EFS at 2022-12-16T19:14:27.033Z"),
				Tags:        []types.Tag{},
			},
			{
				SecurityGroupRuleId: sources.PtrString("sgr-04b583a90b4fa4ada"),
				GroupId:             sources.PtrString("sg-09371b4a54fe7ab38"),
				GroupOwnerId:        sources.PtrString("052392120703"),
				IsEgress:            sources.PtrBool(true),
				IpProtocol:          sources.PtrString("tcp"),
				FromPort:            sources.PtrInt32(2049),
				ToPort:              sources.PtrInt32(2049),
				ReferencedGroupInfo: &types.ReferencedSecurityGroup{
					GroupId: sources.PtrString("sg-0814766e46f201c22"),
					UserId:  sources.PtrString("052392120703"),
				},
				Description: sources.PtrString("Created by the LIW for EFS at 2022-12-16T19:14:27.349Z"),
				Tags:        []types.Tag{},
			},
		},
	}

	items, err := SecurityGroupRuleOutputMapper("foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "sg-0814766e46f201c22",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "sg-09371b4a54fe7ab38",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewSecurityGroupRuleSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewSecurityGroupRuleSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
