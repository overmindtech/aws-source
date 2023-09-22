package iam

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
	log "github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc/iter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type PolicyDetails struct {
	Policy       *types.Policy
	PolicyGroups []types.PolicyGroup
	PolicyRoles  []types.PolicyRole
	PolicyUsers  []types.PolicyUser
}

func policyGetFunc(ctx context.Context, client IAMClient, scope, query string, limit *sources.LimitBucket) (*PolicyDetails, error) {
	// Construct the ARN from the name etc.
	a := sources.ARN{
		ARN: arn.ARN{
			Partition: "aws",
			Service:   "iam",
			Region:    "", // IAM doesn't have a region
			AccountID: scope,
			Resource:  "policy/" + query, // The query should policyFullName which is (path + name)
		},
	}

	limit.Wait(ctx)

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
		err := enrichPolicy(ctx, client, &details, limit)

		if err != nil {
			return nil, err
		}
	}

	return &details, nil
}

func enrichPolicy(ctx context.Context, client IAMClient, details *PolicyDetails, limit *sources.LimitBucket) error {
	err := addTags(ctx, client, details, limit)

	if err != nil {
		return err
	}

	err = addPolicyEntities(ctx, client, details, limit)

	if err != nil {
		return err
	}

	return nil
}

func addTags(ctx context.Context, client IAMClient, details *PolicyDetails, limit *sources.LimitBucket) error {
	var span trace.Span
	if log.GetLevel() == log.TraceLevel {
		// Only create new spans on trace level logging
		ctx, span = tracer.Start(ctx, "addTags")
		defer span.End()
	}

	limit.Wait(ctx)

	out, err := client.ListPolicyTags(ctx, &iam.ListPolicyTagsInput{
		PolicyArn: details.Policy.Arn,
	})

	if err != nil {
		return err
	}

	details.Policy.Tags = out.Tags

	return nil
}

func addPolicyEntities(ctx context.Context, client IAMClient, details *PolicyDetails, limit *sources.LimitBucket) error {
	var span trace.Span
	if log.GetLevel() == log.TraceLevel {
		// Only create new spans on trace level logging
		ctx, span = tracer.Start(ctx, "addPolicyEntities")
		defer span.End()
	}

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
		limit.Wait(ctx)

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
func policyListFunc(ctx context.Context, client IAMClient, scope string, limit *sources.LimitBucket) ([]*PolicyDetails, error) {
	var span trace.Span
	if log.GetLevel() == log.TraceLevel {
		// Only create new spans on trace level logging
		ctx, span = tracer.Start(ctx, "policyListFunc")
		defer span.End()
	} else {
		span = trace.SpanFromContext(ctx)
	}

	policies := make([]types.Policy, 0)

	var iamScope types.PolicyScopeType

	if scope == "aws" {
		iamScope = types.PolicyScopeTypeAws
	} else {
		iamScope = types.PolicyScopeTypeLocal
	}

	paginator := iam.NewListPoliciesPaginator(client, &iam.ListPoliciesInput{
		OnlyAttached: true,
		Scope:        iamScope,
	})

	for paginator.HasMorePages() {
		limit.Wait(ctx)

		out, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		policies = append(policies, out.Policies...)
	}

	span.SetAttributes(
		attribute.Int("om.aws.numPolicies", len(policies)),
	)

	policyDetails, err := iter.MapErr[types.Policy, *PolicyDetails](policies, func(p *types.Policy) (*PolicyDetails, error) {
		details := PolicyDetails{
			Policy: p,
		}

		err := enrichPolicy(ctx, client, &details, limit)

		return &details, err
	})

	if err != nil {
		return nil, err
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

	// Combine the path and policy name to create a unique attribute
	policyFullName := *awsItem.Policy.Path + *awsItem.Policy.PolicyName

	// Trim the leading slash
	policyFullName, _ = strings.CutPrefix(policyFullName, "/")

	// Create a new attribute which is a combination of `path` and `policyName`,
	// this can then be constructed into an ARN when a user calls GET
	attributes.Set("policyFullName", policyFullName)

	item := sdp.Item{
		Type:            "iam-policy",
		UniqueAttribute: "policyFullName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, group := range awsItem.PolicyGroups {
		// +overmind:link iam-group
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "iam-group",
				Query:  *group.GroupName,
				Method: sdp.QueryMethod_GET,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the group won't affect the policy
				In: false,
				// Changing the policy will affect the group
				Out: true,
			},
		})
	}

	for _, user := range awsItem.PolicyUsers {
		// +overmind:link iam-user
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "iam-user",
				Method: sdp.QueryMethod_GET,
				Query:  *user.UserName,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the user won't affect the policy
				In: false,
				// Changing the policy will affect the user
				Out: true,
			},
		})
	}

	for _, role := range awsItem.PolicyRoles {
		// +overmind:link iam-role
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "iam-role",
				Method: sdp.QueryMethod_GET,
				Query:  *role.RoleName,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the role won't affect the policy
				In: false,
				// Changing the policy will affect the role
				Out: true,
			},
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
// +overmind:terraform:queryMap aws_iam_policy.arn
// +overmind:terraform:queryMap aws_iam_user_policy_attachment.policy_arn
// +overmind:terraform:queryMap aws_iam_role_policy_attachment.policy_arn
// +overmind:terraform:method SEARCH

// NewPolicySource Note that this policy source only support polices that are
// user-created due to the fact that the AWS-created ones are basically "global"
// in scope. In order to get this to work I'd have to change the way the source
// is implemented so that it was mart enough to handle different scopes. This
// has been added to the backlog:
// https://github.com/overmindtech/aws-source/issues/68
func NewPolicySource(config aws.Config, accountID string, _ string, limit *sources.LimitBucket) *sources.GetListSource[*PolicyDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*PolicyDetails, IAMClient, *iam.Options]{
		ItemType:      "iam-policy",
		Client:        iam.NewFromConfig(config),
		CacheDuration: 1 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		AccountID:     accountID,
		Region:        "", // IAM policies aren't tied to a region

		// Some IAM policies are global, this means that their ARN doesn't
		// contain an account name and instead just says "aws". Enabling this
		// setting means these also work
		SupportGlobalResources: true,
		GetFunc: func(ctx context.Context, client IAMClient, scope, query string) (*PolicyDetails, error) {
			return policyGetFunc(ctx, client, scope, query, limit)
		},
		ListFunc: func(ctx context.Context, client IAMClient, scope string) ([]*PolicyDetails, error) {
			return policyListFunc(ctx, client, scope, limit)
		},
		ItemMapper: policyItemMapper,
	}
}
