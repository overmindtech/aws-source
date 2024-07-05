package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IAMClient interface {
	GetPolicy(ctx context.Context, params *iam.GetPolicyInput, optFns ...func(*iam.Options)) (*iam.GetPolicyOutput, error)
	GetPolicyVersion(ctx context.Context, params *iam.GetPolicyVersionInput, optFns ...func(*iam.Options)) (*iam.GetPolicyVersionOutput, error)
	GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error)
	GetRolePolicy(ctx context.Context, params *iam.GetRolePolicyInput, optFns ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error)
	GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error)
	ListPolicyTags(ctx context.Context, params *iam.ListPolicyTagsInput, optFns ...func(*iam.Options)) (*iam.ListPolicyTagsOutput, error)
	ListRoleTags(ctx context.Context, params *iam.ListRoleTagsInput, optFns ...func(*iam.Options)) (*iam.ListRoleTagsOutput, error)

	iam.ListAttachedRolePoliciesAPIClient
	iam.ListEntitiesForPolicyAPIClient
	iam.ListGroupsForUserAPIClient
	iam.ListPoliciesAPIClient
	iam.ListRolePoliciesAPIClient
	iam.ListRolesAPIClient
	iam.ListUsersAPIClient
	iam.ListUserTagsAPIClient
}
