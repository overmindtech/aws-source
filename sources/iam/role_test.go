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

func (t *TestIAMClient) GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return &iam.GetRoleOutput{
		Role: &types.Role{
			Path:                     sources.PtrString("/service-role/"),
			RoleName:                 sources.PtrString("AWSControlTowerConfigAggregatorRoleForOrganizations"),
			RoleId:                   sources.PtrString("AROA3VLV2U27YSTBFCGCJ"),
			Arn:                      sources.PtrString("arn:aws:iam::801795385023:role/service-role/AWSControlTowerConfigAggregatorRoleForOrganizations"),
			CreateDate:               sources.PtrTime(time.Now()),
			AssumeRolePolicyDocument: sources.PtrString("FOO"),
			MaxSessionDuration:       sources.PtrInt32(3600),
		},
	}, nil
}

func (t *TestIAMClient) ListRolePolicies(context.Context, *iam.ListRolePoliciesInput, ...func(*iam.Options)) (*iam.ListRolePoliciesOutput, error) {
	return &iam.ListRolePoliciesOutput{
		PolicyNames: []string{
			"one",
			"two",
		},
	}, nil
}

func (t *TestIAMClient) ListRoles(context.Context, *iam.ListRolesInput, ...func(*iam.Options)) (*iam.ListRolesOutput, error) {
	return &iam.ListRolesOutput{
		Roles: []types.Role{
			{
				Path:                     sources.PtrString("/service-role/"),
				RoleName:                 sources.PtrString("AWSControlTowerConfigAggregatorRoleForOrganizations"),
				RoleId:                   sources.PtrString("AROA3VLV2U27YSTBFCGCJ"),
				Arn:                      sources.PtrString("arn:aws:iam::801795385023:role/service-role/AWSControlTowerConfigAggregatorRoleForOrganizations"),
				CreateDate:               sources.PtrTime(time.Now()),
				AssumeRolePolicyDocument: sources.PtrString("FOO"),
				MaxSessionDuration:       sources.PtrInt32(3600),
			},
		},
	}, nil
}

func TestRoleGetFunc(t *testing.T) {
	role, err := RoleGetFunc(context.Background(), &TestIAMClient{}, "foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if role.Role == nil {
		t.Error("role is nil")
	}

	if len(role.Policies) != 2 {
		t.Errorf("expected 2 policies, got %v", len(role.Policies))
	}
}

func TestRoleListFunc(t *testing.T) {
	roles, err := RoleListFunc(context.Background(), &TestIAMClient{}, "foo")

	if err != nil {
		t.Error(err)
	}

	if len(roles) != 1 {
		t.Errorf("expected 1 role, got %b", len(roles))
	}
}

func TestRoleItemMapper(t *testing.T) {
	role := RoleDetails{
		Role: &types.Role{
			Path:                     sources.PtrString("/service-role/"),
			RoleName:                 sources.PtrString("AWSControlTowerConfigAggregatorRoleForOrganizations"),
			RoleId:                   sources.PtrString("AROA3VLV2U27YSTBFCGCJ"),
			Arn:                      sources.PtrString("arn:aws:iam::801795385023:role/service-role/AWSControlTowerConfigAggregatorRoleForOrganizations"),
			CreateDate:               sources.PtrTime(time.Now()),
			AssumeRolePolicyDocument: sources.PtrString("FOO"),
			MaxSessionDuration:       sources.PtrInt32(3600),
			Description:              sources.PtrString("description"),
			PermissionsBoundary: &types.AttachedPermissionsBoundary{
				PermissionsBoundaryArn:  sources.PtrString("arn:aws:iam::801795385023:role/service-role/AWSControlTowerConfigAggregatorRoleForOrganizations"),
				PermissionsBoundaryType: types.PermissionsBoundaryAttachmentTypePolicy,
			},
			RoleLastUsed: &types.RoleLastUsed{
				LastUsedDate: sources.PtrTime(time.Now()),
				Region:       sources.PtrString("us-east-2"),
			},
		},
		Policies: []string{
			"one",
		},
	}

	item, err := RoleItemMapper("foo", &role)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "iam-role-policy",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "one",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}
