package iam

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type PolicyDetails struct {
	Policy       *types.Policy
	PolicyGroups []types.PolicyGroup
	PolicyRoles  []types.PolicyRole
	PolicyUsers  []types.PolicyUser
}

func PolicyGetFunc(ctx context.Context, client IAMClient, scope, query string) (*PolicyDetails, error) {
	out, err := client.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: &query,
	})

	if err != nil {
		return nil, err
	}

	details := PolicyDetails{
		Policy: out.Policy,
	}

	if out.Policy != nil {
		err := addPolicyEntities(ctx, client, &details)

		if err != nil {
			return nil, err
		}
	}

	return &details, nil
}

func addPolicyEntities(ctx context.Context, client IAMClient, details *PolicyDetails) error {
	if details == nil {
		return errors.New("details is nil")
	}

	if details.Policy == nil {
		return errors.New("policy is nil")
	}

	paginator := iam.NewListEntitiesForPolicyPaginator(client, &iam.ListEntitiesForPolicyInput{
		PolicyArn: details.Policy.Arn,
	})

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return err
		}

		details.PolicyGroups = append(details.PolicyGroups, out.PolicyGroups...)
		details.PolicyRoles = append(details.PolicyRoles, out.PolicyRoles...)
		details.PolicyUsers = append(details.PolicyUsers, out.PolicyUsers...)
	}

	return nil
}

// PolicyListFunc Lists all attached policies. There is no way to list
// unattached policies since I don't think it will be very valuable, there are
// hundreds by default and if you aren't using them they aren't very interesting
func PolicyListFunc(ctx context.Context, client IAMClient, scope string) ([]*PolicyDetails, error) {
	policies := make([]types.Policy, 0)

	paginator := iam.NewListPoliciesPaginator(client, &iam.ListPoliciesInput{
		OnlyAttached: true,
	})

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		policies = append(policies, out.Policies...)
	}

	policyDetails := make([]*PolicyDetails, len(policies))

	for i, Policy := range policies {
		details := PolicyDetails{
			Policy: &Policy,
		}

		err := addPolicyEntities(ctx, client, &details)

		if err != nil {
			return nil, err
		}

		policyDetails[i] = &details
	}

	return policyDetails, nil
}

func PolicyItemMapper(scope string, awsItem *PolicyDetails) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem.Policy)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-policy",
		UniqueAttribute: "policyName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, group := range awsItem.PolicyGroups {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "iam-group",
			Query:  *group.GroupName,
			Method: sdp.RequestMethod_GET,
			Scope:  scope,
		})
	}

	for _, user := range awsItem.PolicyUsers {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "iam-user",
			Method: sdp.RequestMethod_GET,
			Query:  *user.UserName,
			Scope:  scope,
		})
	}

	for _, role := range awsItem.PolicyRoles {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "iam-role",
			Method: sdp.RequestMethod_GET,
			Query:  *role.RoleName,
			Scope:  scope,
		})
	}

	return &item, nil
}

func NewPolicySource(config aws.Config, accountID string, region string) *sources.GetListSource[*PolicyDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*PolicyDetails, IAMClient, *iam.Options]{
		ItemType:   "iam-policy",
		Client:     iam.NewFromConfig(config),
		AccountID:  accountID,
		Region:     region,
		GetFunc:    PolicyGetFunc,
		ListFunc:   PolicyListFunc,
		ItemMapper: PolicyItemMapper,
	}
}
