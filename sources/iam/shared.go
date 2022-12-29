package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IAMClient interface {
	GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error)
	GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error)
	GetPolicy(ctx context.Context, params *iam.GetPolicyInput, optFns ...func(*iam.Options)) (*iam.GetPolicyOutput, error)

	iam.ListEntitiesForPolicyAPIClient
	iam.ListPoliciesAPIClient
	iam.ListUsersAPIClient
	iam.ListGroupsForUserAPIClient
	iam.ListRolePoliciesAPIClient
	iam.ListRolesAPIClient
}
