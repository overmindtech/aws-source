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

func (t *TestIAMClient) ListRoleTags(ctx context.Context, params *iam.ListRoleTagsInput, optFns ...func(*iam.Options)) (*iam.ListRoleTagsOutput, error) {
	return &iam.ListRoleTagsOutput{
		Tags: []types.Tag{
			{
				Key:   sources.PtrString("foo"),
				Value: sources.PtrString("bar"),
			},
		},
	}, nil
}

func (t *TestIAMClient) GetRolePolicy(ctx context.Context, params *iam.GetRolePolicyInput, optFns ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error) {
	return &iam.GetRolePolicyOutput{
		PolicyName: params.PolicyName,
		PolicyDocument: sources.PtrString(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Sid": "VisualEditor0",
					"Effect": "Allow",
					"Action": "s3:ListAllMyBuckets",
					"Resource": "*"
				}
			]
		}`),
		RoleName: params.RoleName,
	}, nil
}

func (t *TestIAMClient) ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	return &iam.ListAttachedRolePoliciesOutput{
		AttachedPolicies: []types.AttachedPolicy{
			{
				PolicyArn:  sources.PtrString("arn:aws:iam::aws:policy/AdministratorAccess"),
				PolicyName: sources.PtrString("AdministratorAccess"),
			},
			{
				PolicyArn:  sources.PtrString("arn:aws:iam::aws:policy/AmazonS3FullAccess"),
				PolicyName: sources.PtrString("AmazonS3FullAccess"),
			},
		},
	}, nil
}

func TestRoleGetFunc(t *testing.T) {
	role, err := roleGetFunc(context.Background(), &TestIAMClient{}, "foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if role.Role == nil {
		t.Error("role is nil")
	}

	if len(role.EmbeddedPolicies) != 2 {
		t.Errorf("expected 2 embedded policies, got %v", len(role.EmbeddedPolicies))
	}

	if len(role.AttachedPolicies) != 2 {
		t.Errorf("expected 2 attached policies, got %v", len(role.AttachedPolicies))
	}

	if len(role.Role.Tags) == 0 {
		t.Error("got no role tags")
	}
}

func TestRoleListFunc(t *testing.T) {
	roles, err := roleListFunc(context.Background(), &TestIAMClient{}, "foo")

	if err != nil {
		t.Error(err)
	}

	if len(roles) != 1 {
		t.Errorf("expected 1 role, got %b", len(roles))
	}

	if len(roles[0].Role.Tags) == 0 {
		t.Error("got no role tags")
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
		EmbeddedPolicies: []embeddedPolicy{
			{
				Name: "foo",
				Document: map[string]interface{}{
					"Version": "2012-10-17",
					"Statement": []map[string]interface{}{
						{
							"Sid":      "VisualEditor0",
							"Effect":   "Allow",
							"Action":   "s3:ListAllMyBuckets",
							"Resource": "*",
						},
					},
				},
			},
		},
		AttachedPolicies: []types.AttachedPolicy{
			{
				PolicyArn:  sources.PtrString("arn:aws:iam::aws:policy/AdministratorAccess"),
				PolicyName: sources.PtrString("AdministratorAccess"),
			},
		},
	}

	item, err := roleItemMapper("foo", &role)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "iam-policy",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam::aws:policy/AdministratorAccess",
			ExpectedScope:  "aws",
		},
	}

	tests.Execute(t, item)
}

func TestNewRoleSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewRoleSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Hour,
	}

	test.Run(t)
}
