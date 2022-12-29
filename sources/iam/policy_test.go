package iam

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func (t *TestIAMClient) GetPolicy(ctx context.Context, params *iam.GetPolicyInput, optFns ...func(*iam.Options)) (*iam.GetPolicyOutput, error) {
	return &iam.GetPolicyOutput{
		Policy: &types.Policy{
			PolicyName:                    sources.PtrString("AWSControlTowerStackSetRolePolicy"),
			PolicyId:                      sources.PtrString("ANPA3VLV2U277MP54R2OV"),
			Arn:                           sources.PtrString("arn:aws:iam::801795385023:policy/service-role/AWSControlTowerStackSetRolePolicy"),
			Path:                          sources.PtrString("/service-role/"),
			DefaultVersionId:              sources.PtrString("v1"),
			AttachmentCount:               sources.PtrInt32(1),
			PermissionsBoundaryUsageCount: sources.PtrInt32(0),
			IsAttachable:                  true,
			CreateDate:                    sources.PtrTime(time.Now()),
			UpdateDate:                    sources.PtrTime(time.Now()),
		},
	}, nil
}

func (t *TestIAMClient) ListEntitiesForPolicy(context.Context, *iam.ListEntitiesForPolicyInput, ...func(*iam.Options)) (*iam.ListEntitiesForPolicyOutput, error) {
	return &iam.ListEntitiesForPolicyOutput{
		PolicyGroups: []types.PolicyGroup{
			{
				GroupId:   sources.PtrString("groupId"),
				GroupName: sources.PtrString("groupName"),
			},
		},
		PolicyRoles: []types.PolicyRole{
			{
				RoleId:   sources.PtrString("roleId"),
				RoleName: sources.PtrString("roleName"),
			},
		},
		PolicyUsers: []types.PolicyUser{
			{
				UserId:   sources.PtrString("userId"),
				UserName: sources.PtrString("userName"),
			},
		},
	}, nil
}

func (t *TestIAMClient) ListPolicies(context.Context, *iam.ListPoliciesInput, ...func(*iam.Options)) (*iam.ListPoliciesOutput, error) {
	return &iam.ListPoliciesOutput{
		Policies: []types.Policy{
			{
				PolicyName:                    sources.PtrString("AWSControlTowerAdminPolicy"),
				PolicyId:                      sources.PtrString("ANPA3VLV2U2745H37HTHN"),
				Arn:                           sources.PtrString("arn:aws:iam::801795385023:policy/service-role/AWSControlTowerAdminPolicy"),
				Path:                          sources.PtrString("/service-role/"),
				DefaultVersionId:              sources.PtrString("v1"),
				AttachmentCount:               sources.PtrInt32(1),
				PermissionsBoundaryUsageCount: sources.PtrInt32(0),
				IsAttachable:                  true,
				CreateDate:                    sources.PtrTime(time.Now()),
				UpdateDate:                    sources.PtrTime(time.Now()),
			},
			{
				PolicyName:                    sources.PtrString("AWSControlTowerCloudTrailRolePolicy"),
				PolicyId:                      sources.PtrString("ANPA3VLV2U27UOP7KSM6I"),
				Arn:                           sources.PtrString("arn:aws:iam::801795385023:policy/service-role/AWSControlTowerCloudTrailRolePolicy"),
				Path:                          sources.PtrString("/service-role/"),
				DefaultVersionId:              sources.PtrString("v1"),
				AttachmentCount:               sources.PtrInt32(1),
				PermissionsBoundaryUsageCount: sources.PtrInt32(0),
				IsAttachable:                  true,
				CreateDate:                    sources.PtrTime(time.Now()),
				UpdateDate:                    sources.PtrTime(time.Now()),
			},
		},
	}, nil
}

func TestPolicyGetFunc(t *testing.T) {
	policy, err := PolicyGetFunc(context.Background(), &TestIAMClient{}, "foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if policy.Policy == nil {
		t.Error("policy was nil")
	}

	if len(policy.PolicyGroups) != 1 {
		t.Errorf("expected 1 Group, got %v", len(policy.PolicyGroups))
	}

	if len(policy.PolicyRoles) != 1 {
		t.Errorf("expected 1 Role, got %v", len(policy.PolicyRoles))
	}

	if len(policy.PolicyUsers) != 1 {
		t.Errorf("expected 1 User, got %v", len(policy.PolicyUsers))
	}
}

func TestPolicyListFunc(t *testing.T) {
	policies, err := PolicyListFunc(context.Background(), &TestIAMClient{}, "foo")

	if err != nil {
		t.Error(err)
	}

	if len(policies) != 2 {
		t.Errorf("expected 2 policies, got %v", len(policies))
	}
}

func TestPolicyItemMapper(t *testing.T) {
	item, err := PolicyItemMapper("foo", &PolicyDetails{
		Policy: &types.Policy{
			PolicyName:                    sources.PtrString("AWSControlTowerAdminPolicy"),
			PolicyId:                      sources.PtrString("ANPA3VLV2U2745H37HTHN"),
			Arn:                           sources.PtrString("arn:aws:iam::801795385023:policy/service-role/AWSControlTowerAdminPolicy"),
			Path:                          sources.PtrString("/service-role/"),
			DefaultVersionId:              sources.PtrString("v1"),
			AttachmentCount:               sources.PtrInt32(1),
			PermissionsBoundaryUsageCount: sources.PtrInt32(0),
			IsAttachable:                  true,
			CreateDate:                    sources.PtrTime(time.Now()),
			UpdateDate:                    sources.PtrTime(time.Now()),
		},
		PolicyGroups: []types.PolicyGroup{
			{
				GroupId:   sources.PtrString("groupId"),
				GroupName: sources.PtrString("groupName"),
			},
		},
		PolicyRoles: []types.PolicyRole{
			{
				RoleId:   sources.PtrString("roleId"),
				RoleName: sources.PtrString("roleName"),
			},
		},
		PolicyUsers: []types.PolicyUser{
			{
				UserId:   sources.PtrString("userId"),
				UserName: sources.PtrString("userName"),
			},
		},
	})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "iam-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "groupName",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "iam-user",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "userName",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "roleName",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}
