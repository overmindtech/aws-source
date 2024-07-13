package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/micahhausler/aws-iam-policy/policy"
	"go.opentelemetry.io/otel/attribute"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
	"github.com/sourcegraph/conc/iter"
)

type RoleDetails struct {
	Role             *types.Role
	EmbeddedPolicies []embeddedPolicy
	AttachedPolicies []types.AttachedPolicy
}

func roleGetFunc(ctx context.Context, client IAMClient, _, query string) (*RoleDetails, error) {
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

	// In this section we want to get the embedded polices, and determine links
	// to the attached policies

	// Get embedded policies
	roleDetails.EmbeddedPolicies, err = getEmbeddedPolicies(ctx, client, *roleDetails.Role.RoleName)

	if err != nil {
		return err
	}

	// Get the attached policies and create links to these
	roleDetails.AttachedPolicies, err = getAttachedPolicies(ctx, client, *roleDetails.Role.RoleName)

	if err != nil {
		return err
	}

	return nil
}

type embeddedPolicy struct {
	Name     string
	Document *policy.Policy
}

// getEmbeddedPolicies returns a list of inline policies embedded in the role
func getEmbeddedPolicies(ctx context.Context, client IAMClient, roleName string) ([]embeddedPolicy, error) {
	policiesPaginator := iam.NewListRolePoliciesPaginator(client, &iam.ListRolePoliciesInput{
		RoleName: &roleName,
	})
	ctx, span := tracer.Start(ctx, "getEmbeddedPolicies")
	defer span.End()

	policies := make([]embeddedPolicy, 0)

	for policiesPaginator.HasMorePages() {
		out, err := policiesPaginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, policyName := range out.PolicyNames {
			embeddedPolicy, err := getRolePolicyDetails(ctx, client, roleName, policyName)

			if err != nil {
				// Ignore these errors
				continue
			}

			policies = append(policies, *embeddedPolicy)
		}
	}

	return policies, nil
}

func getRolePolicyDetails(ctx context.Context, client IAMClient, roleName string, policyName string) (*embeddedPolicy, error) {
	ctx, span := tracer.Start(ctx, "getRolePolicyDetails")
	defer span.End()
	policy, err := client.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
		RoleName:   &roleName,
		PolicyName: &policyName,
	})

	if err != nil {
		return nil, err
	}

	if policy == nil || policy.PolicyDocument == nil {
		return nil, errors.New("policy document not found")
	}

	policyDoc, err := parsePolicyDocument(*policy.PolicyDocument)
	if err != nil {
		return nil, fmt.Errorf("error parsing policy document: %w", err)
	}

	return &embeddedPolicy{
		Name:     policyName,
		Document: policyDoc,
	}, nil
}

// getAttachedPolicies Gets the attached policies for a role, these are actual
// managed policies that can be linked to rather than embedded ones
func getAttachedPolicies(ctx context.Context, client IAMClient, roleName string) ([]types.AttachedPolicy, error) {
	paginator := iam.NewListAttachedRolePoliciesPaginator(client, &iam.ListAttachedRolePoliciesInput{
		RoleName: &roleName,
	})

	attachedPolicies := make([]types.AttachedPolicy, 0)

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		attachedPolicies = append(attachedPolicies, out.AttachedPolicies...)
	}

	return attachedPolicies, nil
}

func roleListFunc(ctx context.Context, client IAMClient, _ string) ([]*RoleDetails, error) {
	paginator := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})
	roles := make([]*RoleDetails, 0)
	ctx, span := tracer.Start(ctx, "roleListFunc")
	defer span.End()

	mapper := iter.Mapper[types.Role, *RoleDetails]{
		MaxGoroutines: 100,
	}

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		newRoles, err := mapper.MapErr(out.Roles, func(role *types.Role) (*RoleDetails, error) {
			details := RoleDetails{
				Role: role,
			}

			err := enrichRole(ctx, client, &details)

			if err != nil {
				return nil, err
			}

			return &details, nil
		})

		if err != nil {
			return nil, err
		}

		roles = append(roles, newRoles...)
	}

	span.SetAttributes(
		attribute.Int("ovm.aws.numRoles", len(roles)),
	)

	return roles, nil
}

func roleItemMapper(scope string, awsItem *RoleDetails) (*sdp.Item, error) {
	enrichedRole := struct {
		*types.Role
		EmbeddedPolicies []embeddedPolicy
		// This is a replacement for the URL-encoded policy document so that the
		// user can see the policy
		AssumeRolePolicyDocument *policy.Policy
	}{
		Role:             awsItem.Role,
		EmbeddedPolicies: awsItem.EmbeddedPolicies,
	}

	// Parse the encoded policy document
	if awsItem.Role.AssumeRolePolicyDocument != nil {
		policyDoc, err := parsePolicyDocument(*awsItem.Role.AssumeRolePolicyDocument)
		if err != nil {
			return nil, err
		}

		enrichedRole.AssumeRolePolicyDocument = policyDoc
	}

	attributes, err := sources.ToAttributesCase(enrichedRole)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-role",
		UniqueAttribute: "roleName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, policy := range awsItem.AttachedPolicies {
		if policy.PolicyArn != nil {
			if a, err := sources.ParseARN(*policy.PolicyArn); err == nil {
				// +overmind:link iam-policy
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "iam-policy",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *policy.PolicyArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the policy will affect the role
						In: true,
						// Changing the role won't affect the policy
						Out: false,
					},
				})
			}
		}
	}

	// Extract links from policy documents
	for _, policy := range awsItem.EmbeddedPolicies {
		item.LinkedItemQueries = append(item.LinkedItemQueries, LinksFromPolicy(policy.Document)...)
	}

	// Extract links from the assume role policy document
	if enrichedRole.AssumeRolePolicyDocument != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, LinksFromPolicy(enrichedRole.AssumeRolePolicyDocument)...)
	}

	return &item, nil
}

func roleListTagsFunc(ctx context.Context, r *RoleDetails, client IAMClient) (map[string]string, error) {
	tags := make(map[string]string)

	paginator := iam.NewListRoleTagsPaginator(client, &iam.ListRoleTagsInput{
		RoleName: r.Role.RoleName,
	})

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return sources.HandleTagsError(ctx, err), nil
		}

		for _, tag := range out.Tags {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}
	}

	return tags, nil
}

//go:generate docgen ../../docs-data
// +overmind:type iam-role
// +overmind:descriptiveType IAM Role
// +overmind:get Get an IAM role by name
// +overmind:list List all IAM roles
// +overmind:search Search for IAM roles by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_iam_role.arn
// +overmind:terraform:method SEARCH

func NewRoleSource(client *iam.Client, accountID string, region string) *sources.GetListSource[*RoleDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*RoleDetails, IAMClient, *iam.Options]{
		ItemType:      "iam-role",
		Client:        client,
		CacheDuration: 3 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		AccountID:     accountID,
		GetFunc: func(ctx context.Context, client IAMClient, scope, query string) (*RoleDetails, error) {
			return roleGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client IAMClient, scope string) ([]*RoleDetails, error) {
			return roleListFunc(ctx, client, scope)
		},
		ListTagsFunc: roleListTagsFunc,
		ItemMapper:   roleItemMapper,
	}
}
