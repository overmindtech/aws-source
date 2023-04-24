package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type RoleDetails struct {
	Role     *types.Role
	Policies []string
}

func roleGetFunc(ctx context.Context, client IAMClient, scope, query string) (*RoleDetails, error) {
	out, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: &query,
	})

	if err != nil {
		return nil, err
	}

	details := RoleDetails{
		Role: out.Role,
	}

	err = enrichRole(ctx, client, &details)

	if err != nil {
		return nil, err
	}

	return &details, nil
}

func enrichRole(ctx context.Context, client IAMClient, roleDetails *RoleDetails) error {
	var err error

	roleDetails.Policies, err = getRolePolicies(ctx, client, *roleDetails.Role.RoleName)

	if err != nil {
		return err
	}

	roleDetails.Role.Tags, err = getRoleTags(ctx, client, *roleDetails.Role.RoleName)

	return err
}

func getRolePolicies(ctx context.Context, client IAMClient, roleName string) ([]string, error) {
	policiesPaginator := iam.NewListRolePoliciesPaginator(client, &iam.ListRolePoliciesInput{
		RoleName: &roleName,
	})

	policies := make([]string, 0)

	for policiesPaginator.HasMorePages() {
		out, err := policiesPaginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		policies = append(policies, out.PolicyNames...)
	}

	return policies, nil
}

func getRoleTags(ctx context.Context, client IAMClient, roleName string) ([]types.Tag, error) {
	out, err := client.ListRoleTags(ctx, &iam.ListRoleTagsInput{
		RoleName: &roleName,
	})

	if err != nil {
		return nil, err
	}

	return out.Tags, nil
}

func roleListFunc(ctx context.Context, client IAMClient, scope string) ([]*RoleDetails, error) {
	paginator := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})
	roles := make([]*RoleDetails, 0)

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for i := range out.Roles {
			details := RoleDetails{
				Role: &out.Roles[i],
			}

			err = enrichRole(ctx, client, &details)

			if err != nil {
				return nil, err
			}

			roles = append(roles, &details)
		}
	}

	return roles, nil
}

func roleItemMapper(scope string, awsItem *RoleDetails) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem.Role)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-role",
		UniqueAttribute: "roleName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, policy := range awsItem.Policies {
		// +overmind:link iam-policy
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "iam-policy",
			Method: sdp.QueryMethod_GET,
			Query:  policy,
			Scope:  scope,
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type iam-role
// +overmind:descriptiveType IAM Role
// +overmind:get Get an IAM role by name
// +overmind:list List all IAM roles
// +overmind:search Search for IAM roles by ARN
// +overmind:group AWS

func NewRoleSource(config aws.Config, accountID string, region string) *sources.GetListSource[*RoleDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*RoleDetails, IAMClient, *iam.Options]{
		ItemType:   "iam-role",
		Client:     iam.NewFromConfig(config),
		AccountID:  accountID,
		GetFunc:    roleGetFunc,
		ListFunc:   roleListFunc,
		ItemMapper: roleItemMapper,
	}
}
