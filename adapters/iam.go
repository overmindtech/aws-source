package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/micahhausler/aws-iam-policy/policy"

	"github.com/overmindtech/aws-source/adapterhelpers"
	"github.com/overmindtech/sdp-go"
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

// Extracts linked item queries from an IAM policy. In this case we only link to
// entities that are explicitly mentioned in the policy. If we were to link to
// more you'd end up with way too many links since a policy might for example
// give read access to everything
func LinksFromPolicy(document *policy.Policy) []*sdp.LinkedItemQuery {
	// We want to link all of the resources in the policy document, as long
	// as they have a valid ARN
	var arn *adapterhelpers.ARN
	var err error
	queries := make([]*sdp.LinkedItemQuery, 0)

	if document == nil || document.Statements == nil {
		return queries
	}

	for _, statement := range document.Statements.Values() {
		if statement.Principal != nil {
			// If we are referencing a specific IAM user or role as the
			// principal then we should link them here
			if awsPrincipal := statement.Principal.AWS(); awsPrincipal != nil {
				for _, value := range awsPrincipal.Values() {
					// These are in the format of ARN so we'll parse them
					if arn, err := adapterhelpers.ParseARN(value); err == nil {
						var typ string
						switch arn.Type() {
						case "role":
							typ = "iam-role"
						case "user":
							typ = "iam-user"
						}

						if typ != "" {
							queries = append(queries, &sdp.LinkedItemQuery{
								Query: &sdp.Query{
									Type:   "iam-role",
									Method: sdp.QueryMethod_SEARCH,
									Query:  arn.String(),
									Scope:  adapterhelpers.FormatScope(arn.AccountID, arn.Region),
								},
								BlastPropagation: &sdp.BlastPropagation{
									// If a user or role iex explicitly
									// referenced, I think it's reasonable to
									// assume that they are tightly bound
									In:  true,
									Out: true,
								},
							})
						}
					}
				}
			}
		}

		if statement.Resource != nil {
			for _, resource := range statement.Resource.Values() {
				arn, err = adapterhelpers.ParseARN(resource)
				if err != nil {
					continue
				}

				// If the ARN contains a wildcard then we want to bail out
				possibleWildcards := arn.AccountID + arn.Type() + arn.ResourceID()
				if strings.Contains(possibleWildcards, "*") {
					continue
				}

				// Since this could be an ARN to anything we are going to rely
				// on the fact that we *usually* have a SEARCH method that
				// accepts ARNs
				scope := sdp.WILDCARD
				if arn.AccountID != "aws" {
					// If we have an account and region, then use those
					scope = adapterhelpers.FormatScope(arn.AccountID, arn.Region)
				}

				// It would be good here if we had a way to definitely know what
				// type a given ARN is, but I don't think the types are 1:1 so
				// we are going to have to use a wildcard. This will cause a lot
				// of failed searches which I don't love, but it will work
				itemType := sdp.WILDCARD

				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   itemType,
						Method: sdp.QueryMethod_SEARCH,
						Query:  arn.String(),
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  false,
						Out: true,
					},
				})
			}
		}
	}

	return queries
}

// Parses an IAM policy in it's URL-encoded embedded form
func ParsePolicyDocument(encoded string) (*policy.Policy, error) {
	// Decode the policy document which is RFC 3986 URL encoded
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode policy document: %w", err)
	}

	// Unmarshal the JSON
	policyDocument := policy.Policy{}
	err = json.Unmarshal([]byte(decoded), &policyDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy document: %w", err)
	}

	return &policyDocument, nil
}
