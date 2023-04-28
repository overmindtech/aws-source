package iam

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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

func policyGetFunc(ctx context.Context, client IAMClient, scope, query string) (*PolicyDetails, error) {
	// Construct the ARN from the name etc.
	a := sources.ARN{
		ARN: arn.ARN{
			Partition: "aws",
			Service:   "iam",
			Region:    "", // IAM doesn't have a region
			AccountID: scope,
			Resource:  "policy" + query, // The query should policyFullName which is (path + name)
		},
	}

	out, err := client.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: sources.PtrString(a.String()),
	})

	if err != nil {
		return nil, err
	}

	details := PolicyDetails{
		Policy: out.Policy,
	}

	if out.Policy != nil {
		err := enrichPolicy(ctx, client, &details)

		if err != nil {
			return nil, err
		}
	}

	return &details, nil
}

func enrichPolicy(ctx context.Context, client IAMClient, details *PolicyDetails) error {
	err := addTags(ctx, client, details)

	if err != nil {
		return err
	}

	err = addPolicyEntities(ctx, client, details)

	return err
}

func addTags(ctx context.Context, client IAMClient, details *PolicyDetails) error {
	out, err := client.ListPolicyTags(ctx, &iam.ListPolicyTagsInput{
		PolicyArn: details.Policy.Arn,
	})

	if err != nil {
		return err
	}

	details.Policy.Tags = out.Tags

	return nil
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
func policyListFunc(ctx context.Context, client IAMClient, scope string) ([]*PolicyDetails, error) {
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

	for i := range policies {
		details := PolicyDetails{
			Policy: &policies[i],
		}

		err := enrichPolicy(ctx, client, &details)

		if err != nil {
			return nil, err
		}

		policyDetails[i] = &details
	}

	return policyDetails, nil
}

func policyItemMapper(scope string, awsItem *PolicyDetails) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem.Policy)

	if err != nil {
		return nil, err
	}

	if awsItem.Policy.Path == nil || awsItem.Policy.PolicyName == nil {
		return nil, errors.New("policy Path and PolicyName must be populated")
	}

	// Create a new attribute which is a combination of `path` and `policyName`,
	// this can then be constructed into an ARN when a user calls GET
	attributes.Set("policyFullName", *awsItem.Policy.Path+*awsItem.Policy.PolicyName)

	// Some IAM policies are global

	item := sdp.Item{
		Type:            "iam-policy",
		UniqueAttribute: "policyFullName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, group := range awsItem.PolicyGroups {
		// +overmind:link iam-group
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "iam-group",
			Query:  *group.GroupName,
			Method: sdp.QueryMethod_GET,
			Scope:  scope,
		})
	}

	for _, user := range awsItem.PolicyUsers {
		// +overmind:link iam-user
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "iam-user",
			Method: sdp.QueryMethod_GET,
			Query:  *user.UserName,
			Scope:  scope,
		})
	}

	for _, role := range awsItem.PolicyRoles {
		// +overmind:link iam-role
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "iam-role",
			Method: sdp.QueryMethod_GET,
			Query:  *role.RoleName,
			Scope:  scope,
		})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type iam-policy
// +overmind:descriptiveType IAM Policy
// +overmind:get Get an IAM policy by policyFullName ({path} + {policyName})
// +overmind:list List all IAM policies
// +overmind:search Search for IAM policies by ARN
// +overmind:group AWS

// NewPolicySource Note that this policy source only support polices that are
// user-created due to the fact that the AWS-created ones are basically "global"
// in scope. In order to get this to work I'd have to change the way the source
// is implemented so that it was mart enough to handle different scopes. This
// has been added to the backlog:
// https://github.com/overmindtech/aws-source/issues/68
func NewPolicySource(config aws.Config, accountID string, _ string) *sources.GetListSource[*PolicyDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*PolicyDetails, IAMClient, *iam.Options]{
		ItemType:   "iam-policy",
		Client:     iam.NewFromConfig(config),
		AccountID:  accountID,
		Region:     "", // IAM policies aren't tied to a region
		GetFunc:    policyGetFunc,
		ListFunc:   policyListFunc,
		ItemMapper: policyItemMapper,
	}
}
