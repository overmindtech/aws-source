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

func RoleGetFunc(ctx context.Context, client IAMClient, scope, query string) (*RoleDetails, error) {
	out, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: &query,
	})

	if err != nil {
		return nil, err
	}

	details := RoleDetails{
		Role: out.Role,
	}

	details.Policies, err = getRolePolicies(ctx, client, *out.Role.RoleName)

	if err != nil {
		return nil, err
	}

	return &details, nil
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

func RoleListFunc(ctx context.Context, client IAMClient, scope string) ([]*RoleDetails, error) {
	paginator := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})
	roles := make([]*RoleDetails, 0)

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, role := range out.Roles {
			details := RoleDetails{
				Role: &role,
			}

			details.Policies, err = getRolePolicies(ctx, client, *role.RoleName)

			if err != nil {
				return nil, err
			}

			roles = append(roles, &details)
		}
	}

	return roles, nil
}

func RoleItemMapper(scope string, awsItem *RoleDetails) (*sdp.Item, error) {
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
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "iam-role-policy",
			Method: sdp.RequestMethod_GET,
			Query:  policy,
			Scope:  scope,
		})
	}

	return &item, nil
}

func NewRoleSource(config aws.Config, accountID string, region string) *sources.GetListSource[*RoleDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*RoleDetails, IAMClient, *iam.Options]{
		ItemType:   "iam-role",
		Client:     iam.NewFromConfig(config),
		AccountID:  accountID,
		Region:     region,
		GetFunc:    RoleGetFunc,
		ListFunc:   RoleListFunc,
		ItemMapper: RoleItemMapper,
	}
}
