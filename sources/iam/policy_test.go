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

func (t *TestIAMClient) ListPolicyTags(ctx context.Context, params *iam.ListPolicyTagsInput, optFns ...func(*iam.Options)) (*iam.ListPolicyTagsOutput, error) {
	return &iam.ListPolicyTagsOutput{
		Tags: []types.Tag{
			{
				Key:   sources.PtrString("foo"),
				Value: sources.PtrString("foo"),
			},
		},
	}, nil
}

func TestPolicyGetFunc(t *testing.T) {
	policy, err := policyGetFunc(context.Background(), &TestIAMClient{}, "foo", "bar")

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
	policies, err := policyListFunc(context.Background(), &TestIAMClient{}, "foo")

	if err != nil {
		t.Error(err)
	}

	if len(policies) != 2 {
		t.Errorf("expected 2 policies, got %v", len(policies))
	}
}

func TestPolicyListTagsFunc(t *testing.T) {
	tags, err := policyListTagsFunc(context.Background(), &PolicyDetails{
		Policy: &types.Policy{
			Arn: sources.PtrString("arn:aws:iam::801795385023:policy/service-role/AWSControlTowerAdminPolicy"),
		},
	}, &TestIAMClient{})

	if err != nil {
		t.Error(err)
	}

	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %v", len(tags))
	}
}

func TestPolicyItemMapper(t *testing.T) {
	item, err := policyItemMapper("foo", &PolicyDetails{
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

	tests := sources.QueryTests{
		{
			ExpectedType:   "iam-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "groupName",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "iam-user",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "userName",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "iam-role",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "roleName",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

	if item.UniqueAttributeValue() != "service-role/AWSControlTowerAdminPolicy" {
		t.Errorf("unexpected unique attribute value, got %s", item.UniqueAttributeValue())
	}
}

func TestNewPolicySource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewPolicySource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 30 * time.Second,
	}

	test.Run(t)

	// Test "aws" scoped resources
	t.Run("aws scoped resources in a specific scope", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), t.Name())

		defer span.End()

		t.Parallel()
		// This item shouldn't be found since it lives globally
		_, err := source.Get(ctx, sources.FormatScope(account, ""), "ReadOnlyAccess", false)

		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("aws scoped resources in the aws scope", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), t.Name())
		defer span.End()

		t.Parallel()
		// This item shouldn't be found since it lives globally
		item, err := source.Get(ctx, "aws", "ReadOnlyAccess", false)

		if err != nil {
			t.Error(err)
		}

		if item.UniqueAttributeValue() != "ReadOnlyAccess" {
			t.Errorf("expected globally unique name to be ReadOnlyAccess, got %v", item.GloballyUniqueName())
		}
	})

	t.Run("listing resources in a specific scope", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), t.Name())
		defer span.End()

		items, err := source.List(ctx, sources.FormatScope(account, ""), false)

		if err != nil {
			t.Error(err)
		}

		for _, item := range items {
			arnString, err := item.Attributes.Get("arn")

			if err != nil {
				t.Errorf("expected item to have an arn attribute, got %v", err)
			}

			arn, err := sources.ParseARN(arnString.(string))

			if err != nil {
				t.Error(err)
			}

			if arn.AccountID != account {
				t.Errorf("expected item account to be %v, got %v", account, arn.AccountID)
			}
		}

		t.Run("searching via ARN for a resource in a specific scope", func(t *testing.T) {
			ctx, span := tracer.Start(context.Background(), t.Name())
			defer span.End()

			t.Parallel()

			arn, _ := items[0].Attributes.Get("arn")

			_, err := source.Search(ctx, sources.FormatScope(account, ""), arn.(string), false)

			if err != nil {
				t.Error(err)
			}
		})

		t.Run("searching via ARN for a resource in the aws scope", func(t *testing.T) {
			ctx, span := tracer.Start(context.Background(), t.Name())
			defer span.End()

			t.Parallel()

			arn, _ := items[0].Attributes.Get("arn")

			_, err := source.Search(ctx, "aws", arn.(string), false)

			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	})

	t.Run("listing resources in the AWS scope", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), t.Name())
		defer span.End()

		items, err := source.List(ctx, "aws", false)

		if err != nil {
			t.Error(err)
		}

		for _, item := range items {
			arnString, err := item.Attributes.Get("arn")

			if err != nil {
				t.Errorf("expected item to have an arn attribute, got %v", err)
			}

			arn, err := sources.ParseARN(arnString.(string))

			if err != nil {
				t.Error(err)
			}

			if arn.AccountID != "aws" {
				t.Errorf("expected item account to be aws, got %v", arn.AccountID)
			}
		}

		t.Run("searching via ARN for a resource in a specific scope", func(t *testing.T) {
			ctx, span := tracer.Start(context.Background(), t.Name())
			defer span.End()

			t.Parallel()

			arn, _ := items[0].Attributes.Get("arn")

			_, err := source.Search(ctx, sources.FormatScope(account, ""), arn.(string), false)

			if err == nil {
				t.Error("expected error, got nil")
			}
		})

		t.Run("searching via ARN for a resource in the aws scope", func(t *testing.T) {
			ctx, span := tracer.Start(context.Background(), t.Name())
			defer span.End()

			t.Parallel()

			arn, _ := items[0].Attributes.Get("arn")

			_, err := source.Search(ctx, "aws", arn.(string), false)

			if err != nil {
				t.Error(err)
			}
		})
	})

}
