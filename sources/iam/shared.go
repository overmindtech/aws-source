package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IAMClient interface {
	GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error)
	GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error)

	iam.ListUsersAPIClient
	iam.ListGroupsForUserAPIClient
	iam.ListRolePoliciesAPIClient
	iam.ListRolesAPIClient
}
