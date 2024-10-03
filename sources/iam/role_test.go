package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/micahhausler/aws-iam-policy/policy"
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
}

func TestRoleListFunc(t *testing.T) {
	roles, err := roleListFunc(context.Background(), &TestIAMClient{}, "foo")

	if err != nil {
		t.Error(err)
	}

	if len(roles) != 1 {
		t.Errorf("expected 1 role, got %b", len(roles))
	}
}

func TestRoleListTagsFunc(t *testing.T) {
	tags, err := roleListTagsFunc(context.Background(), &RoleDetails{
		Role: &types.Role{
			Arn: sources.PtrString("arn:aws:iam::801795385023:role/service-role/AWSControlTowerConfigAggregatorRoleForOrganizations"),
		},
	}, &TestIAMClient{})

	if err != nil {
		t.Error(err)
	}

	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %v", len(tags))
	}
}

const listBucketsPolicy = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "VisualEditor0",
			"Effect": "Allow",
			"Action": "s3:ListAllMyBuckets",
			"Resource": "*"
		}
	]
}`

func TestRoleItemMapper(t *testing.T) {
	policyDoc := policy.Policy{}
	err := json.Unmarshal([]byte(listBucketsPolicy), &policyDoc)
	if err != nil {
		t.Fatal(err)
	}

	role := RoleDetails{
		Role: &types.Role{
			Path:                     sources.PtrString("/service-role/"),
			RoleName:                 sources.PtrString("AWSControlTowerConfigAggregatorRoleForOrganizations"),
			RoleId:                   sources.PtrString("AROA3VLV2U27YSTBFCGCJ"),
			Arn:                      sources.PtrString("arn:aws:iam::801795385023:role/service-role/AWSControlTowerConfigAggregatorRoleForOrganizations"),
			CreateDate:               sources.PtrTime(time.Now()),
			AssumeRolePolicyDocument: sources.PtrString(`%7B%22Version%22%3A%222012-10-17%22%2C%22Statement%22%3A%5B%7B%22Effect%22%3A%22Allow%22%2C%22Principal%22%3A%7B%22Service%22%3A%22config.amazonaws.com%22%7D%2C%22Action%22%3A%22sts%3AAssumeRole%22%7D%5D%7D`),
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
				Name:     "foo",
				Document: &policyDoc,
			},
		},
		AttachedPolicies: []types.AttachedPolicy{
			{
				PolicyArn:  sources.PtrString("arn:aws:iam::aws:policy/AdministratorAccess"),
				PolicyName: sources.PtrString("AdministratorAccess"),
			},
		},
	}

	item, err := roleItemMapper("", "foo", &role)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "iam-policy",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam::aws:policy/AdministratorAccess",
			ExpectedScope:  "aws",
		},
	}

	tests.Execute(t, item)

	fmt.Println(item.ToMap())
}

func TestNewRoleSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)
	client := iam.NewFromConfig(config, func(o *iam.Options) {
		o.RetryMode = aws.RetryModeAdaptive
		o.RetryMaxAttempts = 10
	})

	source := NewRoleSource(client, account, region)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 30 * time.Hour,
	}

	test.Run(t)
}
