package securitygroup

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestSecurityGroupsMapping(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		sgName := "sgName"
		sg := types.SecurityGroup{
			GroupName: &sgName,
		}

		item, err := mapSecurityGroupToItem(&sg, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Error("item is nil")
		}
		if item.Attributes == nil || item.Attributes.AttrStruct.Fields["groupName"].GetStringValue() != sgName {
			t.Errorf("unexpected item: %v", item)
		}
	})
	t.Run("with VPC", func(t *testing.T) {
		sgName := "sgName"
		vpcId := "vpcId"
		sg := types.SecurityGroup{
			GroupName: &sgName,
			VpcId:     &vpcId,
		}

		item, err := mapSecurityGroupToItem(&sg, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if len(item.LinkedItemRequests) != 1 {
			t.Fatalf("unexpected LinkedItemRequests: %v", item)
		}
		sources.CheckItem(t, item.LinkedItemRequests[0], "vpc", "ec2-vpc", vpcId, "foo.bar")
	})
}

func TestGet(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := SecurityGroupSource{}

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

	DescribeSecurityGroupsMock func(ctx context.Context, m fakeClient, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
}

// DescribeSecurityGroups implements SecurityGroupsClient
func (m fakeClient) DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return m.DescribeSecurityGroupsMock(ctx, m, params, optFns...)
}

func createFakeClient(t *testing.T) SecurityGroupClient {
	return fakeClient{
		DescribeSecurityGroupsMock: func(ctx context.Context, m fakeClient, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
			m.sgClientCalls += 1
			if m.sgClientCalls > 2 {
				t.Error("Called DescribeSecurityGroupsMock too often (>2)")
				return nil, nil
			}
			if params.NextToken == nil {
				firstName := "first"
				nextToken := "page2"
				return &ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{
							GroupName: &firstName,
						},
					},
					NextToken: &nextToken,
				}, nil
			} else if *params.NextToken == "page2" {
				secondName := "second"
				return &ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{
							GroupName: &secondName,
						},
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
		if item.Attributes.AttrStruct.Fields["groupName"].GetStringValue() != "first" {
			t.Errorf("unexpected first item: %v", item)
		}
	})
}

func TestList(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := SecurityGroupSource{}

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
		items, err := findImpl(context.Background(), createFakeClient(t), "foo.bar")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("unexpected items (len=%v): %v", len(items), items)
		}
		if items[0].Attributes.AttrStruct.Fields["groupName"].GetStringValue() != "first" {
			t.Errorf("unexpected first item: %v", items[0])
		}
		if items[1].Attributes.AttrStruct.Fields["groupName"].GetStringValue() != "second" {
			t.Errorf("unexpected second item: %v", items[0])
		}
	})
}
